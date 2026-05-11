package internal

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestEntitiesReturnsHardcodedListAndIncrementsMetric(t *testing.T) {
	app := NewApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	handler := app.Routes()

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/entities", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var entities []Entity
	if err := json.NewDecoder(recorder.Body).Decode(&entities); err != nil {
		t.Fatalf("decode entities: %v", err)
	}

	if len(entities) != 3 {
		t.Fatalf("len(entities) = %d, want 3", len(entities))
	}

	metricsRecorder := httptest.NewRecorder()
	handler.ServeHTTP(metricsRecorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := metricsRecorder.Body.String()

	if !strings.Contains(body, entityListRequestsMetric+" 1\n") {
		t.Fatalf("metrics body does not contain incremented counter: %q", body)
	}
}

func TestTracingCreatesServerSpan(t *testing.T) {
	recorder := installSpanRecorder(t)
	app := NewApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	handler := app.Routes()

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/live", nil))

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("ended spans = %d, want 1", len(spans))
	}
	if spans[0].Name() != "GET /health/live" {
		t.Fatalf("span name = %q, want GET /health/live", spans[0].Name())
	}
	if !spans[0].SpanContext().TraceID().IsValid() {
		t.Fatalf("span trace id is invalid")
	}
}

func TestTracingContinuesIncomingTrace(t *testing.T) {
	recorder := installSpanRecorder(t)
	app := NewApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	handler := app.Routes()
	incomingTraceID := "0af7651916cd43dd8448eb211c80319c"
	incomingSpanID := "b7ad6b7169203331"
	incoming := "00-" + incomingTraceID + "-" + incomingSpanID + "-01"

	request := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	request.Header.Set("traceparent", incoming)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("ended spans = %d, want 1", len(spans))
	}
	if spans[0].SpanContext().TraceID().String() != incomingTraceID {
		t.Fatalf("trace id = %q, want %q", spans[0].SpanContext().TraceID().String(), incomingTraceID)
	}
	if spans[0].SpanContext().SpanID().String() == incomingSpanID {
		t.Fatalf("server span id was not renewed")
	}
}

func TestReadinessReflectsAppState(t *testing.T) {
	app := NewApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	handler := app.Routes()

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("ready status = %d, want %d", recorder.Code, http.StatusOK)
	}

	app.SetReady(false)

	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("not ready status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
}

func installSpanRecorder(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()

	recorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() {
		otel.SetTracerProvider(noop.NewTracerProvider())
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())
	})

	return recorder
}

func TestSwaggerReturnsOpenAPISchema(t *testing.T) { // TODO: Удалить кринжу
	app := NewApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	handler := app.Routes()

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/swagger/", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	if contentType := recorder.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Fatalf("content type = %q, want application/json", contentType)
	}

	var schema map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&schema); err != nil {
		t.Fatalf("decode openapi schema: %v", err)
	}

	if schema["openapi"] != "3.0.3" {
		t.Fatalf("openapi = %v, want 3.0.3", schema["openapi"])
	}

	paths, ok := schema["paths"].(map[string]any)
	if !ok {
		t.Fatalf("paths has unexpected type: %T", schema["paths"])
	}

	for _, path := range []string{"/entities", "/metrics", "/swagger/", "/health/live", "/health/ready", "/health/startup"} {
		if _, ok := paths[path]; !ok {
			t.Fatalf("schema does not describe %s", path)
		}
	}
}
