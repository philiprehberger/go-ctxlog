# go-ctxlog

[![CI](https://github.com/philiprehberger/go-ctxlog/actions/workflows/ci.yml/badge.svg)](https://github.com/philiprehberger/go-ctxlog/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/philiprehberger/go-ctxlog.svg)](https://pkg.go.dev/github.com/philiprehberger/go-ctxlog)
[![License](https://img.shields.io/github/license/philiprehberger/go-ctxlog)](LICENSE)

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

## API

| Function | Description |
|----------|-------------|
| `With(ctx, args...)` | Attach slog key-value pairs to context |
| `WithAttrs(ctx, attrs...)` | Attach typed `slog.Attr` values to context |
| `WithRequestID(ctx, id)` | Attach a request ID to context |
| `RequestID(ctx)` | Extract request ID from context (returns `""` if not set) |
| `Logger(ctx)` | Returns `slog.Default()` enriched with all context fields |
| `From(ctx, logger)` | Enrich a specific `*slog.Logger` with context fields |
| `Middleware(next)` | HTTP middleware that generates a UUID v4 request ID |
| `MiddlewareWithHeader(header)` | Middleware that reads request ID from a header, or generates one |

## Development

```bash
go test ./...
go vet ./...
```

## License

MIT
