# Release notes

## v0.5.0

**Native `log/slog` integration.** This release replaces the custom field system with `slog.Attr` as the core type for error attributes. The logrus adapter has been removed.

### Highlights

- **`slog.Attr` everywhere** — Use `slog.String()`, `slog.Int()`, `slog.Group()`, etc. directly in `errors.Wrap()` and `errors.Errorf()` alongside existing options.
- **`errors.Attrs(err)`** — Extract all attributes from an error chain for custom logging.
- **`errors.Log(ctx, logger, err)`** — Log an error at Error level with attributes and stack trace.
- **`errors.LogLevel(ctx, logger, level, err)`** — Same as above with an explicit level (Debug, Info, Warn, Error).
- **New attribute options** — `Int64`, `Uint64`, `Float64`, `Any`. `Value` is deprecated in favor of `Any`.
- **Grouped attributes** — Full support for `slog.Group` in errors, JSON marshaling, and `%+v` formatting (dot notation).
- **`slog.LogValuer`** — Errors implement `LogValuer` for native slog handling.

### Breaking changes

- Custom `Field` types and `FieldLogger` are removed; use `slog.Attr` and `*slog.Logger`.
- `LoggableError` now requires `Attrs() []slog.Attr` instead of `LogFields(FieldLogger)`.
- `errors.Log(err, logger)` replaced by `errors.Log(ctx, logger, err)`.
- Package requires **Go 1.21+**.
- **logrus adapter removed** — see migration guide for a small custom adapter if needed.

### Migration

**If you are upgrading from v0.4.x, read the [Migration Guide](https://github.com/muonsoft/errors/blob/main/MIGRATION.md)** for step-by-step instructions and before/after examples.
