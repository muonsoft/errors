package errors_test

import (
	"context"
	stderrors "errors"
	"log/slog"
	"testing"

	"github.com/muonsoft/errors"
	"github.com/muonsoft/errors/errorstest"
)

func TestAttrs_noError(t *testing.T) {
	attrs := errors.Attrs(nil)
	if attrs != nil {
		t.Errorf("expected nil attrs for nil error, got %v", attrs)
	}
}

func TestAttrs_errorWithoutStack(t *testing.T) {
	err := errors.New("ooh")
	attrs := errors.Attrs(err)
	if len(attrs) != 0 {
		t.Errorf("expected no attrs for error without fields, got %v", attrs)
	}
}

func TestAttrs_errorWithStack(t *testing.T) {
	err := errors.Wrap(
		errors.Wrap(
			errors.Errorf("ooh", errors.String("deepestKey", "deepestValue")),
			errors.String("deepKey", "deepValue"),
		),
		errors.String("key", "value"),
	)

	attrs := errors.Attrs(err)

	// Should have 3 attributes
	if len(attrs) != 3 {
		t.Fatalf("expected 3 attrs, got %d", len(attrs))
	}

	// Create a mock logger to use assertion helpers
	logger := errorstest.NewLogger()
	logger.Attrs = attrs

	logger.AssertField(t, "key", "value")
	logger.AssertField(t, "deepKey", "deepValue")
	logger.AssertField(t, "deepestKey", "deepestValue")
}

func TestAttrs_joinedErrors(t *testing.T) {
	err := errors.Wrap(
		errors.Join(
			errors.Wrap(
				errors.Errorf("error 1", errors.String("key1", "value1")),
				errors.String("key2", "value2"),
			),
			errors.Errorf("error 2", errors.String("key3", "value3")),
			stderrors.Join(
				errors.Errorf("error 3", errors.String("key4", "value4")),
				errors.Errorf("error 4", errors.String("key5", "value5")),
			),
		),
	)

	attrs := errors.Attrs(err)

	// Should have 5 attributes
	if len(attrs) < 5 {
		t.Fatalf("expected at least 5 attrs, got %d", len(attrs))
	}

	// Create a mock logger to use assertion helpers
	logger := errorstest.NewLogger()
	logger.Attrs = attrs

	logger.AssertField(t, "key1", "value1")
	logger.AssertField(t, "key2", "value2")
	logger.AssertField(t, "key3", "value3")
	logger.AssertField(t, "key4", "value4")
	logger.AssertField(t, "key5", "value5")
}

func TestLog(t *testing.T) {
	// Create a custom handler to capture log output
	var capturedAttrs []slog.Attr
	var capturedMsg string
	var capturedLevel slog.Level

	handler := &testHandler{
		onHandle: func(ctx context.Context, r slog.Record) error {
			capturedMsg = r.Message
			capturedLevel = r.Level
			r.Attrs(func(a slog.Attr) bool {
				capturedAttrs = append(capturedAttrs, a)
				return true
			})
			return nil
		},
	}

	logger := slog.New(handler)

	err := errors.Wrap(
		errors.Errorf("test error", errors.String("user", "john"), errors.Int("id", 123)),
	)

	errors.Log(context.Background(), logger, slog.LevelError, err)

	if capturedMsg != "test error" {
		t.Errorf("expected message 'test error', got '%s'", capturedMsg)
	}

	if capturedLevel != slog.LevelError {
		t.Errorf("expected level Error, got %v", capturedLevel)
	}

	if len(capturedAttrs) < 2 {
		t.Fatalf("expected at least 2 attrs, got %d", len(capturedAttrs))
	}

	// Check for user and id attributes
	hasUser := false
	hasID := false
	hasStackTrace := false

	for _, attr := range capturedAttrs {
		if attr.Key == "user" && attr.Value.String() == "john" {
			hasUser = true
		}
		if attr.Key == "id" && attr.Value.Any() == int64(123) {
			hasID = true
		}
		if attr.Key == "stackTrace" {
			hasStackTrace = true
		}
	}

	if !hasUser {
		t.Error("expected 'user' attribute")
	}
	if !hasID {
		t.Error("expected 'id' attribute")
	}
	if !hasStackTrace {
		t.Error("expected 'stackTrace' attribute")
	}
}

// testHandler is a simple slog.Handler for testing.
type testHandler struct {
	onHandle func(ctx context.Context, r slog.Record) error
}

func (h *testHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *testHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.onHandle != nil {
		return h.onHandle(ctx, r)
	}
	return nil
}

func (h *testHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *testHandler) WithGroup(name string) slog.Handler {
	return h
}
