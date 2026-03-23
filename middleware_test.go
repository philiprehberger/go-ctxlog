package ctxlog

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMiddlewareSetsRequestIDHeader(t *testing.T) {
	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	id := rec.Header().Get("X-Request-ID")
	if id == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}
	if len(id) != 36 {
		t.Errorf("expected UUID format (36 chars), got %d chars: %s", len(id), id)
	}
}

func TestMiddlewareInjectsRequestIDIntoContext(t *testing.T) {
	var ctxID string
	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxID = RequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if ctxID == "" {
		t.Fatal("expected request ID to be in context")
	}

	headerID := rec.Header().Get("X-Request-ID")
	if ctxID != headerID {
		t.Errorf("context ID %q does not match header ID %q", ctxID, headerID)
	}
}

func TestMiddlewareWithHeaderUsesExistingHeader(t *testing.T) {
	handler := MiddlewareWithHeader("X-Trace-ID")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Trace-ID", "existing-trace-id")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	id := rec.Header().Get("X-Request-ID")
	if id != "existing-trace-id" {
		t.Errorf("expected existing-trace-id, got %s", id)
	}
}

func TestMiddlewareWithHeaderGeneratesIDWhenMissing(t *testing.T) {
	handler := MiddlewareWithHeader("X-Trace-ID")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	id := rec.Header().Get("X-Request-ID")
	if id == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}
	if len(id) != 36 {
		t.Errorf("expected UUID format (36 chars), got %d chars: %s", len(id), id)
	}
}

func TestMiddlewareWithHeaderInjectsIntoContext(t *testing.T) {
	var ctxID string
	handler := MiddlewareWithHeader("X-Trace-ID")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxID = RequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Trace-ID", "my-trace")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if ctxID != "my-trace" {
		t.Errorf("expected context request ID 'my-trace', got %q", ctxID)
	}
}

func TestMiddlewareWithConfigDefaults(t *testing.T) {
	handler := MiddlewareWithConfig(MiddlewareConfig{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	id := rec.Header().Get("X-Request-ID")
	if id == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}
	if len(id) != 36 {
		t.Errorf("expected UUID format (36 chars), got %d chars: %s", len(id), id)
	}
}

func TestMiddlewareWithConfigCustomHeader(t *testing.T) {
	handler := MiddlewareWithConfig(MiddlewareConfig{
		HeaderName: "X-Custom-ID",
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Custom-ID", "custom-123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	id := rec.Header().Get("X-Request-ID")
	if id != "custom-123" {
		t.Errorf("expected custom-123, got %s", id)
	}
}

func TestMiddlewareWithConfigCustomGenerator(t *testing.T) {
	counter := 0
	handler := MiddlewareWithConfig(MiddlewareConfig{
		GenerateID: func() string {
			counter++
			return "generated-id"
		},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	id := rec.Header().Get("X-Request-ID")
	if id != "generated-id" {
		t.Errorf("expected generated-id, got %s", id)
	}
	if counter != 1 {
		t.Errorf("expected generator to be called once, called %d times", counter)
	}
}

func TestMiddlewareWithConfigLogRequests(t *testing.T) {
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))

	handler := MiddlewareWithConfig(MiddlewareConfig{
		LogRequests: true,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	out := buf.String()
	if !strings.Contains(out, "request started") {
		t.Errorf("expected 'request started' log, got: %s", out)
	}
	if !strings.Contains(out, "/test-path") {
		t.Errorf("expected path in log, got: %s", out)
	}
}

func TestMiddlewareWithConfigLogResponses(t *testing.T) {
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))

	handler := MiddlewareWithConfig(MiddlewareConfig{
		LogResponses: true,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/create", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	out := buf.String()
	if !strings.Contains(out, "request completed") {
		t.Errorf("expected 'request completed' log, got: %s", out)
	}
	if !strings.Contains(out, "status=201") {
		t.Errorf("expected status=201 in log, got: %s", out)
	}
	if !strings.Contains(out, "duration_ms=") {
		t.Errorf("expected duration_ms in log, got: %s", out)
	}
}

func TestMiddlewareWithConfigInjectsIntoContext(t *testing.T) {
	var ctxID string
	handler := MiddlewareWithConfig(MiddlewareConfig{
		HeaderName: "X-Correlation-ID",
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxID = RequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-999")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if ctxID != "corr-999" {
		t.Errorf("expected context request ID 'corr-999', got %q", ctxID)
	}
}

func TestMiddlewareWithConfigGeneratesIDWhenHeaderMissing(t *testing.T) {
	handler := MiddlewareWithConfig(MiddlewareConfig{
		HeaderName: "X-Custom-ID",
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	id := rec.Header().Get("X-Request-ID")
	if id == "" {
		t.Fatal("expected X-Request-ID to be set")
	}
	if len(id) != 36 {
		t.Errorf("expected UUID format (36 chars), got %d chars: %s", len(id), id)
	}
}

func TestMiddlewareWithConfigBothLogOptions(t *testing.T) {
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))

	handler := MiddlewareWithConfig(MiddlewareConfig{
		LogRequests:  true,
		LogResponses: true,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/both", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	out := buf.String()
	if !strings.Contains(out, "request started") {
		t.Errorf("expected 'request started' log, got: %s", out)
	}
	if !strings.Contains(out, "request completed") {
		t.Errorf("expected 'request completed' log, got: %s", out)
	}
}
