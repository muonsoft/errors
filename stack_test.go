package errors_test

import (
	"encoding/json"
	"runtime"
	"testing"

	"github.com/muonsoft/errors"
)

var initpc = caller()

type X struct{}

// val returns a Frame pointing to itself.
func (x X) val() errors.Frame {
	return caller()
}

// ptr returns a Frame pointing to itself.
func (x *X) ptr() errors.Frame {
	return caller()
}

func TestFrame_Format(t *testing.T) {
	tests := []struct {
		errors.Frame
		format string
		want   string
	}{
		{
			initpc,
			"%s",
			"stack_test.go",
		},
		{
			initpc,
			"%+s",
			"github.com/muonsoft/errors_test.init\n" +
				"\t.+/errors/stack_test.go",
		},
		{
			0,
			"%s",
			"unknown",
		},
		{
			0,
			"%+s",
			"unknown",
		},
		{
			initpc,
			"%d",
			"11",
		},
		{
			0,
			"%d",
			"0",
		},
		{
			initpc,
			"%n",
			"init",
		},
		{
			func() errors.Frame {
				var x X
				return x.ptr()
			}(),
			"%n",
			`\(\*X\).ptr`,
		},
		{
			func() errors.Frame {
				var x X
				return x.val()
			}(),
			"%n",
			"X.val",
		},
		{
			0,
			"%n",
			"",
		},
		{
			initpc,
			"%v",
			"stack_test.go:11",
		},
		{
			initpc,
			"%+v",
			"github.com/muonsoft/errors_test.init\n" +
				"\t.+/errors/stack_test.go:11",
		},
		{
			0,
			"%v",
			"unknown:0",
		},
	}
	for _, test := range tests {
		t.Run(test.format, func(t *testing.T) {
			assertFormatRegexp(t, test.Frame, test.format, test.want)
		})
	}
}

func TestStackTrace_Format(t *testing.T) {
	tests := []struct {
		errors.StackTrace
		format string
		want   string
	}{
		{
			nil,
			"%s",
			`\[\]`,
		},
		{
			nil,
			"%v",
			`\[\]`,
		},
		{
			nil,
			"%+v",
			"",
		},
		{
			nil,
			"%#v",
			`\[\]errors.Frame\(nil\)`,
		},
		{
			make(errors.StackTrace, 0),
			"%s",
			`\[\]`,
		},
		{
			make(errors.StackTrace, 0),
			"%v",
			`\[\]`,
		},
		{
			make(errors.StackTrace, 0),
			"%+v",
			"",
		},
		{
			make(errors.StackTrace, 0),
			"%#v",
			`\[\]errors.Frame{}`,
		},
		{
			stackTrace()[:2],
			"%s",
			`\[stack_test.go stack_test.go\]`,
		},
		{
			stackTrace()[:2],
			"%v",
			`\[stack_test.go:203 stack_test.go:164\]`,
		},
		{
			stackTrace()[:2],
			"%+v",
			"\n" +
				"github.com/muonsoft/errors_test.stackTrace\n" +
				"\t.+/errors/stack_test.go:203\n" +
				"github.com/muonsoft/errors_test.TestStackTrace_Format\n" +
				"\t.+/errors/stack_test.go:169",
		},
		{
			stackTrace()[:2],
			"%#v",
			`\[\]errors.Frame{stack_test.go:203, stack_test.go:178}`,
		},
	}
	for _, test := range tests {
		t.Run(test.format, func(t *testing.T) {
			assertFormatRegexp(t, test.StackTrace, test.format, test.want)
		})
	}
}

type stack []uintptr

func (s *stack) StackTrace() errors.StackTrace {
	f := make([]errors.Frame, len(*s))
	for i := 0; i < len(f); i++ {
		f[i] = errors.Frame((*s)[i])
	}
	return f
}

func stackTrace() errors.StackTrace {
	const depth = 8
	var pcs [depth]uintptr
	n := runtime.Callers(1, pcs[:])
	var st stack = pcs[0:n]
	return st.StackTrace()
}

// a version of runtime.Caller that returns a Frame, not a uintptr.
func caller() errors.Frame {
	var pcs [3]uintptr
	n := runtime.Callers(2, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	frame, _ := frames.Next()
	return errors.Frame(frame.PC)
}

func TestStackTrace_String(t *testing.T) {
	err := errors.Error("ooh")
	var stacked StackTracer
	if !errors.As(err, &stacked) {
		t.Fatalf("expected %#v to implement errors.StackTracer", err)
	}
	s := stacked.StackTrace().String()

	assertFormatRegexp(t, s, "%s",
		"github.com/muonsoft/errors_test.TestStackTrace_String\n\t.+/errors/stack_test.go:218.*",
	)
}

func TestStackTrace_Strings(t *testing.T) {
	err := errors.Error("ooh")
	var stacked StackTracer
	if !errors.As(err, &stacked) {
		t.Fatalf("expected %#v to implement errors.StackTracer", err)
	}
	s := stacked.StackTrace().Strings()

	assertStringsRegexp(t, s, []string{
		"github.com/muonsoft/errors_test.TestStackTrace_Strings .+/errors/stack_test.go:231",
	})
}

func TestStackTrace_MarshalJSON(t *testing.T) {
	err := errors.Error("ooh")
	var stacked StackTracer
	if !errors.As(err, &stacked) {
		t.Fatalf("expected %#v to implement errors.StackTracer", err)
	}
	st := stacked.StackTrace()
	jsonData, e := json.Marshal(st)
	if e != nil {
		t.Fatalf("expected %#v to be marshalable into json: %v", err, e)
	}
	var s JSONStack
	e = json.Unmarshal(jsonData, &s)
	if e != nil {
		t.Fatalf("failed to unmarshal json: %v", e)
	}

	assertStackRegexp(t, s, JSONStack{
		{
			Function: "github.com/muonsoft/errors_test.TestStackTrace_MarshalJSON",
			File:     ".+/errors/stack_test.go",
			Line:     244,
		},
	})
}
