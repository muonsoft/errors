package errors

import (
	"encoding/json"
	"time"
)

type Options struct {
	skipCallers int
	fields      []Field
}

type Option func(*Options)

type FieldSetter interface {
	SetBool(key string, value bool)
	SetInt(key string, value int)
	SetUint(key string, value uint)
	SetFloat(key string, value float64)
	SetString(key string, value string)
	SetStrings(key string, values []string)
	SetValue(key string, value interface{})
	SetTime(key string, value time.Time)
	SetDuration(key string, value time.Duration)
	SetJSON(key string, value json.RawMessage)
	SetStackTrace(trace StackTrace)
}

type Field interface {
	Set(setter FieldSetter)
}

type FieldWriter interface {
	WriteFieldsTo(setter FieldSetter)
}

type fieldFunc func(setter FieldSetter)

func (f fieldFunc) Set(setter FieldSetter) {
	f(setter)
}

func SkipCaller() Option {
	return func(options *Options) {
		options.skipCallers++
	}
}

func Bool(key string, value bool) Option {
	return func(options *Options) {
		options.fields = append(options.fields, fieldFunc(func(setter FieldSetter) {
			setter.SetBool(key, value)
		}))
	}
}

func Int(key string, value int) Option {
	return func(options *Options) {
		options.fields = append(options.fields, fieldFunc(func(setter FieldSetter) {
			setter.SetInt(key, value)
		}))
	}
}

func Uint(key string, value uint) Option {
	return func(options *Options) {
		options.fields = append(options.fields, fieldFunc(func(setter FieldSetter) {
			setter.SetUint(key, value)
		}))
	}
}

func Float(key string, value float64) Option {
	return func(options *Options) {
		options.fields = append(options.fields, fieldFunc(func(setter FieldSetter) {
			setter.SetFloat(key, value)
		}))
	}
}

func String(key string, value string) Option {
	return func(options *Options) {
		options.fields = append(options.fields, fieldFunc(func(setter FieldSetter) {
			setter.SetString(key, value)
		}))
	}
}

func Strings(key string, values []string) Option {
	return func(options *Options) {
		options.fields = append(options.fields, fieldFunc(func(setter FieldSetter) {
			setter.SetStrings(key, values)
		}))
	}
}

func Value(key string, value interface{}) Option {
	return func(options *Options) {
		options.fields = append(options.fields, fieldFunc(func(setter FieldSetter) {
			setter.SetValue(key, value)
		}))
	}
}

func Time(key string, value time.Time) Option {
	return func(options *Options) {
		options.fields = append(options.fields, fieldFunc(func(setter FieldSetter) {
			setter.SetTime(key, value)
		}))
	}
}

func Duration(key string, value time.Duration) Option {
	return func(options *Options) {
		options.fields = append(options.fields, fieldFunc(func(setter FieldSetter) {
			setter.SetDuration(key, value)
		}))
	}
}

func JSON(key string, value json.RawMessage) Option {
	return func(options *Options) {
		options.fields = append(options.fields, fieldFunc(func(setter FieldSetter) {
			setter.SetJSON(key, value)
		}))
	}
}

func newOptions(options ...Option) *Options {
	opts := &Options{}
	for _, set := range options {
		set(opts)
	}

	return opts
}
