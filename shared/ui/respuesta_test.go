package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondWithJSON(t *testing.T) {
	recorder := httptest.NewRecorder()
	payload := map[string]any{
		"ok":      true,
		"mensaje": "listo",
	}

	RespondWithJSON(recorder, http.StatusCreated, payload)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("content-type = %q, want %q", contentType, "application/json")
	}

	var got map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if got["ok"] != true {
		t.Fatalf("ok = %#v, want true", got["ok"])
	}
	if got["mensaje"] != "listo" {
		t.Fatalf("mensaje = %#v, want %q", got["mensaje"], "listo")
	}
}

func TestRespondWithError(t *testing.T) {
	recorder := httptest.NewRecorder()

	RespondWithError(recorder, http.StatusBadRequest, "datos invalidos")

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("content-type = %q, want %q", contentType, "application/json")
	}

	var got map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if got["error"] != "datos invalidos" {
		t.Fatalf("error = %q, want %q", got["error"], "datos invalidos")
	}
}
