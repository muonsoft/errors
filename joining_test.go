package errors_test

import (
	"fmt"
	"testing"

	"github.com/muonsoft/errors"
)

func TestJoin_ReturnsNil(t *testing.T) {
	if err := errors.Join(); err != nil {
		t.Errorf("errors.Join() = %v, want nil", err)
	}
	if err := errors.Join(nil); err != nil {
		t.Errorf("errors.Join(nil) = %v, want nil", err)
	}
	if err := errors.Join(nil, nil); err != nil {
		t.Errorf("errors.Join(nil, nil) = %v, want nil", err)
	}
}

func TestJoin(t *testing.T) {
	err1 := errors.New("err1")
	err2 := errors.New("err2")
	tests := []struct {
		errs []error
		want []error
	}{
		{
			errs: []error{err1},
			want: []error{err1},
		},
		{
			errs: []error{err1, err2},
			want: []error{err1, err2},
		},
		{
			errs: []error{err1, nil, err2},
			want: []error{err1, err2},
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.errs), func(t *testing.T) {
			got := errors.Join(test.errs...)
			for _, want := range test.want {
				if !errors.Is(got, want) {
					t.Errorf("want err %v in chain", want)
				}
			}
		})
	}
}

func TestJoin_ErrorMethod(t *testing.T) {
	err1 := errors.New("err1")
	err2 := errors.New("err2")
	tests := []struct {
		errs []error
		want string
	}{
		{
			errs: []error{err1},
			want: "err1",
		},
		{
			errs: []error{err1, err2},
			want: "err1\nerr2",
		},
		{
			errs: []error{err1, nil, err2},
			want: "err1\nerr2",
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.errs), func(t *testing.T) {
			got := errors.Join(test.errs...).Error()
			if got != test.want {
				t.Errorf("Join().Error() = %q; want %q", got, test.want)
			}
		})
	}
}
