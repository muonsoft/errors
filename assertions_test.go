package errors_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/muonsoft/errors"
	"github.com/muonsoft/errors/errorstest"
)

var errTest = errors.New("test error")

type StackTracer interface {
	StackTrace() errors.StackTrace
}

func assertSingleStack(t *testing.T, err error) {
	t.Helper()

	count := 0
	for e := err; e != nil; e = errors.Unwrap(e) {
		if _, ok := e.(StackTracer); ok {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected %#v to have exactly one stack in chain, got %d", err, count)
	}
}

func assertFormatRegexp(t *testing.T, arg interface{}, format, want string) {
	t.Helper()

	got := fmt.Sprintf(format, arg)
	gotLines := strings.SplitN(got, "\n", -1)
	wantLines := strings.SplitN(want, "\n", -1)

	if len(wantLines) > len(gotLines) {
		t.Errorf("wantLines(%d) > gotLines(%d):\n got: %q\nwant: %q", len(wantLines), len(gotLines), got, want)
		return
	}

	for i, w := range wantLines {
		match, err := regexp.MatchString(w, gotLines[i])
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("line %d: fmt.Sprintf(%q, err):\n got: %q\nwant: %q", i+1, format, got, want)
		}
	}
}

func assertStringsRegexp(t *testing.T, got, want []string) {
	t.Helper()

	if len(want) > len(got) {
		t.Errorf("want(%d) > got(%d):\n got: %q\nwant: %q", len(want), len(got), got, want)
		return
	}

	for i, w := range want {
		match, err := regexp.MatchString(w, got[i])
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("line %d:\n got: %q\nwant: %q", i+1, got, want)
		}
	}
}

func assertStackRegexp(t *testing.T, got, want errorstest.StackTrace) {
	t.Helper()

	if len(want) > len(got) {
		t.Errorf("want(%d) > got(%d):\n got: %q\nwant: %q", len(want), len(got), got, want)
		return
	}

	for i, w := range want {
		match, err := regexp.MatchString(w.Function, got[i].Function)
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("function on line %d:\n got: %q\nwant: %q", i+1, got[i].Function, w.Function)
		}

		match, err = regexp.MatchString(w.File, got[i].File)
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("file on line %d:\n got: %q\nwant: %q", i+1, got[i].File, w.File)
		}

		if w.Line != got[i].Line {
			t.Errorf("line number on line %d:\n got: %d\nwant: %d", i+1, got[i].Line, w.Line)
		}
	}
}
