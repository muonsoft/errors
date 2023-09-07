package logrusadapter

import (
	"encoding/json"
	"time"

	"github.com/muonsoft/errors"
	"github.com/sirupsen/logrus"
)

type Option func(adapter *adapter)

func SetLevel(level logrus.Level) Option {
	return func(adapter *adapter) {
		adapter.level = level
	}
}

func Log(err error, logger logrus.FieldLogger, options ...Option) {
	a := &adapter{log: logger, level: logrus.ErrorLevel}
	for _, setOption := range options {
		setOption(a)
	}
	errors.Log(err, a)
}

type adapter struct {
	log   logrus.FieldLogger
	level logrus.Level
}

func (a *adapter) SetBool(key string, value bool)              { a.log = a.log.WithField(key, value) }
func (a *adapter) SetInt(key string, value int)                { a.log = a.log.WithField(key, value) }
func (a *adapter) SetUint(key string, value uint)              { a.log = a.log.WithField(key, value) }
func (a *adapter) SetFloat(key string, value float64)          { a.log = a.log.WithField(key, value) }
func (a *adapter) SetString(key string, value string)          { a.log = a.log.WithField(key, value) }
func (a *adapter) SetStrings(key string, values []string)      { a.log = a.log.WithField(key, values) }
func (a *adapter) SetValue(key string, value interface{})      { a.log = a.log.WithField(key, value) }
func (a *adapter) SetTime(key string, value time.Time)         { a.log = a.log.WithField(key, value) }
func (a *adapter) SetDuration(key string, value time.Duration) { a.log = a.log.WithField(key, value) }
func (a *adapter) SetJSON(key string, value json.RawMessage)   { a.log = a.log.WithField(key, value) }

func (a *adapter) SetStackTrace(trace errors.StackTrace) {
	type Frame struct {
		Function string `json:"function"`
		File     string `json:"file,omitempty"`
		Line     int    `json:"line,omitempty"`
	}

	frames := make([]Frame, len(trace))
	for i, frame := range trace {
		frames[i].File = frame.File()
		frames[i].Function = frame.Name()
		frames[i].Line = frame.Line()
	}

	a.log = a.log.WithField("stackTrace", frames)
}

type levelLogger interface {
	Log(level logrus.Level, args ...interface{})
}

func (a *adapter) Log(message string) {
	if levelLog, ok := a.log.(levelLogger); ok {
		levelLog.Log(a.level, message)

		return
	}

	switch a.level {
	case logrus.PanicLevel:
		a.log.Panic(message)
	case logrus.FatalLevel:
		a.log.Fatal(message)
	case logrus.ErrorLevel:
		a.log.Error(message)
	case logrus.WarnLevel:
		a.log.Warn(message)
	case logrus.InfoLevel:
		a.log.Info(message)
	case logrus.DebugLevel:
		a.log.Debug(message)
	default:
		a.log.Error(message)
	}
}
