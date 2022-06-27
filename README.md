# errors

[![Go Reference](https://pkg.go.dev/badge/github.com/muonsoft/errors.svg)](https://pkg.go.dev/github.com/muonsoft/errors)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/muonsoft/errors)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/muonsoft/errors)
![GitHub](https://img.shields.io/github/license/muonsoft/errors)
[![tests](https://github.com/muonsoft/errors/actions/workflows/tests.yml/badge.svg)](https://github.com/muonsoft/errors/actions/workflows/tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/muonsoft/errors)](https://goreportcard.com/report/github.com/muonsoft/errors)
[![Maintainability](https://api.codeclimate.com/v1/badges/fe1720426006f3af30b0/maintainability)](https://codeclimate.com/github/muonsoft/errors/maintainability)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.0-4baaaa.svg)](CODE_OF_CONDUCT.md)

Errors package for structured logging. Adds stack trace without a pain
(no confuse with `Wrap`/`WithMessage` methods).

## Key features

This package is based on well known [github.com/pkg/errors](https://github.com/pkg/errors).
Key differences and features:

* `errors.New()` is an alias to standard library and (it does not add a stack trace)
  and should be used to create sentinel package-level errors;
* minimalistic API: few methods to wrap an error: `errors.Errorf()`, `errors.Wrap()`;
* adds stack trace idempotently (only once in a chain);
* `errors.As()` method is based on typed parameters (aka generics);
* options to skip caller in a stack trace and to add error fields for structured logging;
* error fields are made for the statically typed logger interface;
* package errors can be easily marshaled into JSON with all fields in a chain.

## Additional features

* `errors.IsOfType[T any](err error)` to test for error types.

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
data, err := service.Handle(ctx, userID, message)
if err != nil {
	// Adds a stack trace to the line that was called (if there is no stack trace in the chain already)
	// and adds fields for structured logging.
	return nil, errors.Wrap(
		err,
		errors.Int("userID", userID),
		errors.String("userMessage", message),
	)
}
```

### Printing error with stack trace

You can use formatting with `%+v` modifier to print errors with message, fields for logging and a stack trace.

Example

```golang
func main() {
	err := errors.Errorf(
		"sql error: %w", sql.ErrNoRows,
		errors.String("sql", "SELECT id, name FROM product WHERE id = ?"),
		errors.Int("productID", 123),
	)
	err = errors.Errorf(
		"find product: %w", err,
		errors.String("requestID", "24874020-cab7-4ef3-bac5-76858832f8b0"),
	)
	fmt.Printf("%+v", err)
}
```

Output

```
find product: sql error: sql: no rows in result set
requestID: 24874020-cab7-4ef3-bac5-76858832f8b0
sql: SELECT id, name FROM product WHERE id = ?
productID: 123
main.main
    /home/user/project/main.go:11
runtime.main
    /usr/local/go/src/runtime/proc.go:250
runtime.goexit
    /usr/local/go/src/runtime/asm_amd64.s:1571
```

### Marshal error into JSON

Wrapped errors implements `json.Marshaler` interface. So you can easily marshal errors into JSON.

Example

```golang
func main() {
	err := errors.Errorf(
		"sql error: %w", sql.ErrNoRows,
		errors.String("sql", "SELECT id, name FROM product WHERE id = ?"),
		errors.Int("productID", 123),
	)
	err = errors.Errorf(
		"find product: %w", err,
		errors.String("requestID", "24874020-cab7-4ef3-bac5-76858832f8b0"),
	)
	errJSON, err := json.MarshalIndent(err, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(errJSON))
}
```

Output

```json
{
    "error": "find product: sql error: sql: no rows in result set",
    "productID": 123,
    "requestID": "24874020-cab7-4ef3-bac5-76858832f8b0",
    "sql": "SELECT id, name FROM product WHERE id = ?",
    "stackTrace": [
        {
            "function": "main.main",
            "file": "/home/user/project/main.go",
            "line": 13
        },
        {
            "function": "runtime.main",
            "file": "/usr/local/go/src/runtime/proc.go",
            "line": 250
        },
        {
            "function": "runtime.goexit",
            "file": "/usr/local/go/src/runtime/asm_amd64.s",
            "line": 1571
        }
    ]
}
```

### Structured logging

To use structured logging, you need to use an adapter for your logging system. It can be one of the 
built-in adapters from the `logging` directory, or you can implement your own adapter using `errors.Logger` interface.

Example of using an adapter for [Logrus](https://github.com/sirupsen/logrus).

```golang
err := errors.Errorf(
	"sql error: %w", sql.ErrNoRows,
	errors.String("sql", "SELECT id, name FROM product WHERE id = ?"),
	errors.Int("productID", 123),
)
err = errors.Errorf(
	"find product: %w", err,
	errors.String("requestID", "24874020-cab7-4ef3-bac5-76858832f8b0"),
)
logger := logrus.New()
logrusadapter.Log(err, logger)
```

Output

```
ERRO[0000] find product: sql error: sql: no rows in result set  productID=123 requestID=24874020-cab7-4ef3-bac5-76858832f8b0 sql="SELECT id, name FROM product WHERE id = ?" stackTrace="[{main.main /home/strider/projects/errors/var/scratch.go 12} {runtime.main /usr/local/go/src/runtime/proc.go 250} {runtime.goexit /usr/local/go/src/runtime/asm_amd64.s 1571}]"
```

## Contributing

You may help this project by

* reporting an [issue](https://github.com/muonsoft/errors/issues);
* making translations for error messages;
* suggest an improvement or [discuss](https://github.com/muonsoft/errors/discussions) the usability of the package.

If you'd like to contribute, see [the contribution guide](CONTRIBUTING.md). Pull requests are welcome.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
