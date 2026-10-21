package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegistrarHandler404(t *testing.T) {
	r := chi.NewRouter()
	RegistrarHandler404(r)

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/ruta-inexistente")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", res.StatusCode)
	}

	// Verify it renders the branded page (contains "no encontrada" or similar)
	// Note: We might need to handle the template rendering in tests if we want to check content
}
