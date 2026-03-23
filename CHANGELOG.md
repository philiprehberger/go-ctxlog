# Changelog

## 0.2.0

- Add `WithCorrelationID` and `CorrelationID` for correlation ID context propagation
- Add `WithTraceID` and `TraceID` for trace ID context propagation
- Add `Fields` to extract all typed `slog.Attr` values from context
- Add `MiddlewareWithConfig` with `MiddlewareConfig` struct for configurable middleware (custom header, ID generator, request/response logging)
- `From` and `Logger` now include correlation ID and trace ID in logger output

## 0.1.2

- Consolidate README badges onto single line

## 0.1.1

- Add badges and Development section to README

## 0.1.0

- Initial release
- Context-aware logging with `With`, `WithAttrs`, `WithRequestID`
- `Logger` and `From` for extracting enriched loggers
- HTTP middleware with automatic request ID generation
