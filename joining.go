package errors

// Join returns an error that wraps the given errors with a stack trace
// at the point Join is called. Any nil error values are discarded.
// Join returns nil if errs contains no non-nil values.
// The error formats as the concatenation of the strings obtained
// by calling the Error method of each element of errs, with a newline
// between each string.
// If there is only one error in chain, then it's stack trace will be
// preserved if present.
func Join(errs ...error) error {
	n := 0
	for _, err := range errs {
		if err != nil {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	if n == 1 {
		for _, err := range errs {
			if err != nil {
				if isWrapper(err) {
					return err
				}

				return &stacked{
					wrapped: &wrapped{wrapped: err},
					stack:   newStack(0),
				}
			}
		}
	}

	e := &joinError{errs: make([]error, 0, n)}

	for _, err := range errs {
		if err != nil {
			e.errs = append(e.errs, err)
		}
	}

	return &stacked{
		wrapped: &wrapped{wrapped: e},
		stack:   newStack(0),
	}
}

type joinError struct {
	errs []error
}

// todo: add marshal json?

func (e *joinError) Error() string {
	var b []byte

	for i, err := range e.errs {
		if i > 0 {
			b = append(b, '\n')
		}
		b = append(b, err.Error()...)
	}

	return string(b)
}

func (e *joinError) Unwrap() []error {
	return e.errs
}
