package errors_test

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/muonsoft/errors"
)

// TestSlogAttrDirectUsage tests that slog.Attr can be passed directly to Errorf and Wrap.
func TestSlogAttrDirectUsage(t *testing.T) {
	t.Run("Wrap with slog.Attr", func(t *testing.T) {
		err := errors.Wrap(
			errors.New("base error"),
			slog.String("key1", "value1"),
			slog.Int("key2", 42),
		)

		attrs := errors.Attrs(err)
		if len(attrs) != 2 {
			t.Fatalf("expected 2 attributes, got %d", len(attrs))
		}

		if attrs[0].Key != "key1" || attrs[0].Value.String() != "value1" {
			t.Errorf("unexpected first attribute: %v", attrs[0])
		}
		if attrs[1].Key != "key2" || attrs[1].Value.Any() != int64(42) {
			t.Errorf("unexpected second attribute: %v", attrs[1])
		}
	})

	t.Run("Errorf with slog.Attr", func(t *testing.T) {
		err := errors.Errorf(
			"formatted error: %d",
			100,
			slog.String("key1", "value1"),
			slog.Bool("key2", true),
		)

		attrs := errors.Attrs(err)
		if len(attrs) != 2 {
			t.Fatalf("expected 2 attributes, got %d", len(attrs))
		}

		if attrs[0].Key != "key1" || attrs[0].Value.String() != "value1" {
			t.Errorf("unexpected first attribute: %v", attrs[0])
		}
		if attrs[1].Key != "key2" || attrs[1].Value.Bool() != true {
			t.Errorf("unexpected second attribute: %v", attrs[1])
		}
	})

	t.Run("mixed Option and slog.Attr", func(t *testing.T) {
		err := errors.Wrap(
			errors.New("base error"),
			errors.String("opt1", "value1"),
			slog.String("attr1", "value2"),
			errors.SkipCaller(),
			slog.Int("attr2", 123),
			errors.Int("opt2", 456),
		)

		attrs := errors.Attrs(err)
		if len(attrs) != 4 {
			t.Fatalf("expected 4 attributes, got %d", len(attrs))
		}

		expectedKeys := map[string]interface{}{
			"opt1":  "value1",
			"attr1": "value2",
			"attr2": int64(123),
			"opt2":  int64(456),
		}

		for _, attr := range attrs {
			expected, ok := expectedKeys[attr.Key]
			if !ok {
				t.Errorf("unexpected attribute key: %s", attr.Key)
				continue
			}
			actual := attr.Value.Any()
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("for key %s: expected %v, got %v", attr.Key, expected, actual)
			}
		}
	})

	t.Run("slog.Group directly in Wrap", func(t *testing.T) {
		err := errors.Wrap(
			errors.New("base error"),
			slog.Group("request",
				slog.String("method", "GET"),
				slog.String("path", "/api"),
			),
		)

		attrs := errors.Attrs(err)
		if len(attrs) != 1 {
			t.Fatalf("expected 1 attribute (group), got %d", len(attrs))
		}

		if attrs[0].Key != "request" {
			t.Errorf("expected group key 'request', got '%s'", attrs[0].Key)
		}

		if attrs[0].Value.Kind() != slog.KindGroup {
			t.Errorf("expected group kind, got %v", attrs[0].Value.Kind())
		}
	})

	t.Run("Errorf with wrapped error and slog.Attr", func(t *testing.T) {
		baseErr := errors.Wrap(errors.New("base"), slog.String("base_key", "base_value"))
		err := errors.Errorf(
			"wrapped: %w",
			baseErr,
			slog.String("wrap_key", "wrap_value"),
		)

		attrs := errors.Attrs(err)
		if len(attrs) != 2 {
			t.Fatalf("expected 2 attributes from chain, got %d", len(attrs))
		}

		hasBase := false
		hasWrap := false
		for _, attr := range attrs {
			if attr.Key == "base_key" && attr.Value.String() == "base_value" {
				hasBase = true
			}
			if attr.Key == "wrap_key" && attr.Value.String() == "wrap_value" {
				hasWrap = true
			}
		}

		if !hasBase || !hasWrap {
			t.Errorf("expected both base_key and wrap_key attributes")
		}
	})

	t.Run("all slog types directly", func(t *testing.T) {
		err := errors.Wrap(
			errors.New("test"),
			slog.Bool("b", true),
			slog.Int("i", 42),
			slog.Int64("i64", 9223372036854775807),
			slog.Uint64("u64", 18446744073709551615),
			slog.Float64("f64", 3.14),
			slog.String("s", "text"),
			slog.Any("any", []int{1, 2, 3}),
		)

		attrs := errors.Attrs(err)
		if len(attrs) != 7 {
			t.Fatalf("expected 7 attributes, got %d", len(attrs))
		}

		// Verify all types are present
		keys := make(map[string]bool)
		for _, attr := range attrs {
			keys[attr.Key] = true
		}

		expectedKeys := []string{"b", "i", "i64", "u64", "f64", "s", "any"}
		for _, key := range expectedKeys {
			if !keys[key] {
				t.Errorf("expected attribute with key '%s'", key)
			}
		}
	})
}

// TestSlogAttrVsOption tests that slog.Attr and Option produce equivalent results.
func TestSlogAttrVsOption(t *testing.T) {
	tests := []struct {
		name       string
		withOption func() error
		withAttr   func() error
	}{
		{
			name: "String",
			withOption: func() error {
				return errors.Wrap(errors.New("test"), errors.String("key", "value"))
			},
			withAttr: func() error {
				return errors.Wrap(errors.New("test"), slog.String("key", "value"))
			},
		},
		{
			name: "Int",
			withOption: func() error {
				return errors.Wrap(errors.New("test"), errors.Int("key", 123))
			},
			withAttr: func() error {
				return errors.Wrap(errors.New("test"), slog.Int("key", 123))
			},
		},
		{
			name: "Bool",
			withOption: func() error {
				return errors.Wrap(errors.New("test"), errors.Bool("key", true))
			},
			withAttr: func() error {
				return errors.Wrap(errors.New("test"), slog.Bool("key", true))
			},
		},
		{
			name: "Float64",
			withOption: func() error {
				return errors.Wrap(errors.New("test"), errors.Float64("key", 3.14))
			},
			withAttr: func() error {
				return errors.Wrap(errors.New("test"), slog.Float64("key", 3.14))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errOption := test.withOption()
			errAttr := test.withAttr()

			attrsOption := errors.Attrs(errOption)
			attrsAttr := errors.Attrs(errAttr)

			if len(attrsOption) != 1 || len(attrsAttr) != 1 {
				t.Fatalf("expected both to have 1 attribute, got Option=%d, Attr=%d",
					len(attrsOption), len(attrsAttr))
			}

			opt := attrsOption[0]
			attr := attrsAttr[0]

			if opt.Key != attr.Key {
				t.Errorf("keys differ: Option=%s, Attr=%s", opt.Key, attr.Key)
			}

			if !reflect.DeepEqual(opt.Value.Any(), attr.Value.Any()) {
				t.Errorf("values differ: Option=%v, Attr=%v", opt.Value.Any(), attr.Value.Any())
			}
		})
	}
}
