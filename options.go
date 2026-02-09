package errors

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

// Options holds configuration for error wrapping, including stack trace skip count
// and structured attributes for logging.
type Options struct {
	skipCallers int
	attrs       []slog.Attr
}

func (o *Options) addAttr(attr slog.Attr) {
	o.attrs = append(o.attrs, attr)
}

func (o *Options) addAttrs(attrs []slog.Attr) {
	o.attrs = append(o.attrs, attrs...)
}

// Option is used to set error attributes for structured logging and to skip caller
// for a stack trace.
type Option func(*Options)

// SkipCaller returns an Option that increments the skip count for stack trace by 1.
// Use this to exclude the immediate caller from the stack trace.
func SkipCaller() Option {
	return func(options *Options) {
		options.skipCallers++
	}
}

// SkipCallers returns an Option that adds skip to the skip count for stack trace.
// Use this to exclude multiple callers from the stack trace.
func SkipCallers(skip int) Option {
	return func(options *Options) {
		options.skipCallers += skip
	}
}

// Bool returns an Option that adds a boolean attribute to the error.
func Bool(key string, value bool) Option {
	return func(options *Options) {
		options.addAttr(slog.Bool(key, value))
	}
}

// Int returns an Option that adds an integer attribute to the error.
func Int(key string, value int) Option {
	return func(options *Options) {
		options.addAttr(slog.Int(key, value))
	}
}

// Uint returns an Option that adds an unsigned integer attribute to the error.
// Note: slog doesn't have a native Uint type, so this uses Any().
func Uint(key string, value uint) Option {
	return func(options *Options) {
		options.addAttr(slog.Any(key, value))
	}
}

// Float returns an Option that adds a float64 attribute to the error.
func Float(key string, value float64) Option {
	return func(options *Options) {
		options.addAttr(slog.Float64(key, value))
	}
}

// String returns an Option that adds a string attribute to the error.
func String(key string, value string) Option {
	return func(options *Options) {
		options.addAttr(slog.String(key, value))
	}
}

// Stringer returns an Option that adds a fmt.Stringer attribute to the error.
// The value is converted to string using its String() method.
func Stringer(key string, value fmt.Stringer) Option {
	return String(key, value.String())
}

// Strings returns an Option that adds a string slice attribute to the error.
// Note: slog doesn't have a native Strings type, so this uses Any().
func Strings(key string, values []string) Option {
	return func(options *Options) {
		options.addAttr(slog.Any(key, values))
	}
}

// Value returns an Option that adds an arbitrary value attribute to the error.
func Value(key string, value interface{}) Option {
	return func(options *Options) {
		options.addAttr(slog.Any(key, value))
	}
}

// Time returns an Option that adds a time.Time attribute to the error.
func Time(key string, value time.Time) Option {
	return func(options *Options) {
		options.addAttr(slog.Time(key, value))
	}
}

// Duration returns an Option that adds a time.Duration attribute to the error.
func Duration(key string, value time.Duration) Option {
	return func(options *Options) {
		options.addAttr(slog.Duration(key, value))
	}
}

// JSON returns an Option that adds a JSON attribute to the error.
func JSON(key string, value json.RawMessage) Option {
	return func(options *Options) {
		options.addAttr(slog.Any(key, value))
	}
}

// Attr returns an Option that adds an slog.Attr directly to the error.
// This allows using any slog attribute, including custom types and groups.
//
// Example:
//
//	err := errors.Wrap(err, errors.Attr(slog.Int64("timestamp", time.Now().Unix())))
func Attr(attr slog.Attr) Option {
	return func(options *Options) {
		options.addAttr(attr)
	}
}

// WithAttrs returns an Option that adds multiple slog.Attr values to the error.
//
// Example:
//
//	err := errors.Wrap(err, errors.WithAttrs(
//	    slog.String("user", "john"),
//	    slog.Int("age", 30),
//	))
func WithAttrs(attrs ...slog.Attr) Option {
	return func(options *Options) {
		options.addAttrs(attrs)
	}
}

// Group returns an Option that adds a group attribute to the error.
// Groups allow organizing related attributes under a common key.
//
// Example:
//
//	err := errors.Wrap(err, errors.Group("request",
//	    slog.String("method", "GET"),
//	    slog.String("path", "/api/users"),
//	    slog.Int("status", 404),
//	))
func Group(key string, attrs ...slog.Attr) Option {
	return func(options *Options) {
		options.addAttr(slog.Group(key, attrsToAny(attrs)...))
	}
}

// attrsToAny converts []slog.Attr to []any for use with slog.Group
func attrsToAny(attrs []slog.Attr) []any {
	result := make([]any, len(attrs))
	for i, attr := range attrs {
		result[i] = attr
	}
	return result
}

func newOptions(options ...Option) *Options {
	opts := &Options{}
	for _, set := range options {
		set(opts)
	}
	return opts
}
