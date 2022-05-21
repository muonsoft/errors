package errors_test

import (
	"testing"

	"github.com/muonsoft/errors"
)

func TestFormat_Error(t *testing.T) {
	tests := []struct {
		error
		format string
		want   string
	}{
		{
			errors.Error("error"),
			"%s",
			"error",
		},
		{
			errors.Error("error"),
			"%v",
			"error",
		},
		{
			errors.Error("error"),
			"%+v",
			"error\n" +
				"github.com/muonsoft/errors_test.TestFormat_Error\n" +
				"\t.+/errors/format_test.go:26",
		},
		{
			errors.Error("error"),
			"%q",
			`"error"`,
		},
	}
	for _, test := range tests {
		t.Run(test.format, func(t *testing.T) {
			assertFormatRegexp(t, test.error, test.format, test.want)
		})
	}
}

func TestFormat_Errorf(t *testing.T) {
	tests := []struct {
		error
		format string
		want   string
	}{
		{
			errors.Errorf("%s", "error"),
			"%s",
			"error",
		},
		{
			errors.Errorf("%s", "error"),
			"%v",
			"error",
		},
		{
			errors.Errorf("%s", "error"),
			"%+v",
			"error\n" +
				"github.com/muonsoft/errors_test.TestFormat_Errorf\n" +
				"\t.+/errors/format_test.go:62",
		},
	}
	for _, test := range tests {
		t.Run(test.format, func(t *testing.T) {
			assertFormatRegexp(t, test.error, test.format, test.want)
		})
	}
}
