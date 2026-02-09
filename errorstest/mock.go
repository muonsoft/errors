package errorstest

import (
	"log/slog"
	"reflect"
	"regexp"
	"testing"

	"github.com/muonsoft/errors"
)

type StackTrace []Frame

type Frame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

type Logger struct {
	Attrs      []slog.Attr
	StackTrace errors.StackTrace
	Message    string
	Level      slog.Level
}

func NewLogger() *Logger {
	return &Logger{Attrs: make([]slog.Attr, 0)}
}

func (m *Logger) AssertMessage(t *testing.T, expected string) {
	t.Helper()

	if m.Message != expected {
		t.Errorf(`want logger message "%s", got "%s"`, expected, m.Message)
	}
}

// AssertField checks if an attribute with the given key exists and has the expected value.
// It searches through all attributes, including nested groups (flattened with dot notation).
func (m *Logger) AssertField(t *testing.T, key string, expected interface{}) {
	t.Helper()

	value, exists := m.findAttr(key, m.Attrs, "")
	if !exists {
		t.Errorf(`want logger to have a field with key "%s"`, key)
		return
	}
	if !reflect.DeepEqual(value, expected) {
		t.Errorf(`want logger to have a field with key "%s" and value "%v", got value "%v"`, key, expected, value)
	}
}

// findAttr recursively searches for an attribute by key, handling groups with dot notation
func (m *Logger) findAttr(key string, attrs []slog.Attr, prefix string) (interface{}, bool) {
	for _, attr := range attrs {
		if attr.Value.Kind() == slog.KindGroup {
			groupAttrs := attr.Value.Group()
			if attr.Key == "" {
				// Group without key - search within same prefix
				if value, ok := m.findAttr(key, groupAttrs, prefix); ok {
					return value, true
				}
			} else {
				// Group with key - add to prefix
				newPrefix := prefix + attr.Key + "."
				if value, ok := m.findAttr(key, groupAttrs, newPrefix); ok {
					return value, true
				}
			}
		} else {
			fullKey := prefix + attr.Key
			if fullKey == key {
				return attr.Value.Any(), true
			}
		}
	}
	return nil, false
}

// AssertAttr checks if an attribute with matching key and value exists in the logger.
func (m *Logger) AssertAttr(t *testing.T, expected slog.Attr) {
	t.Helper()

	for _, attr := range m.Attrs {
		if attrsEqual(attr, expected) {
			return
		}
	}

	t.Errorf(`want logger to have attribute %v`, expected)
}

// AssertGroup checks if a group attribute with the given key exists and contains the expected attributes.
func (m *Logger) AssertGroup(t *testing.T, key string, expectedAttrs ...slog.Attr) {
	t.Helper()

	var groupAttrs []slog.Attr
	found := false

	for _, attr := range m.Attrs {
		if attr.Key == key && attr.Value.Kind() == slog.KindGroup {
			groupAttrs = attr.Value.Group()
			found = true
			break
		}
	}

	if !found {
		t.Errorf(`want logger to have a group with key "%s"`, key)
		return
	}

	for _, expected := range expectedAttrs {
		found := false
		for _, actual := range groupAttrs {
			if attrsEqual(actual, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf(`want group "%s" to contain attribute %v`, key, expected)
		}
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

// attrsEqual compares two slog.Attr values for equality
func attrsEqual(a, b slog.Attr) bool {
	if a.Key != b.Key {
		return false
	}
	if a.Value.Kind() != b.Value.Kind() {
		return false
	}

	// For groups, recursively compare
	if a.Value.Kind() == slog.KindGroup {
		aGroup := a.Value.Group()
		bGroup := b.Value.Group()
		if len(aGroup) != len(bGroup) {
			return false
		}
		for i := range aGroup {
			if !attrsEqual(aGroup[i], bGroup[i]) {
				return false
			}
		}
		return true
	}

	// For other types, compare values
	return reflect.DeepEqual(a.Value.Any(), b.Value.Any())
}
