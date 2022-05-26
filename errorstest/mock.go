package errorstest

import (
	"encoding/json"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/muonsoft/errors"
)

type StackTrace []Frame

type Frame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

type Logger struct {
	Fields     map[string]interface{}
	StackTrace errors.StackTrace
	Message    string
}

func NewLogger() *Logger {
	return &Logger{Fields: make(map[string]interface{})}
}

func (m *Logger) SetBool(key string, value bool)              { m.Fields[key] = value }
func (m *Logger) SetInt(key string, value int)                { m.Fields[key] = value }
func (m *Logger) SetUint(key string, value uint)              { m.Fields[key] = value }
func (m *Logger) SetFloat(key string, value float64)          { m.Fields[key] = value }
func (m *Logger) SetString(key string, value string)          { m.Fields[key] = value }
func (m *Logger) SetStrings(key string, values []string)      { m.Fields[key] = values }
func (m *Logger) SetValue(key string, value interface{})      { m.Fields[key] = value }
func (m *Logger) SetTime(key string, value time.Time)         { m.Fields[key] = value }
func (m *Logger) SetDuration(key string, value time.Duration) { m.Fields[key] = value }
func (m *Logger) SetJSON(key string, value json.RawMessage)   { m.Fields[key] = value }
func (m *Logger) SetStackTrace(trace errors.StackTrace)       { m.StackTrace = trace }
func (m *Logger) Log(message string)                          { m.Message = message }

func (m *Logger) AssertMessage(t *testing.T, expected string) {
	t.Helper()

	if m.Message != expected {
		t.Errorf(`want logger message "%s", got "%s"`, expected, m.Message)
	}
}

func (m *Logger) AssertField(t *testing.T, key string, expected interface{}) {
	t.Helper()

	value, exists := m.Fields[key]
	if !exists {
		t.Errorf(`want logger to have a field with key "%s"`, key)
		return
	}
	if !reflect.DeepEqual(value, expected) {
		t.Errorf(`want logger to have a field with key "%s" and value "%v", got value "%v"`, key, expected, value)
	}
}

func (m *Logger) AssertStackTrace(t *testing.T, want StackTrace) {
	t.Helper()

	got := m.StackTrace
	if len(want) > len(got) {
		t.Errorf("unexpected stack: want(%d) > got(%d):\n got: %q\nwant: %q", len(want), len(got), got, want)
		return
	}

	for i, w := range want {
		match, err := regexp.MatchString(w.Function, got[i].Name())
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("unexpected stack: function on line %d:\n got: %q\nwant: %q", i+1, got[i].Name(), w.Function)
		}

		match, err = regexp.MatchString(w.File, got[i].File())
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("unexpected stack: file on line %d:\n got: %q\nwant: %q", i+1, got[i].File(), w.File)
		}

		if w.Line != got[i].Line() {
			t.Errorf("unexpected stack: line number on line %d:\n got: %d\nwant: %d", i+1, got[i].Line(), w.Line)
		}
	}
}
