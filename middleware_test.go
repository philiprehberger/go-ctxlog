package ctxlog

import (
	"net/http"
	"net/http/httptest"
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
