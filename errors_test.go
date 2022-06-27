package errors_test

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
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
					"\t.+/errors/errors_test.go:23",
			},
		},
		{
			name: "Wrap(Error())",
			err:  errors.Wrap(errors.Errorf("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:31",
			},
		},
		{
			name: "Wrap(New())",
			err:  errors.Wrap(errors.New("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:39",
			},
		},
		{
			name: "Wrap(Wrap(New()))",
			err:  errors.Wrap(errors.Wrap(errors.New("ooh"))),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:47",
			},
		},
		{
			name: "Errorf()",
			err:  errors.Errorf("ooh"),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:55",
			},
		},
		{
			name: `Errorf("%w", New())`,
			err:  errors.Errorf("%v", errors.New("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:63",
			},
		},
		{
			name: `Errorf("%w", Wrap(New()))`,
			err:  errors.Errorf("%w", errors.Wrap(errors.New("ooh"))),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:71",
			},
		},
		{
			name: `Errorf("%%w %v", Wrap(New()))`,
			err:  errors.Errorf("%%w %v", errors.New("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:79",
			},
		},
		{
			name: `Errorf("%s: %w", Wrap(New()))`,
			err:  errors.Errorf("%s: %w", "prefix", errors.New("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:87",
			},
		},
		{
			name: `Errorf("%w", Errorf("%w", New()))`,
			err:  errors.Errorf("%w", errors.Errorf("%w", errors.New("ooh"))),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:95",
			},
		},
		{
			name: `Errorf("%w", fmt.Errorf("%w", Error()))`,
			err:  errors.Errorf("%w", fmt.Errorf("%w", errors.New("ooh"))),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:103",
			},
		},
		{
			name: `wrap with New()`,
			err:  wrap(errors.New("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.wrap\n" +
					"\t.+/errors/errors_test.go:160",
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:111",
			},
		},
		{
			name: `wrap skip caller`,
			err:  wrapSkipCaller(errors.New("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:121",
			},
		},
		{
			name: `wrap skip callers`,
			err:  wrapSkipCallers(errors.New("ooh")),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:129",
			},
		},
		{
			name: `errorf skip caller`,
			err:  errorfSkipCaller("ooh"),
			want: []string{
				"github.com/muonsoft/errors_test.TestStackTrace\n" +
					"\t.+/errors/errors_test.go:137",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertSingleStack(t, test.err)
			stacked, ok := errors.As[StackTracer](test.err)
			if !ok {
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

func wrapSkipCallers(err error) error {
	return errors.Wrap(err, errors.SkipCallers(1))
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
			loggable, ok := errors.As[errors.LoggableError](test.err)
			if !ok {
				t.Fatalf("expected %#v to implement errors.LoggableError", test.err)
			}
			logger := errorstest.NewLogger()
			loggable.LogFields(logger)
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

type timeout interface{ Timeout() bool }

func TestAs(t *testing.T) {
	_, errFileNotFound := os.Open("non-existing")
	poserErr := &poser{"oh no", nil}

	tests := []struct {
		name  string
		err   error
		as    func(err error) (any, bool)
		match bool
		want  any // value of target on match
	}{
		{
			"nil",
			nil,
			func(err error) (any, bool) {
				return errors.As[*fs.PathError](err)
			},
			false,
			nil,
		},
		{
			"wrapped error",
			wrapped{"pitied the fool", errorT{"T"}},
			func(err error) (any, bool) {
				return errors.As[errorT](err)
			},
			true,
			errorT{"T"},
		},
		{
			"match path error",
			errFileNotFound,
			func(err error) (any, bool) {
				return errors.As[*fs.PathError](err)
			},
			true,
			errFileNotFound,
		},
		{
			"not match path error",
			errorT{},
			func(err error) (any, bool) {
				return errors.As[*fs.PathError](err)
			},
			false,
			nil,
		},
		{
			"wrapped nil",
			wrapped{"wrapped", nil},
			func(err error) (any, bool) {
				return errors.As[errorT](err)
			},
			false,
			nil,
		},
		{
			"error with matching as method",
			&poser{"error", nil},
			func(err error) (any, bool) {
				return errors.As[errorT](err)
			},
			true,
			errorT{"poser"},
		},
		{
			"error with matching as method",
			&poser{"path", nil},
			func(err error) (any, bool) {
				return errors.As[*fs.PathError](err)
			},
			true,
			poserPathErr,
		},
		{
			"error with matching as method",
			poserErr,
			func(err error) (any, bool) {
				return errors.As[*poser](err)
			},
			true,
			poserErr,
		},
		{
			"timeout error",
			errors.New("err"),
			func(err error) (any, bool) {
				return errors.As[timeout](err)
			},
			false,
			nil,
		},
		{
			"file not found as timeout",
			errFileNotFound,
			func(err error) (any, bool) {
				return errors.As[timeout](err)
			},
			true,
			errFileNotFound,
		},
		{
			"wrapped file not found as timeout",
			wrapped{"path error", errFileNotFound},
			func(err error) (any, bool) {
				return errors.As[timeout](err)
			},
			true,
			errFileNotFound,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, match := test.as(test.err)
			if match != test.match {
				t.Fatalf("match: got %v; want %v", match, test.match)
			}
			if !match {
				return
			}
			if got != test.want {
				t.Fatalf("got %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestIsOfType(t *testing.T) {
	_, errFileNotFound := os.Open("non-existing")
	poserErr := &poser{"oh no", nil}

	tests := []struct {
		name  string
		err   error
		is    func(err error) bool
		match bool
	}{
		{
			"nil",
			nil,
			func(err error) bool {
				return errors.IsOfType[*os.PathError](err)
			},
			false,
		},
		{
			"wrapped error",
			wrapped{"pitied the fool", errorT{"T"}},
			func(err error) bool {
				return errors.IsOfType[errorT](err)
			},
			true,
		},
		{
			"match path error",
			errFileNotFound,
			func(err error) bool {
				return errors.IsOfType[*fs.PathError](err)
			},
			true,
		},
		{
			"not match path error",
			errorT{},
			func(err error) bool {
				return errors.IsOfType[*fs.PathError](err)
			},
			false,
		},
		{
			"wrapped nil",
			wrapped{"wrapped", nil},
			func(err error) bool {
				return errors.IsOfType[errorT](err)
			},
			false,
		},
		{
			"error with matching as method",
			&poser{"error", nil},
			func(err error) bool {
				return errors.IsOfType[errorT](err)
			},
			true,
		},
		{
			"error with matching as method",
			&poser{"path", nil},
			func(err error) bool {
				return errors.IsOfType[*fs.PathError](err)
			},
			true,
		},
		{
			"error with matching as method",
			poserErr,
			func(err error) bool {
				return errors.IsOfType[*poser](err)
			},
			true,
		},
		{
			"timeout error",
			errors.New("err"),
			func(err error) bool {
				return errors.IsOfType[timeout](err)
			},
			false,
		},
		{
			"file not found as timeout",
			errFileNotFound,
			func(err error) bool {
				return errors.IsOfType[timeout](err)
			},
			true,
		},
		{
			"wrapped file not found as timeout",
			wrapped{"path error", errFileNotFound},
			func(err error) bool {
				return errors.IsOfType[timeout](err)
			},
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			match := test.is(test.err)
			if match != test.match {
				t.Fatalf("match: got %v; want %v", match, test.match)
			}
		})
	}
}

type poser struct {
	msg string
	f   func(error) bool
}

var poserPathErr = &fs.PathError{Op: "poser"}

func (p *poser) Error() string     { return p.msg }
func (p *poser) Is(err error) bool { return p.f(err) }
func (p *poser) As(err any) bool {
	switch x := err.(type) {
	case **poser:
		*x = p
	case *errorT:
		*x = errorT{"poser"}
	case **fs.PathError:
		*x = poserPathErr
	default:
		return false
	}
	return true
}

type errorT struct{ s string }

func (e errorT) Error() string { return fmt.Sprintf("errorT(%s)", e.s) }

type wrapped struct {
	msg string
	err error
}

func (e wrapped) Error() string { return e.msg }

func (e wrapped) Unwrap() error { return e.err }
