package errors_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/muonsoft/errors"
)

var errTest = errors.New("test error")

type StackTracer interface {
	StackTrace() errors.StackTrace
}

func assertSingleStack(t *testing.T, err error) {
	t.Helper()

	count := 0
	for e := err; e != nil; e = errors.Unwrap(e) {
		if _, ok := e.(StackTracer); ok {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected %#v to have exactly one stack in chain, got %d", err, count)
	}
}

func assertFormatRegexp(t *testing.T, arg interface{}, format, want string) {
	t.Helper()

	got := fmt.Sprintf(format, arg)
	gotLines := strings.SplitN(got, "\n", -1)
	wantLines := strings.SplitN(want, "\n", -1)

	if len(wantLines) > len(gotLines) {
		t.Errorf("wantLines(%d) > gotLines(%d):\n got: %q\nwant: %q", len(wantLines), len(gotLines), got, want)
		return
	}

	for i, w := range wantLines {
		match, err := regexp.MatchString(w, gotLines[i])
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("line %d: fmt.Sprintf(%q, err):\n got: %q\nwant: %q", i+1, format, got, want)
		}
	}
}

func assertStringsRegexp(t *testing.T, got, want []string) {
	t.Helper()

	if len(want) > len(got) {
		t.Errorf("want(%d) > got(%d):\n got: %q\nwant: %q", len(want), len(got), got, want)
		return
	}

	for i, w := range want {
		match, err := regexp.MatchString(w, got[i])
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("line %d:\n got: %q\nwant: %q", i+1, got, want)
		}
	}
}

func assertStackRegexp(t *testing.T, got, want JSONStack) {
	t.Helper()

	if len(want) > len(got) {
		t.Errorf("want(%d) > got(%d):\n got: %q\nwant: %q", len(want), len(got), got, want)
		return
	}

	for i, w := range want {
		match, err := regexp.MatchString(w.Function, got[i].Function)
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("function on line %d:\n got: %q\nwant: %q", i+1, got[i].Function, w.Function)
		}

		match, err = regexp.MatchString(w.File, got[i].File)
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("file on line %d:\n got: %q\nwant: %q", i+1, got[i].File, w.File)
		}

		if w.Line != got[i].Line {
			t.Errorf("line number on line %d:\n got: %d\nwant: %d", i+1, got[i].Line, w.Line)
		}
	}
}

type LoggerMock struct {
	fields  map[string]interface{}
	trace   errors.StackTrace
	message string
}

func NewLoggerMock() *LoggerMock {
	return &LoggerMock{fields: make(map[string]interface{})}
}

func (m *LoggerMock) SetBool(key string, value bool)              { m.fields[key] = value }
func (m *LoggerMock) SetInt(key string, value int)                { m.fields[key] = value }
func (m *LoggerMock) SetUint(key string, value uint)              { m.fields[key] = value }
func (m *LoggerMock) SetFloat(key string, value float64)          { m.fields[key] = value }
func (m *LoggerMock) SetString(key string, value string)          { m.fields[key] = value }
func (m *LoggerMock) SetStrings(key string, values []string)      { m.fields[key] = values }
func (m *LoggerMock) SetValue(key string, value interface{})      { m.fields[key] = value }
func (m *LoggerMock) SetTime(key string, value time.Time)         { m.fields[key] = value }
func (m *LoggerMock) SetDuration(key string, value time.Duration) { m.fields[key] = value }
func (m *LoggerMock) SetJSON(key string, value json.RawMessage)   { m.fields[key] = value }
func (m *LoggerMock) SetStackTrace(trace errors.StackTrace)       { m.trace = trace }
func (m *LoggerMock) Log(message string)                          { m.message = message }

func (m *LoggerMock) AssertMessage(t *testing.T, expected string) {
	t.Helper()

	if m.message != expected {
		t.Errorf(`want logger message "%s", got "%s"`, expected, m.message)
	}
}

func (m *LoggerMock) AssertField(t *testing.T, key string, expected interface{}) {
	t.Helper()

	value, exists := m.fields[key]
	if !exists {
		t.Errorf(`want logger to have a field with key "%s"`, key)
		return
	}
	if !reflect.DeepEqual(value, expected) {
		t.Errorf(`want logger to have a field with key "%s" and value "%v", got value "%v"`, key, expected, value)
	}
}

func (m *LoggerMock) AssertStackTrace(t *testing.T, want JSONStack) {
	t.Helper()

	got := m.trace
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
