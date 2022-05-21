package errors

import (
	"encoding/json"
	"time"
)

type Options struct {
	skipCallers int
	fields      []Field
}

// Option is used to set error fields for structured logging and to skip caller
// for a stack trace.
type Option func(*Options)

// FieldSetter used to set error fields into structured logger.
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

func SkipCaller() Option {
	return func(options *Options) {
		options.skipCallers++
	}
}

func Bool(key string, value bool) Option {
	return func(options *Options) {
		options.fields = append(options.fields, BoolField{Key: key, Value: value})
	}
}

func Int(key string, value int) Option {
	return func(options *Options) {
		options.fields = append(options.fields, IntField{Key: key, Value: value})
	}
}

func Uint(key string, value uint) Option {
	return func(options *Options) {
		options.fields = append(options.fields, UintField{Key: key, Value: value})
	}
}

func Float(key string, value float64) Option {
	return func(options *Options) {
		options.fields = append(options.fields, FloatField{Key: key, Value: value})
	}
}

func String(key string, value string) Option {
	return func(options *Options) {
		options.fields = append(options.fields, StringField{Key: key, Value: value})
	}
}

func Strings(key string, values []string) Option {
	return func(options *Options) {
		options.fields = append(options.fields, StringsField{Key: key, Values: values})
	}
}

func Value(key string, value interface{}) Option {
	return func(options *Options) {
		options.fields = append(options.fields, ValueField{Key: key, Value: value})
	}
}

func Time(key string, value time.Time) Option {
	return func(options *Options) {
		options.fields = append(options.fields, TimeField{Key: key, Value: value})
	}
}

func Duration(key string, value time.Duration) Option {
	return func(options *Options) {
		options.fields = append(options.fields, DurationField{Key: key, Value: value})
	}
}

func JSON(key string, value json.RawMessage) Option {
	return func(options *Options) {
		options.fields = append(options.fields, JSONField{Key: key, Value: value})
	}
}

type BoolField struct {
	Key   string
	Value bool
}

func (f BoolField) Set(setter FieldSetter) {
	setter.SetBool(f.Key, f.Value)
}

type IntField struct {
	Key   string
	Value int
}

func (f IntField) Set(setter FieldSetter) {
	setter.SetInt(f.Key, f.Value)
}

type UintField struct {
	Key   string
	Value uint
}

func (f UintField) Set(setter FieldSetter) {
	setter.SetUint(f.Key, f.Value)
}

type FloatField struct {
	Key   string
	Value float64
}

func (f FloatField) Set(setter FieldSetter) {
	setter.SetFloat(f.Key, f.Value)
}

type StringField struct {
	Key   string
	Value string
}

func (f StringField) Set(setter FieldSetter) {
	setter.SetString(f.Key, f.Value)
}

type StringsField struct {
	Key    string
	Values []string
}

func (f StringsField) Set(setter FieldSetter) {
	setter.SetStrings(f.Key, f.Values)
}

type ValueField struct {
	Key   string
	Value interface{}
}

func (f ValueField) Set(setter FieldSetter) {
	setter.SetValue(f.Key, f.Value)
}

type TimeField struct {
	Key   string
	Value time.Time
}

func (f TimeField) Set(setter FieldSetter) {
	setter.SetTime(f.Key, f.Value)
}

type DurationField struct {
	Key   string
	Value time.Duration
}

func (f DurationField) Set(setter FieldSetter) {
	setter.SetDuration(f.Key, f.Value)
}

type JSONField struct {
	Key   string
	Value json.RawMessage
}

func (f JSONField) Set(setter FieldSetter) {
	setter.SetJSON(f.Key, f.Value)
}

func newOptions(options ...Option) *Options {
	opts := &Options{}
	for _, set := range options {
		set(opts)
	}
	return opts
}
