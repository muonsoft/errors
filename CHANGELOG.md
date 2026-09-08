# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
this project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.6.0] - 2026-09-08

### Added

- `Log` and `LogLevel` attach a typed `"error"` slog attribute so backends such as
  sentry-go/slog can call `SetException`. The value unwraps to the original chain and
  does not implement `slog.LogValuer`.
- Maintainer-dispatched GitHub Release workflow with changelog finalization and
  CI-owned tag creation.

## [0.5.0] - 2026-02-10

Native `log/slog` integration. This release replaces the custom field system with
`slog.Attr` as the core type for error attributes. The logrus adapter has been
removed.

### Added

- Native `slog.Attr` support in `Wrap` and `Errorf`, including `slog.Group`.
- `Attrs(err)` to extract attributes from an error chain.
- `Log(ctx, logger, err)` and `LogLevel(ctx, logger, level, err)` for slog logging.
- Attribute options `Int64`, `Uint64`, `Float64`, and `Any`.
- `slog.LogValuer` on wrapped errors.

### Changed

- `LoggableError` now requires `Attrs() []slog.Attr` instead of `LogFields(FieldLogger)`.
- `Log(err, logger)` is replaced by `Log(ctx, logger, err)`.
- Minimum Go version is 1.21.

### Deprecated

- `Value` is deprecated in favor of `Any`.

### Removed

- Custom `Field` types and `FieldLogger`.
- logrus adapter. See the [Migration Guide](MIGRATION.md) for a small custom adapter
  if needed.

[Unreleased]: https://github.com/muonsoft/errors/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/muonsoft/errors/releases/tag/v0.5.0
