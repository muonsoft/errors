# errors

[![Go Reference](https://pkg.go.dev/badge/github.com/muonsoft/errors.svg)](https://pkg.go.dev/github.com/muonsoft/errors)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/muonsoft/errors)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/muonsoft/errors)
![GitHub](https://img.shields.io/github/license/muonsoft/errors)
[![tests](https://github.com/muonsoft/errors/actions/workflows/tests.yml/badge.svg)](https://github.com/muonsoft/errors/actions/workflows/tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/muonsoft/errors)](https://goreportcard.com/report/github.com/muonsoft/errors)
[![Maintainability](https://api.codeclimate.com/v1/badges/fe1720426006f3af30b0/maintainability)](https://codeclimate.com/github/muonsoft/errors/maintainability)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.0-4baaaa.svg)](CODE_OF_CONDUCT.md)

Errors package for structured logging with native `log/slog` integration. Adds stack trace without a pain
(no confuse with `Wrap`/`WithMessage` methods).

> **⚠️ Breaking Changes in v0.5.0**  
> Version 0.5.0 replaces the custom field system with native `log/slog` integration.  
> If you're migrating from v0.4.1, please read the [Migration Guide](MIGRATION.md).

## Key features

This package is based on well known [github.com/pkg/errors](https://github.com/pkg/errors).
Key differences and features:

* `errors.New()` is an alias to standard library and (it does not add a stack trace)
  and should be used to create sentinel package-level errors;
* minimalistic API: few methods to wrap an error: `errors.Errorf()`, `errors.Wrap()`;
* adds stack trace idempotently (only once in a chain);
* `errors.As()` method is based on typed parameters (aka generics);
* options to skip caller in a stack trace and to add error attributes for structured logging;
* **native integration with Go's `log/slog`** - error attributes use `slog.Attr`;
* **supports grouped attributes** via `slog.Group`;
* implements `slog.LogValuer` for seamless slog integration;
* package errors can be easily marshaled into JSON with all fields in a chain.

## Additional features

* `errors.IsOfType[T any](err error)` to test for error types.
* `errors.Attrs(err error) []slog.Attr` to extract all attributes from error chain.
* `errors.Log(ctx, logger, level, err)` convenience function for logging with slog.

## Installation

Run the following command to install the package

```
go get -u github.com/muonsoft/errors
```

Requires Go 1.21+ for `log/slog` support.

**Migrating from v0.4.1?** See the [Migration Guide](MIGRATION.md) for detailed instructions.

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

### `errors.Errorf()` for wrapping errors with formatted message, attributes and stack trace

`errors.Errorf()` is an equivalent to standard `fmt.Errorf()`. It formats according to a format specifier 
and returns the string as a value that satisfies error. You can wrap an error using `%w` modifier.

`errors.Errorf()` also records the stack trace at the point it was called. If the wrapped error
contains a stack trace then a new one will not be added to a chain. Also, you can pass 
options to set structured attributes or to skip a caller in a stack trace.
Options must be specified after formatting arguments.

```golang
row := repository.db.QueryRow(ctx, findSQL, id)
var product Product
err := row.Scan(&product.ID, &product.Name)
if err != nil {
	// Use errors.Errorf to wrap the library error with the message context and
	// error attributes to be used for structured logging.
	return nil, errors.Errorf(
		"%w: %v", errSQLError, err.Error(),
		errors.String("sql", findSQL),
		errors.Int("productID", id),
	)
}
```

### `errors.Wrap()` for wrapping errors with attributes and stack trace

`errors.Wrap()` returns an error annotating err with a stack trace at the point `errors.Wrap()` is called.
If the wrapped error contains a stack trace then a new one will not be added to a chain.
If err is nil, Wrap returns nil. Also, you can pass options to set structured attributes or to skip a caller
in a stack trace.

```golang
data, err := service.Handle(ctx, userID, message)
if err != nil {
	// Adds a stack trace to the line that was called (if there is no stack trace in the chain already)
	// and adds attributes for structured logging.
	return nil, errors.Wrap(
		err,
		errors.Int("userID", userID),
		errors.String("userMessage", message),
	)
}
```

### Working with grouped attributes

The package supports grouped attributes via `slog.Group`, allowing you to organize related attributes:

```golang
err := errors.Wrap(
	dbErr,
	errors.Group("request",
		slog.String("method", "POST"),
		slog.String("path", "/api/users"),
		slog.Int("status", 500),
	),
	errors.Group("database",
		slog.String("query", "INSERT INTO users..."),
		slog.Duration("duration", 150*time.Millisecond),
	),
)
```

You can also use `errors.Attr()` to pass `slog.Attr` directly:

```golang
err := errors.Wrap(
	err,
	errors.Attr(slog.Int64("timestamp", time.Now().Unix())),
	errors.Attr(slog.Group("metadata",
		slog.String("version", "v1.2.3"),
		slog.Bool("production", true),
	)),
)
```

### Printing error with stack trace

You can use formatting with `%+v` modifier to print errors with message, attributes and stack trace.
Grouped attributes are displayed using dot notation.

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
		errors.Group("request",
			slog.String("id", "24874020-cab7-4ef3-bac5-76858832f8b0"),
			slog.String("method", "GET"),
		),
	)
	fmt.Printf("%+v", err)
}
```

Output

```
find product: sql error: sql: no rows in result set
request.id: 24874020-cab7-4ef3-bac5-76858832f8b0
request.method: GET
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

Wrapped errors implement `json.Marshaler` interface. Grouped attributes are marshaled as nested objects.

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
		errors.Group("request",
			slog.String("id", "24874020-cab7-4ef3-bac5-76858832f8b0"),
			slog.String("method", "GET"),
		),
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
    "request": {
        "id": "24874020-cab7-4ef3-bac5-76858832f8b0",
        "method": "GET"
    },
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

### Structured logging with slog

The package provides native integration with Go's `log/slog`. Errors implement `slog.LogValuer`,
so they work seamlessly with any slog logger.

#### Using Log convenience function

```golang
err := errors.Errorf(
	"database query failed: %w", dbErr,
	errors.String("query", "SELECT * FROM users WHERE id = ?"),
	errors.Int("userID", 123),
	errors.Group("performance",
		slog.Duration("duration", 250*time.Millisecond),
		slog.Int("retries", 3),
	),
)

// Log error with all attributes and stack trace
errors.Log(ctx, slog.Default(), slog.LevelError, err)
```

#### Extracting attributes manually

```golang
err := errors.Errorf(
	"operation failed: %w", someErr,
	errors.String("operation", "user.create"),
	errors.Int("userID", 123),
)

// Extract all attributes from error chain
attrs := errors.Attrs(err)

// Use with slog
slog.Error("request failed", append([]any{slog.Any("error", err)}, attrsToAny(attrs)...)...)
```

#### Using slog.LogValuer

Errors automatically work as `slog.LogValuer`, so you can log them directly:

```golang
err := errors.Wrap(
	dbErr,
	errors.String("table", "users"),
	errors.Int("id", 123),
)

// The error will automatically provide its attributes to slog
slog.Error("database error", "error", err)
```

### Custom LoggableError types

You can implement `errors.LoggableError` interface on your custom error types to provide
structured attributes:

```golang
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %s", e.Message)
}

// Implement errors.LoggableError
func (e *ValidationError) Attrs() []slog.Attr {
	return []slog.Attr{
		slog.String("field", e.Field),
		slog.Any("value", e.Value),
		slog.String("validation_message", e.Message),
	}
}

// Usage
err := &ValidationError{
	Field:   "email",
	Value:   "invalid-email",
	Message: "must be a valid email address",
}
wrapped := errors.Wrap(err, errors.String("operation", "user.create"))

// All attributes from ValidationError will be included
attrs := errors.Attrs(wrapped)
```

## Available attribute options

The package provides convenience functions for creating attributes:

```golang
errors.Bool(key string, value bool)
errors.Int(key string, value int)
errors.Uint(key string, value uint)
errors.Float(key string, value float64)
errors.String(key string, value string)
errors.Stringer(key string, value fmt.Stringer)
errors.Strings(key string, values []string)
errors.Value(key string, value interface{})
errors.Time(key string, value time.Time)
errors.Duration(key string, value time.Duration)
errors.JSON(key string, value json.RawMessage)

// New slog-specific options
errors.Attr(attr slog.Attr)                    // Add any slog.Attr directly
errors.WithAttrs(attrs ...slog.Attr)           // Add multiple slog.Attr values
errors.Group(key string, attrs ...slog.Attr)   // Create a grouped attribute
```

## Contributing

You may help this project by

* reporting an [issue](https://github.com/muonsoft/errors/issues);
* making translations for error messages;
* suggest an improvement or [discuss](https://github.com/muonsoft/errors/discussions) the usability of the package.

If you'd like to contribute, see [the contribution guide](CONTRIBUTING.md). Pull requests are welcome.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
