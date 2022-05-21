package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

func New(message string) error {
	return errors.New(message)
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func Error(message string, options ...Option) error {
	opts := newOptions(options...)

	return &stacked{
		wrapped: &wrapped{wrapped: New(message), fields: opts.fields},
		stack:   callers(opts.skipCallers),
	}
}

func Errorf(message string, args ...interface{}) error {
	err := fmt.Errorf(message, args...)

	argError := getArgError(message, args)
	if isWrapper(argError) {
		return err
	}

	return &stacked{
		wrapped: &wrapped{wrapped: err},
		stack:   callers(0),
	}
}

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
		} else {
			if s == '%' {
				isFormat = true
			}
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
