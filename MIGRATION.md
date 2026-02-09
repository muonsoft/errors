# Migration Guide: v0.4.1 → v0.5.0

This guide will help you migrate from v0.4.1 to v0.5.0, which introduces native `log/slog` integration and removes the custom field system.

## Overview of Changes

Version 0.5.0 is a **breaking change** release that replaces the custom field system with Go's native `log/slog` attributes:

- ❌ **Removed**: `FieldLogger`, `Logger` interfaces
- ❌ **Removed**: All field types (`BoolField`, `IntField`, etc.)
- ❌ **Removed**: `errors.Log(err, logger)` function
- ❌ **Removed**: `logging/logrusadapter` package
- ✅ **Added**: Native `slog.Attr` support
- ✅ **Added**: `slog.LogValuer` implementation
- ✅ **Added**: Grouped attributes via `slog.Group`
- ✅ **Added**: `errors.Attrs(err)` to extract attributes
- ✅ **Added**: `errors.Log(ctx, logger, level, err)` for slog logging
- 📦 **Minimum Go version**: 1.21 (for `log/slog` support)

## Quick Migration Checklist

- [ ] Update Go version to 1.21 or higher
- [ ] Remove imports of `errors/logging/logrusadapter`
- [ ] Replace old `errors.Log(err, logger)` calls with new `errors.Log(ctx, logger, level, err)` or direct slog usage
- [ ] Update custom error types implementing `LoggableError`
- [ ] Update mock loggers in tests to use `errorstest.Logger`
- [ ] Consider using grouped attributes for better structure

## Breaking Changes

### 1. Removed Interfaces and Types

#### Before (v0.4.1)
```go
// These interfaces no longer exist
type FieldLogger interface {
    SetBool(key string, value bool)
    SetInt(key string, value int)
    // ... 9 more methods
}

type Logger interface {
    FieldLogger
    Log(message string)
}

type Field interface {
    Set(logger FieldLogger)
}

// Field types like BoolField, IntField, etc. are removed
```

#### After (v0.5.0)
```go
// Use slog.Attr directly
import "log/slog"

// LoggableError now returns slog.Attr
type LoggableError interface {
    Attrs() []slog.Attr
}
```

### 2. Logging Function Changes

#### Before (v0.4.1)
```go
import "github.com/muonsoft/errors/logging/logrusadapter"

err := errors.Wrap(dbErr, errors.String("table", "users"))

// Log with logrus adapter
logger := logrus.New()
logrusadapter.Log(err, logger)

// Or with custom logger
errors.Log(err, myLogger)
```

#### After (v0.5.0)
```go
import "log/slog"

err := errors.Wrap(dbErr, errors.String("table", "users"))

// Option 1: Use Log convenience function
errors.Log(ctx, slog.Default(), slog.LevelError, err)

// Option 2: Extract attributes and log manually
attrs := errors.Attrs(err)
slog.ErrorContext(ctx, err.Error(), attrsToAny(attrs)...)

// Option 3: Use LogValuer (error logs its attributes automatically)
slog.Error("database error", "error", err)
```

### 3. Custom Error Types

#### Before (v0.4.1)
```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return e.Message
}

func (e *ValidationError) LogFields(logger errors.FieldLogger) {
    logger.SetString("field", e.Field)
    logger.SetString("validation_error", e.Message)
}
```

#### After (v0.5.0)
```go
import "log/slog"

type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return e.Message
}

func (e *ValidationError) Attrs() []slog.Attr {
    return []slog.Attr{
        slog.String("field", e.Field),
        slog.String("validation_error", e.Message),
    }
}
```

### 4. Test Code Changes

#### Before (v0.4.1)
```go
func TestMyError(t *testing.T) {
    err := errors.Wrap(baseErr, errors.String("key", "value"))
    
    loggable, ok := errors.As[errors.LoggableError](err)
    require.True(t, ok)
    
    logger := errorstest.NewLogger()
    loggable.LogFields(logger)
    
    logger.AssertField(t, "key", "value")
}
```

#### After (v0.5.0)
```go
func TestMyError(t *testing.T) {
    err := errors.Wrap(baseErr, errors.String("key", "value"))
    
    attrs := errors.Attrs(err)
    require.NotEmpty(t, attrs)
    
    // Option 1: Check attrs directly
    require.Equal(t, "key", attrs[0].Key)
    require.Equal(t, "value", attrs[0].Value.String())
    
    // Option 2: Use mock logger
    logger := errorstest.NewLogger()
    logger.Attrs = attrs
    logger.AssertField(t, "key", "value")
}
```

## New Features

### 1. Grouped Attributes

Group related attributes together for better structure:

```go
err := errors.Wrap(
    dbErr,
    errors.Group("database",
        slog.String("host", "localhost"),
        slog.Int("port", 5432),
        slog.String("name", "mydb"),
    ),
    errors.Group("query",
        slog.String("sql", "SELECT * FROM users"),
        slog.Duration("duration", 150*time.Millisecond),
    ),
)

// JSON output:
// {
//   "error": "...",
//   "database": {
//     "host": "localhost",
//     "port": 5432,
//     "name": "mydb"
//   },
//   "query": {
//     "sql": "SELECT * FROM users",
//     "duration": "150ms"
//   }
// }

// %+v output:
// error message
// database.host: localhost
// database.port: 5432
// database.name: mydb
// query.sql: SELECT * FROM users
// query.duration: 150ms
```

### 2. Direct slog.Attr Usage

Pass `slog.Attr` directly for maximum flexibility:

```go
err := errors.Wrap(
    err,
    errors.Attr(slog.Int64("timestamp", time.Now().Unix())),
    errors.Attr(slog.Group("metadata",
        slog.String("version", "v1.2.3"),
        slog.Bool("production", true),
    )),
)
```

### 3. Multiple Attributes at Once

```go
commonAttrs := []slog.Attr{
    slog.String("service", "api"),
    slog.String("version", "v1.0.0"),
}

err := errors.Wrap(err, errors.WithAttrs(commonAttrs...))
```

### 4. slog.LogValuer Implementation

Errors automatically work with slog:

```go
err := errors.Wrap(
    dbErr,
    errors.String("table", "users"),
    errors.Int("id", 123),
)

// The error's attributes are automatically included
slog.Error("operation failed", "error", err)
```

## Migration Strategy

### Step 1: Update Dependencies

```bash
go get -u github.com/muonsoft/errors@v0.5.0
go mod tidy
```

Update your `go.mod` to require Go 1.21+:

```go
go 1.21
```

### Step 2: Remove Logrus Adapter

If you were using the logrus adapter:

```go
// Remove this import
- import "github.com/muonsoft/errors/logging/logrusadapter"

// Replace logrusadapter.Log() calls
- logrusadapter.Log(err, logrusLogger)

// Option 1: Switch to slog
+ errors.Log(ctx, slog.Default(), slog.LevelError, err)

// Option 2: Create your own logrus adapter
+ // See "Custom Logger Adapters" section below
```

### Step 3: Update Custom Error Types

Search for types implementing `LogFields(logger FieldLogger)`:

```bash
# Find custom error types
grep -r "LogFields.*FieldLogger" .
```

Update each one:

```diff
- func (e *MyError) LogFields(logger errors.FieldLogger) {
-     logger.SetString("key", e.value)
- }

+ func (e *MyError) Attrs() []slog.Attr {
+     return []slog.Attr{
+         slog.String("key", e.value),
+     }
+ }
```

### Step 4: Update Tests

Replace mock logger usage:

```diff
- logger := errorstest.NewLogger()
- errors.Log(err, logger)
- logger.AssertField(t, "key", "value")

+ attrs := errors.Attrs(err)
+ logger := errorstest.NewLogger()
+ logger.Attrs = attrs
+ logger.AssertField(t, "key", "value")
```

### Step 5: Update Error Logging

Replace `errors.Log()` calls:

```diff
- errors.Log(err, myLogger)

+ // Option 1: Use Log
+ errors.Log(ctx, slog.Default(), slog.LevelError, err)

+ // Option 2: Extract and log
+ attrs := errors.Attrs(err)
+ slog.ErrorContext(ctx, err.Error(), attrsToAny(attrs)...)
```

## Custom Logger Adapters

If you still need to use logrus or another logging library, create a simple adapter:

### Logrus Adapter Example

```go
package myapp

import (
    "log/slog"
    "github.com/muonsoft/errors"
    "github.com/sirupsen/logrus"
)

func LogWithLogrus(err error, logger *logrus.Logger) {
    if err == nil {
        return
    }
    
    // Extract attributes
    attrs := errors.Attrs(err)
    
    // Convert to logrus fields
    fields := logrus.Fields{}
    for _, attr := range attrs {
        fields[attr.Key] = attrValue(attr)
    }
    
    // Log with logrus
    logger.WithFields(fields).Error(err.Error())
}

func attrValue(attr slog.Attr) interface{} {
    if attr.Value.Kind() == slog.KindGroup {
        // Handle groups recursively
        group := make(map[string]interface{})
        for _, a := range attr.Value.Group() {
            group[a.Key] = attrValue(a)
        }
        return group
    }
    return attr.Value.Any()
}
```

### Zerolog Adapter Example

```go
package myapp

import (
    "log/slog"
    "github.com/muonsoft/errors"
    "github.com/rs/zerolog"
)

func LogWithZerolog(err error, logger zerolog.Logger) {
    if err == nil {
        return
    }
    
    event := logger.Error()
    
    // Add attributes
    attrs := errors.Attrs(err)
    for _, attr := range attrs {
        addAttrToZerolog(event, attr)
    }
    
    event.Msg(err.Error())
}

func addAttrToZerolog(event *zerolog.Event, attr slog.Attr) {
    switch attr.Value.Kind() {
    case slog.KindString:
        event.Str(attr.Key, attr.Value.String())
    case slog.KindInt64:
        event.Int64(attr.Key, attr.Value.Int64())
    case slog.KindBool:
        event.Bool(attr.Key, attr.Value.Bool())
    case slog.KindGroup:
        dict := zerolog.Dict()
        for _, a := range attr.Value.Group() {
            addAttrToDict(dict, a)
        }
        event.Dict(attr.Key, dict)
    default:
        event.Interface(attr.Key, attr.Value.Any())
    }
}
```

## Benefits of Migration

### 1. No External Dependencies

v0.5.0 has zero external dependencies. Everything is based on Go standard library.

### 2. Better Structured Logs

Grouped attributes provide better organization:

```json
{
  "error": "database query failed",
  "database": {
    "host": "localhost",
    "port": 5432,
    "name": "production"
  },
  "query": {
    "sql": "SELECT * FROM users WHERE id = ?",
    "params": [123],
    "duration": "150ms"
  }
}
```

### 3. Native slog Integration

Works seamlessly with any slog-compatible logger:

```go
// OpenTelemetry
logger := otelslog.NewHandler(...)

// Custom handler
logger := slog.New(myHandler)

// Works the same way
errors.Log(ctx, logger, slog.LevelError, err)
```

### 4. Simplified Testing

Testing is more straightforward with direct attribute access:

```go
attrs := errors.Attrs(err)
require.Len(t, attrs, 3)
require.Equal(t, "user_id", attrs[0].Key)
require.Equal(t, int64(123), attrs[0].Value.Any())
```

## Troubleshooting

### Issue: "cannot use errors.String() as type Option"

**Cause**: Import path is wrong or mixing v0.4.1 and v0.5.0.

**Solution**: Ensure you're using v0.5.0 consistently:
```bash
go get github.com/muonsoft/errors@v0.5.0
go mod tidy
```

### Issue: "FieldLogger undefined"

**Cause**: Trying to use removed interface.

**Solution**: Implement `LoggableError.Attrs()` instead:
```go
func (e *MyError) Attrs() []slog.Attr {
    return []slog.Attr{
        slog.String("key", e.Value),
    }
}
```

### Issue: "cannot convert attrs to []any"

**Cause**: Trying to use `[]slog.Attr` directly with slog methods.

**Solution**: Convert to `[]any` or use `Log()`:
```go
// Option 1: Use Log
errors.Log(ctx, logger, level, err)

// Option 2: Convert manually
attrs := errors.Attrs(err)
args := make([]any, len(attrs))
for i, a := range attrs {
    args[i] = a
}
slog.ErrorContext(ctx, err.Error(), args...)
```

### Issue: Tests failing with "Fields undefined"

**Cause**: Old mock logger usage.

**Solution**: Update to new mock logger API:
```go
logger := errorstest.NewLogger()
logger.Attrs = errors.Attrs(err)
logger.AssertField(t, "key", "value")
```

## Support

If you encounter issues during migration:

1. Check [examples](examples/) directory for working code
2. Open an [issue](https://github.com/muonsoft/errors/issues) on GitHub
3. Start a [discussion](https://github.com/muonsoft/errors/discussions) for questions

## Summary

The migration to v0.5.0 modernizes the errors package by embracing Go's native `log/slog`. While it requires some code changes, the benefits include:

- ✅ Zero external dependencies
- ✅ Native slog integration
- ✅ Better structured logging with groups
- ✅ Simplified API
- ✅ Future-proof with Go's standard library

Most migrations can be completed in a few hours by following this guide.
