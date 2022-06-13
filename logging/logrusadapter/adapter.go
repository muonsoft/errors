package logrusadapter

import (
	"encoding/json"
	"time"

	"github.com/muonsoft/errors"
	"github.com/sirupsen/logrus"
)

func Log(err error, logger logrus.FieldLogger) {
	errors.Log(err, &adapter{l: logger})
}

type adapter struct {
	l logrus.FieldLogger
}

func (a *adapter) Log(message string)                          { a.l.Error(message) }
func (a *adapter) SetBool(key string, value bool)              { a.l = a.l.WithField(key, value) }
func (a *adapter) SetInt(key string, value int)                { a.l = a.l.WithField(key, value) }
func (a *adapter) SetUint(key string, value uint)              { a.l = a.l.WithField(key, value) }
func (a *adapter) SetFloat(key string, value float64)          { a.l = a.l.WithField(key, value) }
func (a *adapter) SetString(key string, value string)          { a.l = a.l.WithField(key, value) }
func (a *adapter) SetStrings(key string, values []string)      { a.l = a.l.WithField(key, values) }
func (a *adapter) SetValue(key string, value interface{})      { a.l = a.l.WithField(key, value) }
func (a *adapter) SetTime(key string, value time.Time)         { a.l = a.l.WithField(key, value) }
func (a *adapter) SetDuration(key string, value time.Duration) { a.l = a.l.WithField(key, value) }
func (a *adapter) SetJSON(key string, value json.RawMessage)   { a.l = a.l.WithField(key, value) }
func (a *adapter) SetStackTrace(trace errors.StackTrace)       { a.l = a.l.WithField("stackTrace", trace) }
