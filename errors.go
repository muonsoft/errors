// Copyright 2022 Igor Lazarev. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package errors for structured logging.
//
// This package is based on well known github.com/pkg/errors. Key differences and features:
//
//   - errors.New() is an alias to standard library and (it does not add a stack trace) and
//     should be used to create sentinel package-level errors;
//   - minimalistic API: few methods to wrap an error: errors.Errorf(), errors.Wrap();
//   - adds stack trace idempotently (only once in a chain);
//   - options to skip caller in a stack trace and to add error fields for structured logging;
//   - error fields are made for the statically typed logger interface;
//   - package errors can be easily marshaled into JSON with all fields in a chain.
package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"
)

// New returns an error that formats as the given text.
// Each call to New returns a distinct error value even if the text is identical.
//
// This function is an alias to standard errors.New. Use it only for sentinel package-level errors.
// It should not add a stack trace.
func New(message string) error {
	return errors.New(message)
}

// Is reports whether any error in err's chain matches target.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error is considered to match a target if it is equal to that target or if
// it implements a method Is(error) bool such that Is(target) returns true.
//
// This function is an alias to standard errors.Is.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches type T, and if one is found, returns
// its value and true. Otherwise, it returns zero value and false.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error matches target if the error's concrete value is of type T, or if the error
// has a method As(any) bool such that As(target) returns true. In the latter case,
// the As method is responsible for setting returned value.
//
// An error type might provide an As method so it can be treated as if it were a
// different error type.
func As[T any](err error) (T, bool) {
	for err != nil {
		if t, ok := err.(T); ok {
			return t, true
		}
		if x, ok := err.(interface{ As(any) bool }); ok {
			var t T
			if x.As(&t) {
				return t, true
			}
		}
		err = Unwrap(err)
	}

	var z T
	return z, false
}

// IsOfType finds the first error in err's chain that matches type T, and if one is found, returns
// true. Otherwise, it returns false.
//
// It works exactly as As function, but returns only boolean flag.
func IsOfType[T any](err error) bool {
	_, is := As[T](err)
	return is
}

// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error.
// Otherwise, Unwrap returns nil.
//
// This function is an alias to standard errors.Unwrap.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error. You can wrap an error using %w modifier as it
// does fmt.Errorf function.
// Errorf also records the stack trace at the point it was called. If the wrapped error
// contains a stack trace then a new one will not be added to a chain.
// Also, you can pass an options to set a structured fields or to skip a caller
// in a stack trace. Options must be specified after formatting arguments.
func Errorf(message string, argsAndOptions ...interface{}) error {
	args, options := splitArgsAndOptions(argsAndOptions)
	opts := newOptions(options...)
	err := fmt.Errorf(message, args...)

	argError := getArgError(message, args)
	if isWrapper(argError) {
		return &wrapped{wrapped: err, fields: opts.fields}
	}

	return &stacked{
		wrapped: &wrapped{wrapped: err, fields: opts.fields},
		stack:   newStack(opts.skipCallers),
	}
}

// Wrap returns an error annotating err with a stack trace at the point Wrap is called.
// If the wrapped error contains a stack trace then a new one will not be added to a chain.
// If err is nil, Wrap returns nil.
// Also, you can pass an options to set a structured fields or to skip a caller
// in a stack trace.
func Wrap(err error, options ...Option) error {
	if err == nil {
		return nil
	}
	opts := newOptions(options...)

	if isWrapper(err) {
		return &wrapped{wrapped: err, fields: opts.fields}
	}

	return &stacked{
		wrapped: &wrapped{wrapped: err, fields: opts.fields},
		stack:   newStack(opts.skipCallers),
	}
}

type wrapper interface {
	isWrapper()
}

func isWrapper(err error) bool {
	if err == nil {
		return false
	}

	_, ok := As[wrapper](err)

	return ok
}

type wrapped struct {
	wrapper
	wrapped error
	fields  []Field
}

func (e *wrapped) Fields() []Field { return e.fields }
func (e *wrapped) Error() string   { return e.wrapped.Error() }
func (e *wrapped) Unwrap() error   { return e.wrapped }

func (e *wrapped) LogFields(logger FieldLogger) {
	for _, field := range e.fields {
		field.Set(logger)
	}
}

func (e *wrapped) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		io.WriteString(s, e.Error())
		if s.Flag('+') {
			fieldsWriter := &stringWriter{writer: s}
			var err error
			for err = e; err != nil; err = Unwrap(err) {
				if loggable, ok := err.(LoggableError); ok {
					loggable.LogFields(fieldsWriter)
				}
				if tracer, ok := err.(stackTracer); ok {
					tracer.StackTrace().Format(s, verb)
				}
			}
		}
	case 's', 'q':
		io.WriteString(s, e.Error())
	}
}

func (e *wrapped) MarshalJSON() ([]byte, error) {
	data := mapWriter{"error": e.Error()}

	var err error
	for err = e; err != nil; err = Unwrap(err) {
		if loggable, ok := err.(LoggableError); ok {
			loggable.LogFields(data)
		}
		if tracer, ok := err.(stackTracer); ok {
			data.SetStackTrace(tracer.StackTrace())
		}
	}

	return json.Marshal(data)
}

type stacked struct {
	*wrapped
	*stack
}

func (e *stacked) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			io.WriteString(s, e.wrapped.Error())
			e.wrapped.LogFields(&stringWriter{writer: s})
			e.stack.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, e.wrapped.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.wrapped.Error())
	}
}

func (e *stacked) MarshalJSON() ([]byte, error) {
	data := mapWriter{"error": e.Error()}
	data.SetStackTrace(e.StackTrace())
	e.LogFields(data)

	return json.Marshal(data)
}

func splitArgsAndOptions(argsAndOptions []interface{}) ([]interface{}, []Option) {
	argsCount := len(argsAndOptions)
	for i := argsCount - 1; i >= 0; i-- {
		if _, ok := argsAndOptions[i].(Option); ok {
			argsCount--
		} else {
			break
		}
	}

	args := argsAndOptions[:argsCount]
	options := make([]Option, 0, len(argsAndOptions)-argsCount)
	for i := argsCount; i < len(argsAndOptions); i++ {
		options = append(options, argsAndOptions[i].(Option))
	}

	return args, options
}

func getArgError(message string, args []interface{}) error {
	index := getErrorIndex(message)

	if index >= 0 && index < len(args) {
		if err, ok := args[index].(error); ok {
			return err
		}
	}

	return nil
}

func getErrorIndex(message string) int {
	i := -1
	isFormat := false

	for _, s := range message {
		if isFormat {
			if s != '%' {
				i++
				if s == 'w' {
					return i
				}
			}
			isFormat = false
		} else if s == '%' {
			isFormat = true
		}
	}

	return -1
}

type mapWriter map[string]interface{}

func (m mapWriter) SetBool(key string, value bool)              { m[key] = value }
func (m mapWriter) SetInt(key string, value int)                { m[key] = value }
func (m mapWriter) SetUint(key string, value uint)              { m[key] = value }
func (m mapWriter) SetFloat(key string, value float64)          { m[key] = value }
func (m mapWriter) SetString(key string, value string)          { m[key] = value }
func (m mapWriter) SetStrings(key string, values []string)      { m[key] = values }
func (m mapWriter) SetValue(key string, value interface{})      { m[key] = value }
func (m mapWriter) SetTime(key string, value time.Time)         { m[key] = value }
func (m mapWriter) SetDuration(key string, value time.Duration) { m[key] = value }
func (m mapWriter) SetJSON(key string, value json.RawMessage)   { m[key] = value }
func (m mapWriter) SetStackTrace(trace StackTrace)              { m["stackTrace"] = trace }

type stringWriter struct {
	writer io.Writer
}

func (s *stringWriter) SetBool(key string, value bool) {
	if value {
		io.WriteString(s.writer, "\n"+key+": true")
	} else {
		io.WriteString(s.writer, "\n"+key+": false")
	}
}

func (s *stringWriter) SetInt(key string, value int) {
	io.WriteString(s.writer, "\n"+key+": "+strconv.Itoa(value))
}

func (s *stringWriter) SetUint(key string, value uint) {
	io.WriteString(s.writer, "\n"+key+": "+strconv.FormatUint(uint64(value), 10))
}

func (s *stringWriter) SetFloat(key string, value float64) {
	io.WriteString(s.writer, "\n"+key+": "+fmt.Sprintf("%f", value))
}

func (s *stringWriter) SetString(key string, value string) {
	io.WriteString(s.writer, "\n"+key+": "+value)
}

func (s *stringWriter) SetStrings(key string, values []string) {
	io.WriteString(s.writer, "\n"+key+": ")
	for i, value := range values {
		if i > 0 {
			io.WriteString(s.writer, ", ")
		}
		io.WriteString(s.writer, value)
	}
}

func (s *stringWriter) SetValue(key string, value interface{}) {
	io.WriteString(s.writer, "\n"+key+": "+fmt.Sprintf("%v", value))
}

func (s *stringWriter) SetTime(key string, value time.Time) {
	io.WriteString(s.writer, "\n"+key+": "+value.String())
}

func (s *stringWriter) SetDuration(key string, value time.Duration) {
	io.WriteString(s.writer, "\n"+key+": "+value.String())
}

func (s *stringWriter) SetJSON(key string, value json.RawMessage) {
	io.WriteString(s.writer, "\n"+key+": "+string(value))
}

func (s *stringWriter) SetStackTrace(trace StackTrace) {}
