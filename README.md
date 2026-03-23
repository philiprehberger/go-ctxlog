# go-ctxlog

[![CI](https://github.com/philiprehberger/go-ctxlog/actions/workflows/ci.yml/badge.svg)](https://github.com/philiprehberger/go-ctxlog/actions/workflows/ci.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/philiprehberger/go-ctxlog.svg)](https://pkg.go.dev/github.com/philiprehberger/go-ctxlog) [![License](https://img.shields.io/github/license/philiprehberger/go-ctxlog)](LICENSE)

Context-aware structured logging helpers for Go's `log/slog`

## Installation

```bash
go get github.com/philiprehberger/go-ctxlog
```

## Usage

```go
import "github.com/philiprehberger/go-ctxlog"
```

### Attaching fields to context

```go
ctx := ctxlog.With(ctx, "user_id", 42, "tenant", "acme")
ctx = ctxlog.WithAttrs(ctx, slog.String("service", "api"))
ctx = ctxlog.WithRequestID(ctx, "req-abc-123")
```

### Correlation ID

Attach a correlation ID to group related requests across services:

```go
ctx = ctxlog.WithCorrelationID(ctx, "corr-abc-123")

// Later, extract it
id := ctxlog.CorrelationID(ctx) // "corr-abc-123"

// Automatically included in logger output as "correlation_id"
ctxlog.Logger(ctx).Info("processing order")
```

### Trace ID

Attach a trace ID for distributed tracing:

```go
ctx = ctxlog.WithTraceID(ctx, "trace-xyz-789")

// Later, extract it
id := ctxlog.TraceID(ctx) // "trace-xyz-789"

// Automatically included in logger output as "trace_id"
ctxlog.Logger(ctx).Info("calling downstream service")
```

### Extracting fields

Retrieve all typed `slog.Attr` values attached via `WithAttrs`:

```go
ctx = ctxlog.WithAttrs(ctx, slog.String("service", "api"), slog.Int("port", 8080))
attrs := ctxlog.Fields(ctx) // []slog.Attr{slog.String("service", "api"), slog.Int("port", 8080)}
```

### Logging with context fields

```go
// Uses slog.Default() enriched with context fields
ctxlog.Logger(ctx).Info("request handled", "status", 200)

// Enrich a specific logger
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
ctxlog.From(ctx, logger).Error("something failed", "err", err)
```

### HTTP middleware

```go
mux := http.NewServeMux()
mux.HandleFunc("/", handler)

// Auto-generate request IDs
http.ListenAndServe(":8080", ctxlog.Middleware(mux))
```

```go
// Use an existing header for the request ID
wrapped := ctxlog.MiddlewareWithHeader("X-Trace-ID")(mux)
http.ListenAndServe(":8080", wrapped)
```

### Configurable middleware

Use `MiddlewareWithConfig` for full control over request ID generation, header names, and request/response logging:

```go
wrapped := ctxlog.MiddlewareWithConfig(ctxlog.MiddlewareConfig{
    HeaderName:   "X-Correlation-ID",        // read ID from this header (default "X-Request-ID")
    GenerateID:   func() string { return myCustomID() }, // custom ID generator
    LogRequests:  true,                       // log incoming requests
    LogResponses: true,                       // log completed responses with status and duration
})(mux)
http.ListenAndServe(":8080", wrapped)
```

Zero-value fields use sensible defaults (header `"X-Request-ID"`, UUID v4 generator, no logging).

## API

| Function | Description |
|----------|-------------|
| `With(ctx, args...)` | Attach slog key-value pairs to context |
| `WithAttrs(ctx, attrs...)` | Attach typed `slog.Attr` values to context |
| `WithRequestID(ctx, id)` | Attach a request ID to context |
| `RequestID(ctx)` | Extract request ID from context (returns `""` if not set) |
| `WithCorrelationID(ctx, id)` | Attach a correlation ID to context |
| `CorrelationID(ctx)` | Extract correlation ID from context (returns `""` if not set) |
| `WithTraceID(ctx, id)` | Attach a trace ID to context |
| `TraceID(ctx)` | Extract trace ID from context (returns `""` if not set) |
| `Fields(ctx)` | Extract all typed `slog.Attr` values from context |
| `Logger(ctx)` | Returns `slog.Default()` enriched with all context fields |
| `From(ctx, logger)` | Enrich a specific `*slog.Logger` with context fields |
| `Middleware(next)` | HTTP middleware that generates a UUID v4 request ID |
| `MiddlewareWithHeader(header)` | Middleware that reads request ID from a header, or generates one |
| `MiddlewareWithConfig(cfg)` | Configurable middleware with custom header, ID generator, and logging |

## Development

```bash
go test ./...
go vet ./...
```

## License

MIT
