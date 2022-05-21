package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

// As finds the first error in err's chain that matches target, and if one is found, sets
// target to that error value and returns true. Otherwise, it returns false.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error matches target if the error's concrete value is assignable to the value
// pointed to by target, or if the error has a method As(interface{}) bool such that
// As(target) returns true. In the latter case, the As method is responsible for
// setting target.
//
// An error type might provide an As method so it can be treated as if it were a
// different error type.
//
// As panics if target is not a non-nil pointer to either a type that implements
// error, or to any interface type.
//
// This function is an alias to standard errors.As.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error.
// Otherwise, Unwrap returns nil.
//
// This function is an alias to standard errors.Unwrap.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Error returns an error with the supplied message.
// Error also records the stack trace at the point it was called.
// Also, you can pass an options to set a structured fields or to skip a caller
// in a stack trace.
//
// This is equivalent to New function from github.com/pkg/errors.
func Error(message string, options ...Option) error {
	opts := newOptions(options...)

	return &stacked{
		wrapped: &wrapped{wrapped: New(message), fields: opts.fields},
		stack:   callers(opts.skipCallers),
	}
}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error. You can wrap an error using %w modifier as it
// does fmt.Errorf function.
// Errorf also records the stack trace at the point it was called. If the wrapped error
// contains a stack trace then a new one will not be added to a chain.
// Also, you can pass an options to set a structured fields or to skip a caller
// in a stack trace. Options must be set at the end of the parameters.
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
		stack:   callers(opts.skipCallers),
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
		stack:   callers(opts.skipCallers),
	}
}

type wrapper interface {
	isWrapper()
}

func isWrapper(err error) bool {
	if err == nil {
		return false
	}

	var w wrapper

	return errors.As(err, &w)
}

type wrapped struct {
	wrapper
	wrapped error
	fields  []Field
}

func (e *wrapped) Fields() []Field { return e.fields }
func (e *wrapped) Error() string   { return e.wrapped.Error() }
func (e *wrapped) Unwrap() error   { return e.wrapped }

func (e *wrapped) WriteFieldsTo(setter FieldSetter) {
	for _, field := range e.fields {
		field.Set(setter)
	}
}

func (e *wrapped) MarshalJSON() ([]byte, error) {
	data := mapWriter{"error": e.Error()}

	var err error
	for err = e; err != nil; err = Unwrap(err) {
		if w, ok := err.(FieldWriter); ok {
			w.WriteFieldsTo(data)
		}
		if s, ok := err.(stackTracer); ok {
			data.SetStackTrace(s.StackTrace())
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
	e.WriteFieldsTo(data)

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
