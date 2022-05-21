# errors

Errors package for structured logging. Adds stack trace without a pain 
(no confuse with `Wrap`/`WithMessage` methods).

## Key features

This package is based on well known [github.com/pkg/errors](https://github.com/pkg/errors).
Key differences and features:

* `errors.New()` is an alias to standard library and (it does not add a stack trace)
  and should be used to create sentinel package-level errors;
* few methods to create or wrap an error: `errors.Error()`, `errors.Errorf()`, `errors.Wrap()`;
* adds stack trace idempotently (only once in a chain);
* options to skip caller in a stack trace and to add error fields for structured logging;
* package errors can be easily marshaled into JSON with all fields.

## Installation

Run the following command to install the package

```
go get -u github.com/muonsoft/errors
```
