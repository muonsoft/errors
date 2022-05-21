package errors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"
)

type stackTracer interface {
	StackTrace() StackTrace
}

// Frame represents a program counter inside a stack frame.
// For historical reasons if Frame is interpreted as a uintptr
// its value represents the program counter + 1.
type Frame uintptr

// pc returns the program counter for this frame;
// multiple frames may have the same PC value.
func (f Frame) pc() uintptr { return uintptr(f) - 1 }

// File returns the full path to the file that contains the
// function for this Frame's pc.
func (f Frame) File() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	file, _ := fn.FileLine(f.pc())
	return file
}

// Line returns the line number of source code of the
// function for this Frame's pc.
func (f Frame) Line() int {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.pc())
	return line
}

// Name returns the name of this function, if known.
func (f Frame) Name() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

// Format formats the frame according to the fmt.Formatter interface.
//
//    %s    source file
//    %d    source line
//    %n    function name
//    %v    equivalent to %s:%d
//
// Format accepts flags that alter the printing of some verbs, as follows:
//
//    %+s   function name and path of source file relative to the compile time
//          GOPATH separated by \n\t (<funcname>\n\t<path>)
//    %+v   equivalent to %+s:%d
func (f Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		switch {
		case s.Flag('+'):
			io.WriteString(s, f.Name())
			io.WriteString(s, "\n\t")
			io.WriteString(s, f.File())
		default:
			io.WriteString(s, path.Base(f.File()))
		}
	case 'd':
		io.WriteString(s, strconv.Itoa(f.Line()))
	case 'n':
		io.WriteString(s, funcname(f.Name()))
	case 'v':
		f.Format(s, 's')
		io.WriteString(s, ":")
		f.Format(s, 'd')
	}
}

// String formats a stacktrace Frame as a text string. The output is the
// same as that of fmt.Sprintf("%+v", f), but without newlines or tabs.
func (f Frame) String() string {
	name := f.Name()
	if name == "unknown" {
		return name
	}
	return fmt.Sprintf("%s %s:%d", name, f.File(), f.Line())
}

// MarshalText formats a stacktrace Frame as a text string. The output is the
// same as that of fmt.Sprintf("%+v", f), but without newlines or tabs.
func (f Frame) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

func (f Frame) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Function string `json:"function"`
		File     string `json:"file,omitempty"`
		Line     int    `json:"line,omitempty"`
	}{
		Function: f.Name(),
		File:     f.File(),
		Line:     f.Line(),
	})
}

// StackTrace is stack of Frames from innermost (newest) to outermost (oldest).
type StackTrace []Frame

// Format formats the stack of Frames according to the fmt.Formatter interface.
//
//    %s	lists source files for each Frame in the stack
//    %v	lists the source file and line number for each Frame in the stack
//
// Format accepts flags that alter the printing of some verbs, as follows:
//
//    %+v   Prints filename, function, and line number for each Frame in the stack.
func (st StackTrace) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			for _, f := range st {
				io.WriteString(s, "\n")
				f.Format(s, verb)
			}
		case s.Flag('#'):
			fmt.Fprintf(s, "%#v", []Frame(st))
		default:
			st.formatSlice(s, verb)
		}
	case 's':
		st.formatSlice(s, verb)
	}
}

func (st StackTrace) String() string {
	var s strings.Builder
	for i, frame := range st {
		if i > 0 {
			s.WriteString("\n")
		}
		fmt.Fprintf(&s, "%+v", frame)
	}
	return s.String()
}

func (st StackTrace) Strings() []string {
	s := make([]string, len(st))
	for i, frame := range st {
		s[i] = frame.String()
	}
	return s
}

func (st StackTrace) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteString("[")
	for i, frame := range st {
		if i > 0 {
			b.WriteString(",")
		}
		d, err := frame.MarshalJSON()
		if err != nil {
			return nil, err
		}
		b.Write(d)
	}
	b.WriteString("]")

	return b.Bytes(), nil
}

// formatSlice will format this StackTrace into the given buffer as a slice of
// Frame, only valid when called with '%s' or '%v'.
func (st StackTrace) formatSlice(s fmt.State, verb rune) {
	io.WriteString(s, "[")
	for i, f := range st {
		if i > 0 {
			io.WriteString(s, " ")
		}
		f.Format(s, verb)
	}
	io.WriteString(s, "]")
}

// stack represents a stack of program counters.
type stack []uintptr

func (s *stack) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case st.Flag('+'):
			for _, pc := range *s {
				f := Frame(pc)
				fmt.Fprintf(st, "\n%+v", f)
			}
		}
	}
}

func (s *stack) StackTrace() StackTrace {
	f := make([]Frame, len(*s))
	for i := 0; i < len(f); i++ {
		f[i] = Frame((*s)[i])
	}
	return f
}

func callers(skip int) *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3+skip, pcs[:])
	var st stack = pcs[0:n]

	return &st
}

// funcname removes the path prefix component of a function's name reported by func.Name().
func funcname(name string) string {
	i := strings.LastIndex(name, "/")
	name = name[i+1:]
	i = strings.Index(name, ".")
	return name[i+1:]
}
