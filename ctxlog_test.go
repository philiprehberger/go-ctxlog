package ctxlog

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestWithAddsFieldsToContext(t *testing.T) {
	ctx := context.Background()
	ctx = With(ctx, "key1", "value1", "key2", 42)

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: stripTimeAndLevel,
	}))
	From(ctx, logger).Info("test")

	out := buf.String()
	if !strings.Contains(out, "key1=value1") {
		t.Errorf("expected key1=value1 in output, got: %s", out)
	}
	if !strings.Contains(out, "key2=42") {
		t.Errorf("expected key2=42 in output, got: %s", out)
	}
}

func TestWithAttrsAddsTypedAttrs(t *testing.T) {
	ctx := context.Background()
	ctx = WithAttrs(ctx,
		slog.String("service", "api"),
		slog.Int("port", 8080),
	)

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: stripTimeAndLevel,
	}))
	From(ctx, logger).Info("test")

	out := buf.String()
	if !strings.Contains(out, "service=api") {
		t.Errorf("expected service=api in output, got: %s", out)
	}
	if !strings.Contains(out, "port=8080") {
		t.Errorf("expected port=8080 in output, got: %s", out)
	}
}

func TestWithRequestIDRoundTrip(t *testing.T) {
	ctx := context.Background()
	id := "abc-123-def"
	ctx = WithRequestID(ctx, id)

	got := RequestID(ctx)
	if got != id {
		t.Errorf("expected request ID %q, got %q", id, got)
	}
}

func TestRequestIDReturnsEmptyWhenNotSet(t *testing.T) {
	ctx := context.Background()
	got := RequestID(ctx)
	if got != "" {
		t.Errorf("expected empty request ID, got %q", got)
	}
}

func TestLoggerWithNoContextFields(t *testing.T) {
	ctx := context.Background()

	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: stripTimeAndLevel,
	})
	slog.SetDefault(slog.New(handler))

	Logger(ctx).Info("bare message")

	out := buf.String()
	if !strings.Contains(out, "bare message") {
		t.Errorf("expected 'bare message' in output, got: %s", out)
	}
}

func TestFromEnrichesProvidedLogger(t *testing.T) {
	ctx := context.Background()
	ctx = With(ctx, "env", "prod")
	ctx = WithRequestID(ctx, "req-999")

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: stripTimeAndLevel,
	}))

	From(ctx, logger).Info("enriched")

	out := buf.String()
	if !strings.Contains(out, "env=prod") {
		t.Errorf("expected env=prod in output, got: %s", out)
	}
	if !strings.Contains(out, "request_id=req-999") {
		t.Errorf("expected request_id=req-999 in output, got: %s", out)
	}
}

func TestMultipleWithCallsAccumulateFields(t *testing.T) {
	ctx := context.Background()
	ctx = With(ctx, "a", 1)
	ctx = With(ctx, "b", 2)
	ctx = With(ctx, "c", 3)

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: stripTimeAndLevel,
	}))
	From(ctx, logger).Info("test")

	out := buf.String()
	for _, key := range []string{"a=1", "b=2", "c=3"} {
		if !strings.Contains(out, key) {
			t.Errorf("expected %s in output, got: %s", key, out)
		}
	}
}

func TestMultipleWithAttrsCallsAccumulate(t *testing.T) {
	ctx := context.Background()
	ctx = WithAttrs(ctx, slog.String("x", "1"))
	ctx = WithAttrs(ctx, slog.String("y", "2"))

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: stripTimeAndLevel,
	}))
	From(ctx, logger).Info("test")

	out := buf.String()
	if !strings.Contains(out, "x=1") {
		t.Errorf("expected x=1 in output, got: %s", out)
	}
	if !strings.Contains(out, "y=2") {
		t.Errorf("expected y=2 in output, got: %s", out)
	}
}

func TestLoggerIncludesRequestID(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "rid-abc")

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: stripTimeAndLevel,
	}))
	From(ctx, logger).Info("test")

	out := buf.String()
	if !strings.Contains(out, "request_id=rid-abc") {
		t.Errorf("expected request_id=rid-abc in output, got: %s", out)
	}
}

func TestWithCorrelationIDRoundTrip(t *testing.T) {
	ctx := context.Background()
	id := "corr-abc-123"
	ctx = WithCorrelationID(ctx, id)

	got := CorrelationID(ctx)
	if got != id {
		t.Errorf("expected correlation ID %q, got %q", id, got)
	}
}

func TestCorrelationIDReturnsEmptyWhenNotSet(t *testing.T) {
	ctx := context.Background()
	got := CorrelationID(ctx)
	if got != "" {
		t.Errorf("expected empty correlation ID, got %q", got)
	}
}

func TestWithTraceIDRoundTrip(t *testing.T) {
	ctx := context.Background()
	id := "trace-xyz-789"
	ctx = WithTraceID(ctx, id)

	got := TraceID(ctx)
	if got != id {
		t.Errorf("expected trace ID %q, got %q", id, got)
	}
}

func TestTraceIDReturnsEmptyWhenNotSet(t *testing.T) {
	ctx := context.Background()
	got := TraceID(ctx)
	if got != "" {
		t.Errorf("expected empty trace ID, got %q", got)
	}
}

func TestFromIncludesCorrelationID(t *testing.T) {
	ctx := context.Background()
	ctx = WithCorrelationID(ctx, "corr-001")

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: stripTimeAndLevel,
	}))
	From(ctx, logger).Info("test")

	out := buf.String()
	if !strings.Contains(out, "correlation_id=corr-001") {
		t.Errorf("expected correlation_id=corr-001 in output, got: %s", out)
	}
}

func TestFromIncludesTraceID(t *testing.T) {
	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-002")

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: stripTimeAndLevel,
	}))
	From(ctx, logger).Info("test")

	out := buf.String()
	if !strings.Contains(out, "trace_id=trace-002") {
		t.Errorf("expected trace_id=trace-002 in output, got: %s", out)
	}
}

func TestFieldsReturnsNilWhenEmpty(t *testing.T) {
	ctx := context.Background()
	got := Fields(ctx)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestFieldsReturnsAttachedAttrs(t *testing.T) {
	ctx := context.Background()
	ctx = WithAttrs(ctx, slog.String("k1", "v1"), slog.Int("k2", 99))

	got := Fields(ctx)
	if len(got) != 2 {
		t.Fatalf("expected 2 attrs, got %d", len(got))
	}
	if got[0].Key != "k1" || got[0].Value.String() != "v1" {
		t.Errorf("unexpected first attr: %v", got[0])
	}
	if got[1].Key != "k2" || got[1].Value.String() != "99" {
		t.Errorf("unexpected second attr: %v", got[1])
	}
}

func TestFieldsReturnsCopy(t *testing.T) {
	ctx := context.Background()
	ctx = WithAttrs(ctx, slog.String("k", "v"))

	got := Fields(ctx)
	got[0] = slog.String("modified", "bad")

	original := Fields(ctx)
	if original[0].Key != "k" {
		t.Error("Fields returned a reference instead of a copy")
	}
}

func TestContextChainingAllIDs(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-1")
	ctx = WithCorrelationID(ctx, "corr-1")
	ctx = WithTraceID(ctx, "trace-1")
	ctx = With(ctx, "env", "test")
	ctx = WithAttrs(ctx, slog.String("service", "api"))

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: stripTimeAndLevel,
	}))
	From(ctx, logger).Info("chained")

	out := buf.String()
	for _, want := range []string{
		"request_id=req-1",
		"correlation_id=corr-1",
		"trace_id=trace-1",
		"env=test",
		"service=api",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %s in output, got: %s", want, out)
		}
	}
}

// stripTimeAndLevel removes time and level attrs to simplify test output.
func stripTimeAndLevel(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey || a.Key == slog.LevelKey {
		return slog.Attr{}
	}
	return a
}
