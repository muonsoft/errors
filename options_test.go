package errors_test

import (
	"reflect"
	"testing"

	"github.com/muonsoft/errors"
)

func TestNewAttributeOptions(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected interface{}
	}{
		{
			name:     "Int64",
			err:      errors.Wrap(errors.New("test"), errors.Int64("key", 9223372036854775807)),
			expected: int64(9223372036854775807),
		},
		{
			name:     "Uint64",
			err:      errors.Wrap(errors.New("test"), errors.Uint64("key", 18446744073709551615)),
			expected: uint64(18446744073709551615),
		},
		{
			name:     "Float64",
			err:      errors.Wrap(errors.New("test"), errors.Float64("key", 3.14159)),
			expected: 3.14159,
		},
		{
			name:     "Any with struct",
			err:      errors.Wrap(errors.New("test"), errors.Any("key", struct{ Name string }{"test"})),
			expected: struct{ Name string }{"test"},
		},
		{
			name:     "Any with map",
			err:      errors.Wrap(errors.New("test"), errors.Any("key", map[string]int{"count": 42})),
			expected: map[string]int{"count": 42},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			attrs := errors.Attrs(test.err)
			if len(attrs) == 0 {
				t.Fatalf("expected %#v to have attributes", test.err)
			}

			var found bool
			var value interface{}
			for _, attr := range attrs {
				if attr.Key == "key" {
					value = attr.Value.Any()
					found = true
					break
				}
			}

			if !found {
				t.Fatalf("expected %#v to have attribute with key 'key'", test.err)
			}

			if !reflect.DeepEqual(value, test.expected) {
				t.Errorf("want value %v (%T), got %v (%T)", test.expected, test.expected, value, value)
			}
		})
	}
}

func TestValueDeprecated(t *testing.T) {
	// Test that Value still works but behaves the same as Any
	errWithValue := errors.Wrap(errors.New("test"), errors.Value("key", "test-value"))
	errWithAny := errors.Wrap(errors.New("test"), errors.Any("key", "test-value"))

	attrsValue := errors.Attrs(errWithValue)
	attrsAny := errors.Attrs(errWithAny)

	if len(attrsValue) != 1 || len(attrsAny) != 1 {
		t.Fatalf("expected both errors to have exactly 1 attribute")
	}

	valueAttr := attrsValue[0]
	anyAttr := attrsAny[0]

	if valueAttr.Key != anyAttr.Key {
		t.Errorf("expected same key, got Value=%s, Any=%s", valueAttr.Key, anyAttr.Key)
	}

	if valueAttr.Value.Any() != anyAttr.Value.Any() {
		t.Errorf("expected same value, got Value=%v, Any=%v", valueAttr.Value.Any(), anyAttr.Value.Any())
	}
}
