package errors_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/muonsoft/errors"
)

// TestGroupAttributes tests error attributes with slog groups
func TestGroupAttributes(t *testing.T) {
	t.Run("simple group", func(t *testing.T) {
		err := errors.Wrap(
			errors.New("test error"),
			errors.Group("request",
				slog.String("method", "GET"),
				slog.String("path", "/api/users"),
				slog.Int("status", 404),
			),
		)

		attrs := errors.Attrs(err)
		if len(attrs) != 1 {
			t.Fatalf("expected 1 attribute (group), got %d", len(attrs))
		}

		if attrs[0].Key != "request" {
			t.Errorf("expected group key 'request', got '%s'", attrs[0].Key)
		}

		if attrs[0].Value.Kind() != slog.KindGroup {
			t.Errorf("expected group kind, got %v", attrs[0].Value.Kind())
		}

		groupAttrs := attrs[0].Value.Group()
		if len(groupAttrs) != 3 {
			t.Fatalf("expected 3 attributes in group, got %d", len(groupAttrs))
		}

		// Check group contents
		if groupAttrs[0].Key != "method" || groupAttrs[0].Value.String() != "GET" {
			t.Errorf("unexpected first group attribute: %v", groupAttrs[0])
		}
		if groupAttrs[1].Key != "path" || groupAttrs[1].Value.String() != "/api/users" {
			t.Errorf("unexpected second group attribute: %v", groupAttrs[1])
		}
		if groupAttrs[2].Key != "status" || groupAttrs[2].Value.Any() != int64(404) {
			t.Errorf("unexpected third group attribute: %v", groupAttrs[2])
		}
	})

	t.Run("nested groups", func(t *testing.T) {
		err := errors.Wrap(
			errors.New("test error"),
			errors.Group("outer",
				slog.String("field1", "value1"),
				slog.Group("inner",
					slog.String("field2", "value2"),
					slog.Int("field3", 123),
				),
			),
		)

		attrs := errors.Attrs(err)
		if len(attrs) != 1 {
			t.Fatalf("expected 1 attribute (outer group), got %d", len(attrs))
		}

		outerGroup := attrs[0].Value.Group()
		if len(outerGroup) != 2 {
			t.Fatalf("expected 2 attributes in outer group, got %d", len(outerGroup))
		}

		// Check nested group
		if outerGroup[1].Key != "inner" || outerGroup[1].Value.Kind() != slog.KindGroup {
			t.Errorf("expected nested group 'inner', got %v", outerGroup[1])
		}

		innerGroup := outerGroup[1].Value.Group()
		if len(innerGroup) != 2 {
			t.Fatalf("expected 2 attributes in inner group, got %d", len(innerGroup))
		}
	})

	t.Run("group without key", func(t *testing.T) {
		err := errors.Wrap(
			errors.New("test error"),
			errors.Attr(slog.Group("",
				slog.String("field1", "value1"),
				slog.String("field2", "value2"),
			)),
		)

		attrs := errors.Attrs(err)
		if len(attrs) != 1 {
			t.Fatalf("expected 1 attribute (group without key), got %d", len(attrs))
		}

		// Group without key should still be a group
		if attrs[0].Key != "" {
			t.Errorf("expected empty key, got '%s'", attrs[0].Key)
		}
		if attrs[0].Value.Kind() != slog.KindGroup {
			t.Errorf("expected group kind, got %v", attrs[0].Value.Kind())
		}
	})

	t.Run("groups at different levels", func(t *testing.T) {
		err := errors.Wrap(
			errors.Wrap(
				errors.New("test error"),
				errors.Group("inner", slog.String("a", "1")),
			),
			errors.Group("outer", slog.String("b", "2")),
		)

		attrs := errors.Attrs(err)
		if len(attrs) != 2 {
			t.Fatalf("expected 2 attributes (2 groups), got %d", len(attrs))
		}

		// Groups should be collected from all levels
		hasInner := false
		hasOuter := false
		for _, attr := range attrs {
			if attr.Key == "inner" {
				hasInner = true
			}
			if attr.Key == "outer" {
				hasOuter = true
			}
		}

		if !hasInner || !hasOuter {
			t.Errorf("expected both 'inner' and 'outer' groups, got attrs: %v", attrs)
		}
	})

	t.Run("groups in joined errors", func(t *testing.T) {
		err1 := errors.Wrap(
			errors.New("error 1"),
			errors.Group("group1", slog.String("a", "1")),
		)
		err2 := errors.Wrap(
			errors.New("error 2"),
			errors.Group("group2", slog.String("b", "2")),
		)

		joined := errors.Join(err1, err2)

		attrs := errors.Attrs(joined)
		if len(attrs) < 2 {
			t.Fatalf("expected at least 2 attributes from joined errors, got %d", len(attrs))
		}

		// Check that both groups are present
		hasGroup1 := false
		hasGroup2 := false
		for _, attr := range attrs {
			if attr.Key == "group1" {
				hasGroup1 = true
			}
			if attr.Key == "group2" {
				hasGroup2 = true
			}
		}

		if !hasGroup1 || !hasGroup2 {
			t.Errorf("expected both 'group1' and 'group2', got attrs: %v", attrs)
		}
	})

	t.Run("mixed flat and grouped attributes", func(t *testing.T) {
		err := errors.Wrap(
			errors.New("test error"),
			errors.String("flat1", "value1"),
			errors.Group("grouped",
				slog.String("nested1", "value2"),
				slog.Int("nested2", 42),
			),
			errors.Int("flat2", 100),
		)

		attrs := errors.Attrs(err)
		if len(attrs) != 3 {
			t.Fatalf("expected 3 attributes (2 flat + 1 group), got %d", len(attrs))
		}

		// Check that we have both flat and grouped attributes
		hasFlatString := false
		hasFlatInt := false
		hasGroup := false

		for _, attr := range attrs {
			if attr.Key == "flat1" && attr.Value.String() == "value1" {
				hasFlatString = true
			}
			if attr.Key == "flat2" && attr.Value.Any() == int64(100) {
				hasFlatInt = true
			}
			if attr.Key == "grouped" && attr.Value.Kind() == slog.KindGroup {
				hasGroup = true
			}
		}

		if !hasFlatString || !hasFlatInt || !hasGroup {
			t.Errorf("expected flat1, flat2, and grouped, got attrs: %v", attrs)
		}
	})
}

// TestGroupJSON tests JSON marshaling of grouped attributes
func TestGroupJSON(t *testing.T) {
	t.Run("group with key creates nested object", func(t *testing.T) {
		err := errors.Wrap(
			errors.Errorf("test error"),
			errors.Group("request",
				slog.String("method", "GET"),
				slog.String("path", "/api"),
			),
		)

		data, marshalErr := json.Marshal(err)
		if marshalErr != nil {
			t.Fatalf("failed to marshal error: %v", marshalErr)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		// Check that request is a nested object
		request, ok := result["request"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected 'request' to be a nested object, got %T: %v", result["request"], result["request"])
		}

		if request["method"] != "GET" {
			t.Errorf("expected method=GET, got %v", request["method"])
		}
		if request["path"] != "/api" {
			t.Errorf("expected path=/api, got %v", request["path"])
		}
	})

	t.Run("group without key merges fields", func(t *testing.T) {
		err := errors.Wrap(
			errors.Errorf("test error"),
			errors.Attr(slog.Group("",
				slog.String("field1", "value1"),
				slog.String("field2", "value2"),
			)),
		)

		data, marshalErr := json.Marshal(err)
		if marshalErr != nil {
			t.Fatalf("failed to marshal error: %v", marshalErr)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		// Fields should be at top level (merged)
		if result["field1"] != "value1" {
			t.Errorf("expected field1=value1 at top level, got %v", result["field1"])
		}
		if result["field2"] != "value2" {
			t.Errorf("expected field2=value2 at top level, got %v", result["field2"])
		}
	})

	t.Run("nested groups", func(t *testing.T) {
		err := errors.Wrap(
			errors.Errorf("test error"),
			errors.Group("outer",
				slog.String("field1", "value1"),
				slog.Group("inner",
					slog.String("field2", "value2"),
				),
			),
		)

		data, marshalErr := json.Marshal(err)
		if marshalErr != nil {
			t.Fatalf("failed to marshal error: %v", marshalErr)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		outer, ok := result["outer"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected 'outer' to be a nested object, got %T", result["outer"])
		}

		inner, ok := outer["inner"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected 'inner' to be a nested object, got %T", outer["inner"])
		}

		if inner["field2"] != "value2" {
			t.Errorf("expected inner.field2=value2, got %v", inner["field2"])
		}
	})
}

// TestGroupFormatting tests %+v formatting of grouped attributes
func TestGroupFormatting(t *testing.T) {
	t.Run("group with key uses dot notation", func(t *testing.T) {
		err := errors.Wrap(
			errors.Errorf("test error"),
			errors.Group("request",
				slog.String("method", "GET"),
				slog.String("path", "/api"),
			),
		)

		formatted := strings.TrimSpace(strings.Split(errors.Errorf("%+v", err).Error(), "\n")[0])
		output := errors.Errorf("%+v", err).Error()

		// Check for dot notation in output
		if !strings.Contains(output, "request.method: GET") {
			t.Errorf("expected 'request.method: GET' in output, got:\n%s", output)
		}
		if !strings.Contains(output, "request.path: /api") {
			t.Errorf("expected 'request.path: /api' in output, got:\n%s", output)
		}

		_ = formatted // Use the variable to avoid "declared and not used" error
	})

	t.Run("group without key uses no prefix", func(t *testing.T) {
		err := errors.Wrap(
			errors.Errorf("test error"),
			errors.Attr(slog.Group("",
				slog.String("field1", "value1"),
				slog.String("field2", "value2"),
			)),
		)

		output := errors.Errorf("%+v", err).Error()

		// Fields should appear without prefix
		if !strings.Contains(output, "field1: value1") {
			t.Errorf("expected 'field1: value1' in output, got:\n%s", output)
		}
		if !strings.Contains(output, "field2: value2") {
			t.Errorf("expected 'field2: value2' in output, got:\n%s", output)
		}
	})

	t.Run("nested groups", func(t *testing.T) {
		err := errors.Wrap(
			errors.Errorf("test error"),
			errors.Group("outer",
				slog.Group("inner",
					slog.String("field", "value"),
				),
			),
		)

		output := errors.Errorf("%+v", err).Error()

		// Should show nested dot notation
		if !strings.Contains(output, "outer.inner.field: value") {
			t.Errorf("expected 'outer.inner.field: value' in output, got:\n%s", output)
		}
	})
}

// TestSlogLogValuer tests that wrapped errors implement slog.LogValuer
func TestSlogLogValuer(t *testing.T) {
	t.Run("error implements LogValuer", func(t *testing.T) {
		err := errors.Wrap(
			errors.New("test error"),
			errors.String("key", "value"),
		)

		// Check if error implements slog.LogValuer
		_, ok := err.(slog.LogValuer)
		if !ok {
			t.Error("expected error to implement slog.LogValuer")
		}
	})

	t.Run("LogValue returns group", func(t *testing.T) {
		err := errors.Wrap(
			errors.New("test error"),
			errors.String("key1", "value1"),
			errors.Int("key2", 42),
		)

		logValuer, ok := err.(slog.LogValuer)
		if !ok {
			t.Fatal("error does not implement slog.LogValuer")
		}

		value := logValuer.LogValue()
		if value.Kind() != slog.KindGroup {
			t.Errorf("expected LogValue to return a group, got %v", value.Kind())
		}

		attrs := value.Group()
		if len(attrs) != 2 {
			t.Fatalf("expected 2 attributes in LogValue group, got %d", len(attrs))
		}
	})
}

// TestLogAttrs tests the LogAttrs convenience function
func TestLogAttrsComplete(t *testing.T) {
	t.Run("logs all attributes and stack trace", func(t *testing.T) {
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
			errors.Errorf("database error",
				errors.String("query", "SELECT * FROM users"),
				errors.Int("userID", 123),
			),
		)

		errors.Log(context.Background(), logger, slog.LevelError, err)

		if capturedMsg != "database error" {
			t.Errorf("expected message 'database error', got '%s'", capturedMsg)
		}

		if capturedLevel != slog.LevelError {
			t.Errorf("expected level Error, got %v", capturedLevel)
		}

		// Should have query, userID, and stackTrace attributes
		if len(capturedAttrs) < 3 {
			t.Fatalf("expected at least 3 attributes, got %d: %v", len(capturedAttrs), capturedAttrs)
		}

		hasQuery := false
		hasUserID := false
		hasStackTrace := false

		for _, attr := range capturedAttrs {
			if attr.Key == "query" {
				hasQuery = true
			}
			if attr.Key == "userID" {
				hasUserID = true
			}
			if attr.Key == "stackTrace" {
				hasStackTrace = true
			}
		}

		if !hasQuery {
			t.Error("expected 'query' attribute")
		}
		if !hasUserID {
			t.Error("expected 'userID' attribute")
		}
		if !hasStackTrace {
			t.Error("expected 'stackTrace' attribute")
		}
	})

	t.Run("handles nil error", func(t *testing.T) {
		handlerCalled := false

		handler := &testHandler{
			onHandle: func(ctx context.Context, r slog.Record) error {
				handlerCalled = true
				return nil
			},
		}

		logger := slog.New(handler)

		errors.Log(context.Background(), logger, slog.LevelError, nil)

		if handlerCalled {
			t.Error("expected handler not to be called for nil error")
		}
	})

	t.Run("logs groups correctly", func(t *testing.T) {
		var capturedAttrs []slog.Attr

		handler := &testHandler{
			onHandle: func(ctx context.Context, r slog.Record) error {
				r.Attrs(func(a slog.Attr) bool {
					capturedAttrs = append(capturedAttrs, a)
					return true
				})
				return nil
			},
		}

		logger := slog.New(handler)

		err := errors.Wrap(
			errors.Errorf("test error"),
			errors.Group("request",
				slog.String("method", "POST"),
				slog.Int("status", 500),
			),
		)

		errors.Log(context.Background(), logger, slog.LevelError, err)

		// Should have request group and stackTrace
		hasRequestGroup := false
		for _, attr := range capturedAttrs {
			if attr.Key == "request" && attr.Value.Kind() == slog.KindGroup {
				hasRequestGroup = true
				break
			}
		}

		if !hasRequestGroup {
			t.Errorf("expected 'request' group attribute, got: %v", capturedAttrs)
		}
	})
}
