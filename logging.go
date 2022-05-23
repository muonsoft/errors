package errors

import (
	"encoding/json"
	"errors"
	"time"
)

// FieldLogger used to set error fields into structured logger.
type FieldLogger interface {
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

type Logger interface {
	FieldLogger
	Log(message string)
}

type Field interface {
	Set(logger FieldLogger)
}

type LoggableError interface {
	LogFields(logger FieldLogger)
}

func Log(err error, logger Logger) {
	if err == nil {
		return
	}

	for e := err; e != nil; e = errors.Unwrap(e) {
		if s, ok := e.(stackTracer); ok {
			logger.SetStackTrace(s.StackTrace())
		}
		if w, ok := e.(LoggableError); ok {
			w.LogFields(logger)
		}
	}

	logger.Log(err.Error())
}

type BoolField struct {
	Key   string
	Value bool
}

func (f BoolField) Set(logger FieldLogger) {
	logger.SetBool(f.Key, f.Value)
}

type IntField struct {
	Key   string
	Value int
}

func (f IntField) Set(logger FieldLogger) {
	logger.SetInt(f.Key, f.Value)
}

type UintField struct {
	Key   string
	Value uint
}

func (f UintField) Set(logger FieldLogger) {
	logger.SetUint(f.Key, f.Value)
}

type FloatField struct {
	Key   string
	Value float64
}

func (f FloatField) Set(logger FieldLogger) {
	logger.SetFloat(f.Key, f.Value)
}

type StringField struct {
	Key   string
	Value string
}

func (f StringField) Set(logger FieldLogger) {
	logger.SetString(f.Key, f.Value)
}

type StringsField struct {
	Key    string
	Values []string
}

func (f StringsField) Set(logger FieldLogger) {
	logger.SetStrings(f.Key, f.Values)
}

type ValueField struct {
	Key   string
	Value interface{}
}

func (f ValueField) Set(logger FieldLogger) {
	logger.SetValue(f.Key, f.Value)
}

type TimeField struct {
	Key   string
	Value time.Time
}

func (f TimeField) Set(logger FieldLogger) {
	logger.SetTime(f.Key, f.Value)
}

type DurationField struct {
	Key   string
	Value time.Duration
}

func (f DurationField) Set(logger FieldLogger) {
	logger.SetDuration(f.Key, f.Value)
}

type JSONField struct {
	Key   string
	Value json.RawMessage
}

func (f JSONField) Set(logger FieldLogger) {
	logger.SetJSON(f.Key, f.Value)
}
