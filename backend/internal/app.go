package internal

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"pulse-check-backend/internal/utils"
)

const entityListRequestsMetric = "pulse_check_entity_list_requests_total"

type App struct {
	logger             *slog.Logger
	entityListRequests atomic.Uint64
	ready              atomic.Bool
	started            atomic.Bool
}

type Entity struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

func NewApp(logger *slog.Logger) *App {
	app := &App{logger: logger}
	app.ready.Store(true)
	app.started.Store(true)

	return app
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/entities", a.handleEntities)
	mux.HandleFunc("/metrics", a.handleMetrics)
	mux.HandleFunc("/swagger/", a.handleSwagger)

	mux.HandleFunc("/health/live", a.handleLive)
	mux.HandleFunc("/health/ready", a.handleReady)
	mux.HandleFunc("/health/startup", a.handleStartup)
	mux.HandleFunc("/livez", a.handleLive)
	mux.HandleFunc("/readyz", a.handleReady)
	mux.HandleFunc("/startupz", a.handleStartup)

	return otelhttp.NewHandler(
		utils.WithAccessLog(a.logger, mux),
		"http.server",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return r.Method + " " + r.URL.Path
		}),
	)
}

func (a *App) SetReady(ready bool) {
	a.ready.Store(ready)
}

func (a *App) handleEntities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	a.entityListRequests.Add(1)

	respondJSON(w, http.StatusOK, []Entity{
		{ID: "2b72d045-d9d7-43ce-9952-59a0f3e35e88", State: "active"},
		{ID: "fd57c2c6-3424-45cf-93b4-f5c9d984f611", State: "pending"},
		{ID: "61bb56a8-f758-424b-a82b-9885d8bea7b3", State: "disabled"},
	})
}

func (a *App) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	_, _ = fmt.Fprintf(w, "# HELP %s Total number of entity list requests.\n", entityListRequestsMetric)
	_, _ = fmt.Fprintf(w, "# TYPE %s counter\n", entityListRequestsMetric)
	_, _ = fmt.Fprintf(w, "%s %d\n", entityListRequestsMetric, a.entityListRequests.Load())
}

func (a *App) handleSwagger(w http.ResponseWriter, r *http.Request) { // TODO: Переделать на заполняемую из кода
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(utils.OpenAPISchema))
}

func (a *App) handleLive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) handleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	if !a.ready.Load() {
		respondJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) handleStartup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	if !a.started.Load() {
		respondJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "starting"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func methodNotAllowed(w http.ResponseWriter, allowedMethods ...string) {
	for _, method := range allowedMethods {
		w.Header().Add("Allow", method)
	}

	respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
}
