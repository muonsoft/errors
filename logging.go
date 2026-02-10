package errors

import (
	"context"
	"errors"
	"log/slog"
)

// LoggableError is an interface for errors that provide structured attributes
// for logging. Implement this interface on custom error types to add fields
// to structured logs.
type LoggableError interface {
	Attrs() []slog.Attr
}

// Attrs extracts all structured attributes from an error chain.
// It traverses the error chain via Unwrap() and collects attributes from
// any error that implements LoggableError. For joined errors (multiple unwrapped
// errors), it recursively extracts attributes from all branches.
func Attrs(err error) []slog.Attr {
	if err == nil {
		return nil
	}
	return attrsFromError(err)
}

func attrsFromError(err error) []slog.Attr {
	var attrs []slog.Attr

	for e := err; e != nil; e = errors.Unwrap(e) {
		if loggable, ok := e.(LoggableError); ok {
			attrs = append(attrs, loggable.Attrs()...)
		}

		// Handle joined errors (multiple unwrapped errors)
		if joined, ok := e.(interface{ Unwrap() []error }); ok {
			for _, u := range joined.Unwrap() {
				attrs = append(attrs, attrsFromError(u)...)
			}
		}
	}

	return attrs
}

// Log logs an error at Error level with all its structured attributes and stack trace
// using the provided slog.Logger. It is a shorthand for LogLevel(ctx, logger, slog.LevelError, err).
//
// If err is nil, this function does nothing.
//
// Example:
//
//	err := errors.Wrap(dbErr, errors.String("query", sql), errors.Int("userID", 123))
//	errors.Log(ctx, slog.Default(), err)
func Log(ctx context.Context, logger *slog.Logger, err error) {
	LogLevel(ctx, logger, slog.LevelError, err)
}

// LogLevel logs an error at the specified level with all its structured attributes
// and stack trace using the provided slog.Logger.
//
// If err is nil, this function does nothing.
//
// Example:
//
//	errors.LogLevel(ctx, slog.Default(), slog.LevelWarn, err)
//	errors.LogLevel(ctx, slog.Default(), slog.LevelError, err)
func LogLevel(ctx context.Context, logger *slog.Logger, level slog.Level, err error) {
	if err == nil {
		return
	}

	// Collect all attributes
	attrs := Attrs(err)

	// Find and add stack trace if present
	for e := err; e != nil; e = errors.Unwrap(e) {
		if s, ok := e.(stackTracer); ok {
			attrs = append(attrs, slog.Any("stackTrace", s.StackTrace()))
			break
		}
	}

	// Convert attrs to []any for logger.Log
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}

	logger.Log(ctx, level, err.Error(), args...)
}
