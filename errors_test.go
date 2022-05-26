package errors_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/muonsoft/errors"
	"github.com/muonsoft/errors/errorstest"
)

func TestStackTrace(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want []string
	}{
		{
			name: "Errorf()",
			err:  errors.Errorf("ooh"),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:21",
			},
		},
		{
			name: "Wrap(Error())",
			err:  errors.Wrap(errors.Errorf("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:29",
			},
		},
		{
			name: "Wrap(New())",
			err:  errors.Wrap(errors.New("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:37",
			},
		},
		{
			name: "Wrap(Wrap(New()))",
			err:  errors.Wrap(errors.Wrap(errors.New("ooh"))),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:45",
			},
		},
		{
			name: "Errorf()",
			err:  errors.Errorf("ooh"),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:53",
			},
		},
		{
			name: `Errorf("%w", New())`,
			err:  errors.Errorf("%v", errors.New("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:61",
			},
		},
		{
			name: `Errorf("%w", Wrap(New()))`,
			err:  errors.Errorf("%w", errors.Wrap(errors.New("ooh"))),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:69",
			},
		},
		{
			name: `Errorf("%%w %v", Wrap(New()))`,
			err:  errors.Errorf("%%w %v", errors.New("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:77",
			},
		},
		{
			name: `Errorf("%s: %w", Wrap(New()))`,
			err:  errors.Errorf("%s: %w", "prefix", errors.New("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:85",
			},
		},
		{
			name: `Errorf("%w", Errorf("%w", New()))`,
			err:  errors.Errorf("%w", errors.Errorf("%w", errors.New("ooh"))),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:93",
			},
		},
		{
			name: `Errorf("%w", fmt.Errorf("%w", Error()))`,
			err:  errors.Errorf("%w", fmt.Errorf("%w", errors.New("ooh"))),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:101",
			},
		},
		{
			name: `wrap with New()`,
			err:  wrap(errors.New("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.wrap\n" +
					"\t.+/errors/errors_test.go:150",
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:109",
			},
		},
		{
			name: `wrap skip caller`,
			err:  wrapSkipCaller(errors.New("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:119",
			},
		},
		{
			name: `errorf skip caller`,
			err:  errorfSkipCaller("ooh"),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:127",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertSingleStack(t, test.err)
			var stacked StackTracer
			if !errors.As(test.err, &stacked) {
				t.Fatalf("expected %#v to implement errors.StackTracer", test.err)
			}
			st := stacked.StackTrace()
			for j, want := range test.want {
				assertFormatRegexp(t, st[j], "%+v", want)
			}
		})
	}
}

func wrap(err error) error {
	return errors.Wrap(err)
}

func wrapSkipCaller(err error) error {
	return errors.Wrap(err, errors.SkipCaller())
}

func errorfSkipCaller(message string) error {
	return errors.Errorf(message, errors.SkipCaller())
}

func TestFields(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected interface{}
	}{
		{
			name:     "bool",
			err:      errors.Wrap(errors.Errorf("error"), errors.Bool("key", true)),
			expected: true,
		},
		{
			name:     "int",
			err:      errors.Wrap(errors.Errorf("error"), errors.Int("key", 1)),
			expected: 1,
		},
		{
			name:     "uint",
			err:      errors.Wrap(errors.Errorf("error"), errors.Uint("key", 1)),
			expected: uint(1),
		},
		{
			name:     "float",
			err:      errors.Wrap(errors.Errorf("error"), errors.Float("key", 1.0)),
			expected: 1.0,
		},
		{
			name:     "string",
			err:      errors.Wrap(errors.Errorf("error"), errors.String("key", "value")),
			expected: "value",
		},
		{
			name:     "strings",
			err:      errors.Wrap(errors.Errorf("error"), errors.Strings("key", []string{"value"})),
			expected: []string{"value"},
		},
		{
			name:     "value",
			err:      errors.Wrap(errors.Errorf("error"), errors.Value("key", "value")),
			expected: "value",
		},
		{
			name: "time",
			err: errors.Wrap(
				errors.Errorf("error"),
				errors.Time("key", time.Date(2022, 0o5, 20, 12, 0, 0, 0, time.UTC)),
			),
			expected: time.Date(2022, 0o5, 20, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "duration",
			err:      errors.Wrap(errors.Errorf("error"), errors.Duration("key", time.Hour)),
			expected: time.Hour,
		},
		{
			name:     "JSON",
			err:      errors.Wrap(errors.Errorf("error"), errors.JSON("key", []byte(`{"key":"value"}`))),
			expected: json.RawMessage(`{"key":"value"}`),
		},
		{
			name:     "wrap with stack",
			err:      errors.Wrap(errors.New("error"), errors.String("key", "value")),
			expected: "value",
		},
		{
			name:     "error with fields",
			err:      errors.Errorf("error", errors.String("key", "value")),
			expected: "value",
		},
		{
			name:     "errorf with fields",
			err:      errors.Errorf("error: %s", "string", errors.String("key", "value")),
			expected: "value",
		},
		{
			name:     "errorf with fields and wrapped error",
			err:      errors.Errorf("%w", errors.Errorf("ooh"), errors.String("key", "value")),
			expected: "value",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var err errors.LoggableError
			if !errors.As(test.err, &err) {
				t.Fatalf("expected %#v to implement errors.LoggableError", test.err)
			}
			logger := errorstest.NewLogger()
			err.LogFields(logger)
			logger.AssertField(t, "key", test.expected)
		})
	}
}

func TestWrap_nil(t *testing.T) {
	err := errors.Wrap(nil)
	if err != nil {
		t.Error("want nil error, got:", err)
	}
}

func TestIs(t *testing.T) {
	err := errors.Wrap(errTest)

	is := errors.Is(err, errTest)

	if !is {
		t.Error("want errors is true")
	}
}
