package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"pulse-check-backend/internal/utils"
)

const entityListRequestsMetric = "pulse_check_entity_list_requests_total"

type App struct {
	logger             *slog.Logger
	authenticator      *Authenticator
	authError          error
	entityListRequests atomic.Uint64
	statusStore        StatusStore
	ready              atomic.Bool
	started            atomic.Bool
}

type Entity struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

func NewApp(logger *slog.Logger, authConfig ...AuthConfig) *App {
	app := &App{
		logger:      logger,
		statusStore: NewMemoryStatusStore(DefaultStatusLimit),
	}
	authReady := true
	if len(authConfig) > 0 {
		authenticator, err := NewAuthenticator(authConfig[0])
		if err != nil && err != errAuthDisabled {
			logger.Error("auth configuration failed", slog.Any("error", err))
			app.authError = err
			authReady = false
		}
		app.authenticator = authenticator
	}
	app.ready.Store(authReady)
	app.started.Store(true)

	return app
}

func (a *App) SetStatusStore(statusStore StatusStore) {
	if statusStore != nil {
		a.statusStore = statusStore
	}
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/entities", a.protected(http.HandlerFunc(a.handleEntities)))
	mux.Handle("/status-set", a.protected(http.HandlerFunc(a.handleStatusSet)))
	mux.Handle("/statuses", a.protected(http.HandlerFunc(a.handleStatuses)))
	mux.Handle("/statuses/", a.protected(http.HandlerFunc(a.handleStatusByID)))
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

func (a *App) protected(next http.Handler) http.Handler {
	if a.authError != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "auth_not_configured"})
		})
	}
	if a.authenticator == nil {
		return next
	}

	return a.authenticator.Middleware(next)
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

func (a *App) handleStatusSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	statusSet, err := a.statusStore.GetStatusSet(r.Context(), currentUserID(r))
	if err != nil {
		a.logger.Error("get status set failed", slog.Any("error", err))
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "status_set_unavailable"})
		return
	}

	respondJSON(w, http.StatusOK, statusSet)
}

func (a *App) handleStatuses(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		statuses, err := a.statusStore.ListStatuses(r.Context(), currentUserID(r))
		if err != nil {
			a.logger.Error("list statuses failed", slog.Any("error", err))
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "statuses_unavailable"})
			return
		}

		respondJSON(w, http.StatusOK, statuses)
	case http.MethodPost:
		var input StatusInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
			return
		}

		status, err := a.statusStore.CreateStatus(r.Context(), currentUserID(r), input)
		if err != nil {
			respondStatusError(w, err)
			return
		}

		respondJSON(w, http.StatusCreated, status)
	default:
		methodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (a *App) handleStatusByID(w http.ResponseWriter, r *http.Request) {
	statusID := strings.TrimPrefix(r.URL.Path, "/statuses/")
	if statusID == "" || strings.Contains(statusID, "/") {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodPatch:
		var input StatusInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
			return
		}

		status, err := a.statusStore.UpdateStatus(r.Context(), currentUserID(r), statusID, input)
		if err != nil {
			respondStatusError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, status)
	case http.MethodDelete:
		if err := a.statusStore.DeleteStatus(r.Context(), currentUserID(r), statusID); err != nil {
			respondStatusError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	default:
		methodNotAllowed(w, http.MethodPatch, http.MethodDelete)
	}
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

func respondStatusError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidStatus):
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_status"})
	case errors.Is(err, ErrStatusForbidden):
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "status_set_read_only"})
	case errors.Is(err, ErrStatusLimit):
		respondJSON(w, http.StatusConflict, map[string]string{"error": "status_limit_reached"})
	case errors.Is(err, ErrStatusNotFound):
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "status_not_found"})
	default:
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "status_operation_failed"})
	}
}

func currentUserID(r *http.Request) string {
	if claims, ok := AuthClaims(r.Context()); ok && strings.TrimSpace(claims.Subject) != "" {
		return claims.Subject
	}

	return "local-user"
}
