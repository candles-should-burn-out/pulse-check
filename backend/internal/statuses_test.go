package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatusRoutesCreateUpdateDelete(t *testing.T) {
	app := NewApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	handler := app.Routes()

	getRecorder := httptest.NewRecorder()
	handler.ServeHTTP(getRecorder, httptest.NewRequest(http.MethodGet, "/status-set", nil))
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("GET /status-set status = %d, want %d", getRecorder.Code, http.StatusOK)
	}

	var statusSet StatusSet
	if err := json.NewDecoder(getRecorder.Body).Decode(&statusSet); err != nil {
		t.Fatalf("decode status set: %v", err)
	}
	if statusSet.Role != StatusRoleOwner {
		t.Fatalf("role = %q, want %q", statusSet.Role, StatusRoleOwner)
	}

	created := requestStatus(t, handler, http.MethodPost, "/statuses", StatusInput{
		Name:            "In progress",
		BorderColor:     "#5e81ac",
		BackgroundColor: "#e8f0fb",
		TextColor:       "#24344d",
	}, http.StatusCreated)
	if created.ID == "" {
		t.Fatal("created status ID is empty")
	}

	updated := requestStatus(t, handler, http.MethodPatch, "/statuses/"+created.ID, StatusInput{
		Name:            "Done",
		BorderColor:     "#4f8f5f",
		BackgroundColor: "#e7f4ea",
		TextColor:       "#1f3f29",
	}, http.StatusOK)
	if updated.ID != created.ID {
		t.Fatalf("updated ID = %q, want %q", updated.ID, created.ID)
	}
	if updated.Name != "Done" {
		t.Fatalf("updated name = %q, want Done", updated.Name)
	}

	deleteRecorder := httptest.NewRecorder()
	handler.ServeHTTP(deleteRecorder, httptest.NewRequest(http.MethodDelete, "/statuses/"+created.ID, nil))
	if deleteRecorder.Code != http.StatusNoContent {
		t.Fatalf("DELETE /statuses/{id} status = %d, want %d", deleteRecorder.Code, http.StatusNoContent)
	}
}

func TestStatusLimitReturnsConflict(t *testing.T) {
	app := NewApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	app.SetStatusStore(NewMemoryStatusStore(1))
	handler := app.Routes()

	requestStatus(t, handler, http.MethodPost, "/statuses", StatusInput{
		Name:            "One",
		BorderColor:     "#5e81ac",
		BackgroundColor: "#e8f0fb",
		TextColor:       "#24344d",
	}, http.StatusCreated)

	recorder := requestJSON(t, handler, http.MethodPost, "/statuses", StatusInput{
		Name:            "Two",
		BorderColor:     "#4f8f5f",
		BackgroundColor: "#e7f4ea",
		TextColor:       "#1f3f29",
	})
	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusConflict)
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte("status_limit_reached")) {
		t.Fatalf("body = %q, want status_limit_reached", recorder.Body.String())
	}
}

func TestStatusNameMaxLength(t *testing.T) {
	app := NewApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	handler := app.Routes()

	requestStatus(t, handler, http.MethodPost, "/statuses", StatusInput{
		Name:            "1234567890123456789012345678901234567890",
		BorderColor:     "#5e81ac",
		BackgroundColor: "#e8f0fb",
		TextColor:       "#24344d",
	}, http.StatusCreated)

	recorder := requestJSON(t, handler, http.MethodPost, "/statuses", StatusInput{
		Name:            "12345678901234567890123456789012345678901",
		BorderColor:     "#5e81ac",
		BackgroundColor: "#e8f0fb",
		TextColor:       "#24344d",
	})
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte("invalid_status")) {
		t.Fatalf("body = %q, want invalid_status", recorder.Body.String())
	}
}

func TestParticipantCannotModifyStatuses(t *testing.T) {
	store := NewMemoryStatusStore(DefaultStatusLimit).(*memoryStatusStore)
	ownerSet, err := store.GetStatusSet(context.Background(), "owner")
	if err != nil {
		t.Fatalf("create owner set: %v", err)
	}

	store.mu.Lock()
	store.memberships["participant"] = statusSetMembership{
		UserID:      "participant",
		StatusSetID: ownerSet.ID,
		OwnerUserID: "owner",
	}
	store.mu.Unlock()

	app := NewApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	app.SetStatusStore(store)
	handler := app.Routes()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/statuses", encodeJSON(t, StatusInput{
		Name:            "Participant status",
		BorderColor:     "#5e81ac",
		BackgroundColor: "#e8f0fb",
		TextColor:       "#24344d",
	}))
	request = request.WithContext(context.WithValue(request.Context(), authContextKey, &TokenClaims{Subject: "participant"}))
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func requestStatus(t *testing.T, handler http.Handler, method string, path string, input StatusInput, expectedStatus int) Status {
	t.Helper()

	recorder := requestJSON(t, handler, method, path, input)
	if recorder.Code != expectedStatus {
		t.Fatalf("%s %s status = %d, want %d; body = %q", method, path, recorder.Code, expectedStatus, recorder.Body.String())
	}

	var status Status
	if err := json.NewDecoder(recorder.Body).Decode(&status); err != nil {
		t.Fatalf("decode status: %v", err)
	}

	return status
}

func requestJSON(t *testing.T, handler http.Handler, method string, path string, payload any) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, encodeJSON(t, payload))
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(recorder, request)

	return recorder
}

func encodeJSON(t *testing.T, payload any) io.Reader {
	t.Helper()

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		t.Fatalf("encode json: %v", err)
	}

	return &body
}
