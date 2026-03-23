package ctxlog

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Middleware generates a UUID v4 request ID, injects it into the request
// context via WithRequestID, and sets the X-Request-ID response header.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := newUUID()
		ctx := WithRequestID(r.Context(), id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// MiddlewareWithHeader returns middleware that reads the request ID from the
// specified header. If the header is present and non-empty, its value is used;
// otherwise a new UUID v4 is generated. The request ID is injected into the
// context and set as the X-Request-ID response header.
func MiddlewareWithHeader(headerName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(headerName)
			if id == "" {
				id = newUUID()
			}
			ctx := WithRequestID(r.Context(), id)
			w.Header().Set("X-Request-ID", id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// MiddlewareConfig configures the behavior of MiddlewareWithConfig.
type MiddlewareConfig struct {
	HeaderName   string        // request ID header (default "X-Request-ID")
	GenerateID   func() string // custom ID generator (default UUID v4)
	LogRequests  bool          // log incoming requests
	LogResponses bool          // log completed responses
}

// MiddlewareWithConfig returns middleware configured with the provided options.
// Zero-value fields use sensible defaults: HeaderName defaults to "X-Request-ID"
// and GenerateID defaults to UUID v4 generation.
func MiddlewareWithConfig(cfg MiddlewareConfig) func(http.Handler) http.Handler {
	if cfg.HeaderName == "" {
		cfg.HeaderName = "X-Request-ID"
	}
	if cfg.GenerateID == nil {
		cfg.GenerateID = newUUID
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(cfg.HeaderName)
			if id == "" {
				id = cfg.GenerateID()
			}
			ctx := WithRequestID(r.Context(), id)
			w.Header().Set("X-Request-ID", id)

			if cfg.LogRequests {
				slog.InfoContext(ctx, "request started",
					"method", r.Method,
					"path", r.URL.Path,
					"request_id", id,
				)
			}

			if cfg.LogResponses {
				start := time.Now()
				sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
				next.ServeHTTP(sw, r.WithContext(ctx))
				slog.InfoContext(ctx, "request completed",
					"method", r.Method,
					"path", r.URL.Path,
					"status", sw.status,
					"duration_ms", time.Since(start).Milliseconds(),
					"request_id", id,
				)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// statusWriter wraps http.ResponseWriter to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (sw *statusWriter) WriteHeader(code int) {
	if !sw.wroteHeader {
		sw.status = code
		sw.wroteHeader = true
	}
	sw.ResponseWriter.WriteHeader(code)
}

func (sw *statusWriter) Write(b []byte) (int, error) {
	if !sw.wroteHeader {
		sw.wroteHeader = true
	}
	return sw.ResponseWriter.Write(b)
}

// newUUID generates a UUID v4 using crypto/rand.
func newUUID() string {
	var b [16]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(fmt.Sprintf("ctxlog: failed to generate UUID: %v", err))
	}
	// Set version 4 bits (0100xxxx in byte 6).
	b[6] = (b[6] & 0x0f) | 0x40
	// Set variant bits (10xxxxxx in byte 8).
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
