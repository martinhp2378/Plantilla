package entorno

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func TestGetEnvEarlyValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		env       map[string]string
		wantError string
	}{
		{
			name: "invalid environment",
			env: map[string]string{
				"AMBIENTE": "OTRO",
			},
			wantError: "ambiente",
		},
		{
			name: "missing servicio",
			env: map[string]string{
				"AMBIENTE": AmbienteLocal,
			},
			wantError: "SERVICIO",
		},
		{
			name: "missing project id",
			env: map[string]string{
				"AMBIENTE":    AmbienteLocal,
				"SERVICIO":    "encuestas",
				"ID_PROYECTO": "",
			},
			wantError: "ID_PROYECTO",
		},
		{
			name: "missing version",
			env: map[string]string{
				"AMBIENTE":    AmbienteLocal,
				"SERVICIO":    "encuestas",
				"ID_PROYECTO": "issemym",
			},
			wantError: "VERSION",
		},
	}

	keysToClear := []string{"AMBIENTE", "SERVICIO", "ID_PROYECTO", "VERSION"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, key := range keysToClear {
				t.Setenv(key, "")
			}
			for key, value := range tt.env {
				t.Setenv(key, value)
			}

			env, err := GetEnv()
			if err == nil {
				t.Fatalf("expected error, got env %#v", env)
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.wantError)) {
				t.Fatalf("expected error containing %q, got %q", tt.wantError, err.Error())
			}
		})
	}
}

func TestEnvMiddlewareAndHelpers(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantNil bool
		wantAmb string
	}{
		{
			name:    "missing env returns nil",
			ctx:     context.Background(),
			wantNil: true,
		},
		{
			name:    "wrong type returns nil",
			ctx:     context.WithValue(context.Background(), envKey, "not-an-env"),
			wantNil: true,
		},
		{
			name:    "stored env is recovered",
			ctx:     context.WithValue(context.Background(), envKey, &Env{Ambiente: AmbienteDesarrollo}),
			wantAmb: AmbienteDesarrollo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EnvDeContexto(tt.ctx)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected nil env, got %#v", got)
				}
				if amb := ObtenerAmbiente(tt.ctx); amb != "" {
					t.Fatalf("expected empty ambiente, got %q", amb)
				}
				return
			}

			if got == nil {
				t.Fatal("expected env in context, got nil")
			}
			if got.Ambiente != tt.wantAmb {
				t.Fatalf("expected ambiente %q, got %q", tt.wantAmb, got.Ambiente)
			}
			if amb := ObtenerAmbiente(tt.ctx); amb != tt.wantAmb {
				t.Fatalf("expected ambiente helper to return %q, got %q", tt.wantAmb, amb)
			}
		})
	}

	t.Run("middleware injects env into request context", func(t *testing.T) {
		expectedEnv := &Env{Ambiente: AmbienteProduccion}
		var got *Env

		handler := EnvMiddleware(expectedEnv)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got = EnvDeContexto(r.Context())
			if amb := ObtenerAmbiente(r.Context()); amb != AmbienteProduccion {
				t.Fatalf("expected ambiente %q, got %q", AmbienteProduccion, amb)
			}
			w.WriteHeader(http.StatusNoContent)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
		}
		if got != expectedEnv {
			t.Fatalf("expected same env pointer injected in context")
		}
	})
}

func TestHelpersReadEnvironmentValues(t *testing.T) {
	t.Run("getEnvODefault returns env value when set", func(t *testing.T) {
		t.Setenv("ARCHIVOS_DIR_LOCAL", "/tmp/custom")

		if got := getEnvODefault("ARCHIVOS_DIR_LOCAL", "./static"); got != "/tmp/custom" {
			t.Fatalf("expected env value, got %q", got)
		}
	})

	t.Run("getEnvODefault returns default when missing", func(t *testing.T) {
		t.Setenv("ARCHIVOS_DIR_LOCAL", "")

		if got := getEnvODefault("ARCHIVOS_DIR_LOCAL", "./static"); got != "./static" {
			t.Fatalf("expected default value, got %q", got)
		}
	})

	t.Run("parseTimeEnv returns parsed duration", func(t *testing.T) {
		t.Setenv("ARCHIVOS_EXPIRACION_URLS_TEST", "45m")

		if got := parseTimeEnv("ARCHIVOS_EXPIRACION_URLS_TEST", 5*time.Minute); got != 45*time.Minute {
			t.Fatalf("expected parsed duration, got %s", got)
		}
	})

	t.Run("parseTimeEnv returns default for invalid duration", func(t *testing.T) {
		t.Setenv("ARCHIVOS_EXPIRACION_URLS_TEST", "nope")

		if got := parseTimeEnv("ARCHIVOS_EXPIRACION_URLS_TEST", 5*time.Minute); got != 5*time.Minute {
			t.Fatalf("expected default duration, got %s", got)
		}
	})
}

func TestObtenerRequestId(t *testing.T) {
	var got string

	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = ObtenerRequestId(r.Context())
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
	if got == "" {
		t.Fatal("expected request id generated by middleware")
	}
}
