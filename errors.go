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
//   - error attributes use slog.Attr for native integration with Go's structured logging;
//   - package errors can be easily marshaled into JSON with all fields in a chain.
package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
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
		switch x := err.(type) {
		case interface{ Unwrap() error }:
			err = x.Unwrap()
			if err == nil {
				var z T
				return z, false
			}
		case interface{ Unwrap() []error }:
			for _, err := range x.Unwrap() {
				if t, ok := As[T](err); ok {
					return t, ok
				}
			}
			var z T
			return z, false
		default:
			var z T
			return z, false
		}
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
//
// Errorf also records the stack trace at the point it was called. If the wrapped error
// contains a stack trace then a new one will not be added to a chain.
//
// You can pass options to set structured attributes or to skip a caller in a stack trace.
// Both Option functions and slog.Attr values are accepted.
// Options/attributes must be specified after formatting arguments:
//
//	errors.Errorf("failed: %w", err, errors.String("key", "value"))
//	errors.Errorf("failed: %w", err, slog.String("key", "value"))
//	errors.Errorf("failed: %w", err, errors.SkipCaller(), slog.Int("id", 123))
func Errorf(message string, argsAndOptions ...interface{}) error {
	args, options := splitArgsAndOptions(argsAndOptions)
	opts := newOptions(options...)
	err := fmt.Errorf(message, args...)

	argErrors := getArgErrors(message, args)
	if len(argErrors) == 1 && isWrapper(argErrors[0]) {
		return &wrapped{wrapped: err, attrs: opts.attrs}
	}

	return &stacked{
		wrapped: &wrapped{wrapped: err, attrs: opts.attrs},
		stack:   newStack(opts.skipCallers),
	}
}

// Wrap returns an error annotating err with a stack trace at the point Wrap is called.
// If the wrapped error contains a stack trace then a new one will not be added to a chain.
// If err is nil, Wrap returns nil.
//
// You can pass options to set structured attributes or to skip a caller in a stack trace.
// Both Option functions and slog.Attr values are accepted:
//
//	errors.Wrap(err, errors.String("key", "value"))        // Using Option
//	errors.Wrap(err, slog.String("key", "value"))          // Using slog.Attr directly
//	errors.Wrap(err, errors.SkipCaller(), slog.Int("id", 123))  // Mixed
func Wrap(err error, optsOrAttrs ...interface{}) error {
	if err == nil {
		return nil
	}

	options := convertToOptions(optsOrAttrs)

	if isWrapper(err) {
		if len(options) == 0 {
			return err
		}

		return &wrapped{wrapped: err, attrs: newOptions(options...).attrs}
	}

	opts := newOptions(options...)

	return &stacked{
		wrapped: &wrapped{wrapped: err, attrs: opts.attrs},
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
	attrs   []slog.Attr
}

func (e *wrapped) Attrs() []slog.Attr { return e.attrs }
func (e *wrapped) Error() string      { return e.wrapped.Error() }
func (e *wrapped) Unwrap() error      { return e.wrapped }

// LogValue implements slog.LogValuer, allowing the error to be logged
// directly with slog and have its attributes automatically extracted.
func (e *wrapped) LogValue() slog.Value {
	return slog.GroupValue(e.attrs...)
}

func (e *wrapped) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		io.WriteString(s, e.Error())
		if s.Flag('+') {
			var err error
			for err = e; err != nil; err = Unwrap(err) {
				if loggable, ok := err.(LoggableError); ok {
					writeAttrs(s, loggable.Attrs(), "")
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
	data := map[string]interface{}{"error": e.Error()}

	var err error
	for err = e; err != nil; err = Unwrap(err) {
		if loggable, ok := err.(LoggableError); ok {
			attrsToMap(data, loggable.Attrs())
		}
		if tracer, ok := err.(stackTracer); ok {
			data["stackTrace"] = tracer.StackTrace()
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
			io.WriteString(s, e.Error())
			writeAttrs(s, e.Attrs(), "")
			e.stack.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	}
}

func (e *stacked) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{"error": e.Error()}
	data["stackTrace"] = e.StackTrace()

	var err error
	for err = e; err != nil; err = Unwrap(err) {
		if loggable, ok := err.(LoggableError); ok {
			attrsToMap(data, loggable.Attrs())
		}
	}

	return json.Marshal(data)
}

func splitArgsAndOptions(argsAndOptions []interface{}) ([]interface{}, []Option) {
	argsCount := len(argsAndOptions)
	for i := argsCount - 1; i >= 0; i-- {
		if isOptionOrAttr(argsAndOptions[i]) {
			argsCount--
		} else {
			break
		}
	}

	args := argsAndOptions[:argsCount]
	optsOrAttrs := argsAndOptions[argsCount:]
	options := convertToOptions(optsOrAttrs)

	return args, options
}

// isOptionOrAttr checks if a value is either an Option or slog.Attr.
func isOptionOrAttr(v interface{}) bool {
	if _, ok := v.(Option); ok {
		return true
	}
	if _, ok := v.(slog.Attr); ok {
		return true
	}
	return false
}

// convertToOptions converts a slice of Option and/or slog.Attr to []Option.
func convertToOptions(items []interface{}) []Option {
	options := make([]Option, 0, len(items))
	for _, item := range items {
		switch v := item.(type) {
		case Option:
			options = append(options, v)
		case slog.Attr:
			options = append(options, Attr(v))
		}
	}
	return options
}

func getArgErrors(message string, args []interface{}) []error {
	indices := getErrorIndices(message)
	errs := make([]error, 0, len(indices))

	for _, i := range indices {
		if err, ok := args[i].(error); ok {
			errs = append(errs, err)
		}
	}

	return errs
}

func getErrorIndices(message string) []int {
	indices := make([]int, 0, 1)
	isFormat := false

	i := -1
	for _, s := range message {
		if isFormat {
			if s != '%' {
				i++
				if s == 'w' {
					indices = append(indices, i)
				}
			}
			isFormat = false
		} else if s == '%' {
			isFormat = true
		}
	}

	return indices
}

// attrsToMap converts slog.Attr slice to a map for JSON marshaling.
// Groups with keys create nested maps. Groups without keys merge into the parent map.
func attrsToMap(target map[string]interface{}, attrs []slog.Attr) {
	for _, attr := range attrs {
		if attr.Value.Kind() == slog.KindGroup {
			groupAttrs := attr.Value.Group()
			if attr.Key == "" {
				// Group without key - merge into parent
				attrsToMap(target, groupAttrs)
			} else {
				// Group with key - create nested map
				nested := make(map[string]interface{})
				attrsToMap(nested, groupAttrs)
				target[attr.Key] = nested
			}
		} else {
			target[attr.Key] = attr.Value.Any()
		}
	}
}

// writeAttrs writes slog.Attr values to an io.Writer for %+v formatting.
// Groups with keys use dot notation (e.g., "group.key: value").
// Groups without keys merge their attributes at the current prefix level.
func writeAttrs(w io.Writer, attrs []slog.Attr, prefix string) {
	for _, attr := range attrs {
		if attr.Value.Kind() == slog.KindGroup {
			groupAttrs := attr.Value.Group()
			if attr.Key == "" {
				// Group without key - use same prefix
				writeAttrs(w, groupAttrs, prefix)
			} else {
				// Group with key - add to prefix
				newPrefix := prefix + attr.Key + "."
				writeAttrs(w, groupAttrs, newPrefix)
			}
		} else {
			writeAttr(w, prefix+attr.Key, attr.Value)
		}
	}
}

// writeAttr writes a single attribute value to an io.Writer.
func writeAttr(w io.Writer, key string, value slog.Value) {
	io.WriteString(w, "\n"+key+": ")

	switch value.Kind() {
	case slog.KindBool:
		if value.Bool() {
			io.WriteString(w, "true")
		} else {
			io.WriteString(w, "false")
		}
	case slog.KindInt64:
		io.WriteString(w, strconv.FormatInt(value.Int64(), 10))
	case slog.KindUint64:
		io.WriteString(w, strconv.FormatUint(value.Uint64(), 10))
	case slog.KindFloat64:
		io.WriteString(w, fmt.Sprintf("%f", value.Float64()))
	case slog.KindString:
		io.WriteString(w, value.String())
	case slog.KindTime:
		io.WriteString(w, value.Time().String())
	case slog.KindDuration:
		io.WriteString(w, value.Duration().String())
	default:
		// For Any and other types, handle special cases
		v := value.Any()
		switch typed := v.(type) {
		case []string:
			// Format string slices with comma separation
			for i, s := range typed {
				if i > 0 {
					io.WriteString(w, ", ")
				}
				io.WriteString(w, s)
			}
		case json.RawMessage:
			// Format JSON as string
			w.Write(typed)
		default:
			// Default formatting
			io.WriteString(w, fmt.Sprintf("%v", v))
		}
	}
}
