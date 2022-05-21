package errors_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/muonsoft/errors"
)

func TestStackTrace(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want []string
	}{
		{
			name: "Error()",
			err:  errors.Error("ooh"),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:19",
			},
		},
		{
			name: "Wrap(Error())",
			err:  errors.Wrap(errors.Error("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:27",
			},
		},
		{
			name: "Wrap(New())",
			err:  errors.Wrap(errors.New("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:35",
			},
		},
		{
			name: "Wrap(Wrap(New()))",
			err:  errors.Wrap(errors.Wrap(errors.New("ooh"))),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:43",
			},
		},
		{
			name: "Errorf()",
			err:  errors.Errorf("ooh"),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:51",
			},
		},
		{
			name: `Errorf("%w", New())`,
			err:  errors.Errorf("%v", errors.New("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:59",
			},
		},
		{
			name: `Errorf("%w", Wrap(New()))`,
			err:  errors.Errorf("%w", errors.Wrap(errors.New("ooh"))),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:67",
			},
		},
		{
			name: `Errorf("%%w %v", Wrap(New()))`,
			err:  errors.Errorf("%%w %v", errors.New("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:75",
			},
		},
		{
			name: `Errorf("%s: %w", Wrap(New()))`,
			err:  errors.Errorf("%s: %w", "prefix", errors.New("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:83",
			},
		},
		{
			name: `Errorf("%w", Errorf("%w", New()))`,
			err:  errors.Errorf("%w", errors.Errorf("%w", errors.New("ooh"))),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:91",
			},
		},
		{
			name: `wrap with New()`,
			err:  wrap(errors.New("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.wrap\n" +
					"\t.+/errors/errors_test.go:140",
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:99",
			},
		},
		{
			name: `wrap skip caller`,
			err:  wrapSkipCaller(errors.New("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:109",
			},
		},
		{
			name: `error skip caller`,
			err:  errorSkipCaller("ooh"),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:117",
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

func errorSkipCaller(message string) error {
	return errors.Error(message, errors.SkipCaller())
}

func TestFields(t *testing.T) {
	const key = "key"

	tests := []struct {
		name     string
		err      error
		expected interface{}
	}{
		{
			name:     "bool",
			err:      errors.Wrap(errors.Error("error"), errors.Bool(key, true)),
			expected: true,
		},
		{
			name:     "int",
			err:      errors.Wrap(errors.Error("error"), errors.Int(key, 1)),
			expected: 1,
		},
		{
			name:     "uint",
			err:      errors.Wrap(errors.Error("error"), errors.Uint(key, 1)),
			expected: uint(1),
		},
		{
			name:     "float",
			err:      errors.Wrap(errors.Error("error"), errors.Float(key, 1.0)),
			expected: 1.0,
		},
		{
			name:     "string",
			err:      errors.Wrap(errors.Error("error"), errors.String(key, "value")),
			expected: "value",
		},
		{
			name:     "strings",
			err:      errors.Wrap(errors.Error("error"), errors.Strings(key, []string{"value"})),
			expected: []string{"value"},
		},
		{
			name:     "value",
			err:      errors.Wrap(errors.Error("error"), errors.Value(key, "value")),
			expected: "value",
		},
		{
			name:     "time",
			err:      errors.Wrap(errors.Error("error"), errors.Time(key, time.Date(2022, 05, 20, 12, 0, 0, 0, time.UTC))),
			expected: time.Date(2022, 05, 20, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "duration",
			err:      errors.Wrap(errors.Error("error"), errors.Duration(key, time.Hour)),
			expected: time.Hour,
		},
		{
			name:     "JSON",
			err:      errors.Wrap(errors.Error("error"), errors.JSON(key, []byte(`{"key":"value"}`))),
			expected: json.RawMessage(`{"key":"value"}`),
		},
		{
			name:     "wrap with stack",
			err:      errors.Wrap(errors.New("error"), errors.String(key, "value")),
			expected: "value",
		},
		{
			name:     "error with fields",
			err:      errors.Error("error", errors.String(key, "value")),
			expected: "value",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var err errors.FieldWriter
			if !errors.As(test.err, &err) {
				t.Fatalf("expected %#v to implement errors.FieldWriter", test.err)
			}
			logger := NewLoggerMock()
			err.WriteFieldsTo(logger)
			logger.AssertField(t, key, test.expected)
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
