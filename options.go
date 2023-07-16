package errors

import (
	"encoding/json"
	"fmt"
	"time"
)

type Options struct {
	skipCallers int
	fields      []Field
}

func (o *Options) AddField(field Field) {
	o.fields = append(o.fields, field)
}

// Option is used to set error fields for structured logging and to skip caller
// for a stack trace.
type Option func(*Options)

func SkipCaller() Option {
	return func(options *Options) {
		options.skipCallers++
	}
}

func SkipCallers(skip int) Option {
	return func(options *Options) {
		options.skipCallers += skip
	}
}

func Bool(key string, value bool) Option {
	return func(options *Options) {
		options.AddField(BoolField{Key: key, Value: value})
	}
}

func Int(key string, value int) Option {
	return func(options *Options) {
		options.AddField(IntField{Key: key, Value: value})
	}
}

func Uint(key string, value uint) Option {
	return func(options *Options) {
		options.AddField(UintField{Key: key, Value: value})
	}
}

func Float(key string, value float64) Option {
	return func(options *Options) {
		options.AddField(FloatField{Key: key, Value: value})
	}
}

func String(key string, value string) Option {
	return func(options *Options) {
		options.AddField(StringField{Key: key, Value: value})
	}
}

func Stringer(key string, value fmt.Stringer) Option {
	return String(key, value.String())
}

func Strings(key string, values []string) Option {
	return func(options *Options) {
		options.AddField(StringsField{Key: key, Values: values})
	}
}

func Value(key string, value interface{}) Option {
	return func(options *Options) {
		options.AddField(ValueField{Key: key, Value: value})
	}
}

func Time(key string, value time.Time) Option {
	return func(options *Options) {
		options.AddField(TimeField{Key: key, Value: value})
	}
}

func Duration(key string, value time.Duration) Option {
	return func(options *Options) {
		options.AddField(DurationField{Key: key, Value: value})
	}
}

func JSON(key string, value json.RawMessage) Option {
	return func(options *Options) {
		options.AddField(JSONField{Key: key, Value: value})
	}
}

func newOptions(options ...Option) *Options {
	opts := &Options{}
	for _, set := range options {
		set(opts)
	}
	return opts
}
