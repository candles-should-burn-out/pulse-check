package internal

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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

