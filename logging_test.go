package errors_test

import (
	"bytes"
	"context"
	"encoding/json"
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

	errors.Log(context.Background(), logger, err)

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

func TestLog_errorAttrIsTypedError(t *testing.T) {
	orig := errors.Wrap(
		errors.Errorf("typed error", errors.String("user", "john"), errors.Int("id", 123)),
	)

	attr, ok := lastTypedErrorAttr(captureLogAttrs(t, orig))
	if !ok {
		t.Fatal("expected typed 'error' attribute")
	}
	if attr.Value.Kind() == slog.KindLogValuer {
		t.Fatal("expected 'error' attribute not to be KindLogValuer")
	}

	resolved, ok := attr.Value.Resolve().Any().(error)
	if !ok {
		t.Fatalf("expected Resolve().Any() to be error, got %T", attr.Value.Resolve().Any())
	}
	if _, isLogValuer := resolved.(slog.LogValuer); isLogValuer {
		t.Fatal("expected wrapped log error not to implement slog.LogValuer")
	}
	if errors.Unwrap(resolved) != orig {
		t.Fatal("expected Unwrap() to return the original error")
	}
	if !errors.Is(resolved, orig) {
		t.Fatal("expected errors.Is to match the original error")
	}

	var tracer interface{ StackTrace() errors.StackTrace }
	if !stderrors.As(resolved, &tracer) {
		t.Fatal("expected StackTrace() on the unwrapped muonsoft error")
	}
	if len(tracer.StackTrace()) == 0 {
		t.Fatal("expected non-empty stack trace")
	}
}

func TestLog_errorAttrPreservesJoin(t *testing.T) {
	err1 := errors.New("one")
	err2 := errors.New("two")
	joined := errors.Join(err1, err2)

	attr, ok := lastTypedErrorAttr(captureLogAttrs(t, joined))
	if !ok {
		t.Fatal("expected typed 'error' attribute")
	}
	resolved, ok := attr.Value.Resolve().Any().(error)
	if !ok {
		t.Fatalf("expected Resolve().Any() to be error, got %T", attr.Value.Resolve().Any())
	}
	if !errors.Is(resolved, err1) || !errors.Is(resolved, err2) {
		t.Fatal("expected Join chain to be preserved through Unwrap")
	}
}

func TestLog_errorAttrAfterStringErrorField(t *testing.T) {
	orig := errors.Wrap(errors.New("boom"), errors.String("error", "already used"))

	attrs := captureLogAttrs(t, orig)
	typed, ok := lastTypedErrorAttr(attrs)
	if !ok {
		t.Fatal("expected typed 'error' attribute after string 'error' attr")
	}
	if _, isError := typed.Value.Resolve().Any().(error); !isError {
		t.Fatal("expected last 'error' attribute to type-assert to error")
	}

	hasStringError := false
	for _, attr := range attrs {
		if attr.Key == "error" && attr.Value.Kind() == slog.KindString {
			hasStringError = true
			break
		}
	}
	if !hasStringError {
		t.Fatal("expected original string 'error' attribute to remain")
	}
}

func TestSlogAnyErrorIsLogValuer(t *testing.T) {
	orig := errors.Wrap(errors.Errorf("wrap me"), errors.String("key", "value"))

	attr := slog.Any("error", orig)
	if attr.Value.Kind() != slog.KindLogValuer {
		t.Fatalf("expected KindLogValuer, got %v", attr.Value.Kind())
	}
	resolved := attr.Value.Resolve()
	if resolved.Kind() != slog.KindGroup {
		t.Fatalf("expected resolved group, got %v", resolved.Kind())
	}
	if _, ok := resolved.Any().(error); ok {
		t.Fatal("expected slog.Any without wrapper not to resolve to error")
	}
}

func TestLog_jsonHandlerSerializesErrorAsString(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	orig := errors.Wrap(errors.New("boom"), errors.String("user", "john"))

	errors.Log(context.Background(), logger, orig)

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	errorField, ok := payload["error"].(string)
	if !ok {
		t.Fatalf("expected error to be a string, got %T (%v)", payload["error"], payload["error"])
	}
	if errorField != "boom" {
		t.Errorf("expected error %q, got %q", "boom", errorField)
	}
	if payload["msg"] != "boom" {
		t.Errorf("expected msg %q, got %v", "boom", payload["msg"])
	}
}

func TestLogLevel(t *testing.T) {
	for _, level := range []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError} {
		t.Run(level.String(), func(t *testing.T) {
			var capturedLevel slog.Level
			handler := &testHandler{
				onHandle: func(ctx context.Context, r slog.Record) error {
					capturedLevel = r.Level
					return nil
				},
			}
			logger := slog.New(handler)
			err := errors.New("test")
			errors.LogLevel(context.Background(), logger, level, err)
			if capturedLevel != level {
				t.Errorf("expected level %v, got %v", level, capturedLevel)
			}
		})
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

func captureLogAttrs(t *testing.T, err error) []slog.Attr {
	t.Helper()

	var captured []slog.Attr
	handler := &testHandler{
		onHandle: func(ctx context.Context, r slog.Record) error {
			r.Attrs(func(a slog.Attr) bool {
				captured = append(captured, a)
				return true
			})
			return nil
		},
	}
	errors.Log(context.Background(), slog.New(handler), err)
	if len(captured) == 0 {
		t.Fatal("expected log attributes")
	}
	return captured
}

func lastTypedErrorAttr(attrs []slog.Attr) (slog.Attr, bool) {
	var found slog.Attr
	ok := false
	for _, attr := range attrs {
		if attr.Key != "error" {
			continue
		}
		if _, isError := attr.Value.Resolve().Any().(error); isError {
			found = attr
			ok = true
		}
	}
	return found, ok
}
