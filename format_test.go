package errors_test

import (
	"testing"
	"time"

	"github.com/muonsoft/errors"
)

func TestFormat_Errorf(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		format string
		want   string
	}{
		{
			"%s",
			errors.Errorf("%s", "error"),
			"%s",
			"error",
		},
		{
			"%v",
			errors.Errorf("%s", "error"),
			"%v",
			"error",
		},
		{
			"%+v for one error",
			errors.Errorf("%s", "error"),
			"%+v",
			"error\n" +
				"github.com/muonsoft/errors_test.TestFormat_Errorf\n" +
				"\t.+/errors/format_test.go:31",
		},
		{
			"%+v for wrapped error",
			errors.Wrap(errors.Errorf("%s", "error")),
			"%+v",
			"error\n" +
				"github.com/muonsoft/errors_test.TestFormat_Errorf\n" +
				"\t.+/errors/format_test.go:39",
		},
		{
			"%+v for recursively wrapped error",
			errors.Errorf("wrapped: %w", errors.Errorf("wrapped: %w", errors.New("error"))),
			"%+v",
			"wrapped: wrapped: error\n" +
				"github.com/muonsoft/errors_test.TestFormat_Errorf\n" +
				"\t.+/errors/format_test.go:47",
		},
		{
			"%+v for error with fields",
			errors.Errorf("%s", "error", errors.String("key", "value")),
			"%+v",
			"error\n" +
				"key: value\n" +
				"github.com/muonsoft/errors_test.TestFormat_Errorf\n" +
				"\t.+/errors/format_test.go:55",
		},
		{
			"%+v for wrapped error with fields",
			errors.Errorf(
				"wrapped: %w",
				errors.Errorf("%s", "error", errors.String("key", "value")),
				errors.String("wrappedKey", "wrappedValue"),
			),
			"%+v",
			"error\n" +
				"wrappedKey: wrappedValue\n" +
				"key: value\n" +
				"github.com/muonsoft/errors_test.TestFormat_Errorf\n" +
				"\t.+/errors/format_test.go:66",
		},
		{
			"%+v for error with bool true field",
			errors.Errorf("%s", "error", errors.Bool("key", true)),
			"%+v",
			"error\nkey: true\n",
		},
		{
			"%+v for error with bool false field",
			errors.Errorf("%s", "error", errors.Bool("key", false)),
			"%+v",
			"error\nkey: false\n",
		},
		{
			"%+v for error with int field",
			errors.Errorf("%s", "error", errors.Int("key", 123)),
			"%+v",
			"error\nkey: 123\n",
		},
		{
			"%+v for error with uint field",
			errors.Errorf("%s", "error", errors.Uint("key", 123)),
			"%+v",
			"error\nkey: 123\n",
		},
		{
			"%+v for error with float field",
			errors.Errorf("%s", "error", errors.Float("key", 123.123)),
			"%+v",
			"error\nkey: 123.123\n",
		},
		{
			"%+v for error with string field",
			errors.Errorf("%s", "error", errors.String("key", "value")),
			"%+v",
			"error\nkey: value\n",
		},
		{
			"%+v for error with strings field",
			errors.Errorf("%s", "error", errors.Strings("key", []string{"foo", "bar", "baz"})),
			"%+v",
			"error\nkey: foo, bar, baz\n",
		},
		{
			"%+v for error with value field",
			errors.Errorf("%s", "error", errors.Value("key", []string{"foo", "bar", "baz"})),
			"%+v",
			"error\nkey: \\[foo bar baz\\]\n",
		},
		{
			"%+v for error with time field",
			errors.Errorf("%s", "error", errors.Time("key", time.Date(2022, time.June, 13, 12, 0, 0, 0, time.UTC))),
			"%+v",
			"error\nkey: 2022\\-06\\-13 12:00:00 \\+0000 UTC\n",
		},
		{
			"%+v for error with duration field",
			errors.Errorf("%s", "error", errors.Duration("key", time.Hour+3*time.Minute)),
			"%+v",
			"error\nkey: 1h3m0s\n",
		},
		{
			"%+v for error with JSON field",
			errors.Errorf("%s", "error", errors.JSON("key", []byte(`{"key":"value"}`))),
			"%+v",
			"error\nkey: {\\\"key\\\":\\\"value\\\"}\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertFormatRegexp(t, test.err, test.format, test.want)
		})
	}
}
