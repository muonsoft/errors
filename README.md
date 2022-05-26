# errors

Errors package for structured logging. Adds stack trace without a pain
(no confuse with `Wrap`/`WithMessage` methods).

## Key features

This package is based on well known [github.com/pkg/errors](https://github.com/pkg/errors).
Key differences and features:

* `errors.New()` is an alias to standard library and (it does not add a stack trace)
  and should be used to create sentinel package-level errors;
* few methods to create or wrap an error: `errors.Errorf()`, `errors.Wrap()`;
* adds stack trace idempotently (only once in a chain);
* options to skip caller in a stack trace and to add error fields for structured logging;
* error fields are made for the statically typed logger interface;
* package errors can be easily marshaled into JSON with all fields.

## Installation

Run the following command to install the package

```
go get -u github.com/muonsoft/errors
```

## How to use

### `errors.New()` for package-level errors

`errors.New()` is an alias to the standard `errors.New()` function. Use it only for sentinel package-level errors.
This function would not add a stack trace.

```golang
var ErrNotFound      = errors.New("not found")
var errInternalError = errors.New("internal error")

// To initiate a sentinel error with a stack trace it is recommended to use a
// constructor function and wrap the error with errors.Wrap().
// Use errors.SkipCaller() option to remove constructor function from a stack trace.
func NewNotFoundError() error {
	return errors.Wrap(ErrNotFound, errors.SkipCaller())
}
```

### `errors.Errorf()` for wrapping errors with formatted message, fields and stack trace

`errors.Errorf()` is an equivalent to standard `fmt.Errorf()`. It formats according to a format specifier 
and returns the string as a value that satisfies error. You can wrap an error using `%w` modifier.

`errors.Errorf()` also records the stack trace at the point it was called. If the wrapped error
contains a stack trace then a new one will not be added to a chain. Also, you can pass an 
options to set a structured fields or to skip a caller in a stack trace.
Options must be specified after formatting arguments.

```golang
row := repository.db.QueryRow(ctx, findSQL, id)
var product Product
err := row.Scan(&product.ID, &product.Name)
if err != nil {
	// Use errors.Errorf to wrap the library error with the message context and
	// error fields to be used for structured logging.
	return nil, errors.Errorf(
		"%w: %v", errSQLError, err.Error(),
		errors.String("sql", findSQL),
		errors.Int("productID", id),
	)
}
```

### `errors.Wrap()` for wrapping errors with fields and stack trace

`errors.Wrap()` returns an error annotating err with a stack trace at the point `errors.Wrap()` is called.
If the wrapped error contains a stack trace then a new one will not be added to a chain.
If err is nil, Wrap returns nil.  Also, you can pass an options to set a structured fields or to skip a caller
in a stack trace.

```golang
data, err := service.Handle(ctx, userID)
if err != nil {
	// Adds a stack trace to the line that was called (if there is no stack trace in the chain already)
	// and adds fields for structured logging.
	return nil, errors.Wrap(
		err,
		errors.Int("userID", userID),
	)
}
```
