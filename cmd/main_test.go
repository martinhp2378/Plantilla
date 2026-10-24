package main

import (
	"PlantillaInstitucional/institucion"
	"PlantillaInstitucional/shared/sesionutil"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/sesiones"
)

func TestIsMigrationCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "starts the HTTP server without a command", args: nil, want: false},
		{name: "runs migrations with the dedicated command", args: []string{"migrate"}, want: true},
		{name: "does not treat extra arguments as a migration command", args: []string{"migrate", "extra"}, want: false},
		{name: "does not treat another command as a migration command", args: []string{"serve"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMigrationCommand(tt.args); got != tt.want {
				t.Errorf("isMigrationCommand(%q) = %t, want %t", tt.args, got, tt.want)
			}
		})
	}
}

func TestNewHTTPServerTimeouts(t *testing.T) {
	handler := http.NewServeMux()
	srv := newHTTPServer(":8080", handler)

	if srv.Addr != ":8080" {
		t.Fatalf("Addr = %q, want %q", srv.Addr, ":8080")
	}
	if srv.ReadHeaderTimeout != 5*time.Second {
		t.Errorf("ReadHeaderTimeout = %v, want %v", srv.ReadHeaderTimeout, 5*time.Second)
	}
	if srv.ReadTimeout != 15*time.Second {
		t.Errorf("ReadTimeout = %v, want %v", srv.ReadTimeout, 15*time.Second)
	}
	if srv.WriteTimeout != 30*time.Second {
		t.Errorf("WriteTimeout = %v, want %v", srv.WriteTimeout, 30*time.Second)
	}
	if srv.IdleTimeout != 60*time.Second {
		t.Errorf("IdleTimeout = %v, want %v", srv.IdleTimeout, 60*time.Second)
	}
	if srv.MaxHeaderBytes != 1<<20 {
		t.Errorf("MaxHeaderBytes = %d, want %d", srv.MaxHeaderBytes, 1<<20)
	}
}

func TestHealthRoutes(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		ping       func(context.Context) error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "liveness probe returns 200 ok",
			path:       "/healthz",
			ping:       nil,
			wantStatus: http.StatusOK,
			wantBody:   `{"status":"ok"}`,
		},
		{
			name: "readiness probe returns 200 ready when database responds",
			path: "/readyz",
			ping: func(context.Context) error {
				return nil
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"status":"ready"}`,
		},
		{
			name: "readiness probe returns 503 when database ping fails",
			path: "/readyz",
			ping: func(context.Context) error {
				return errors.New("db down")
			},
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   `{"status":"unavailable"}`,
		},
		{
			name:       "readiness probe returns 503 when ping function is nil",
			path:       "/readyz",
			ping:       nil,
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   `{"status":"unavailable"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			registerHealthRoutes(r, tt.ping)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if rec.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestBuildSessionUserLoader(t *testing.T) {
	personID := uuid.Must(uuid.FromString("11111111-1111-1111-1111-111111111111"))

	tests := []struct {
		name       string
		lookup     func(context.Context, uuid.UUID) (institucion.Persona, error)
		perfil     sesiones.Perfil
		datosExtra map[string]any
		want       sesionutil.Usuario
	}{
		{
			name: "hydrates user session from persona repository and extra metadata",
			lookup: func(_ context.Context, id uuid.UUID) (institucion.Persona, error) {
				return institucion.Persona{
					UuidPersona:  id,
					Nombre:       "Ana",
					Apellido1:    "Pérez",
					Apellido2:    "Gómez",
					Curp:         "PEGA900101HDFRNN01",
					ClaveIssemym: "123456",
					ImagenPerfil: "avatar.png",
				}, nil
			},
			perfil: sesiones.PerfilSuperAdmin,
			datosExtra: map[string]any{
				"email":                              "ana@issemym.gob.mx",
				sesionutil.DatoExtraPerfilAplicacion: "admin",
				sesionutil.DatoExtraTenantSchema:     "sistema",
			},
			want: sesionutil.Usuario{
				UuidPersona:         personID.String(),
				Nombre:              "Ana",
				Apellido1:           "Pérez",
				Apellido2:           "Gómez",
				Curp:                "PEGA900101HDFRNN01",
				ClaveIssemym:        "123456",
				ArchivoImagenPerfil: "avatar.png",
				Email:               "ana@issemym.gob.mx",
				Perfil:              "admin",
				TenantSchema:        "sistema",
			},
		},
		{
			name: "falls back safely when lookup returns an error",
			lookup: func(_ context.Context, _ uuid.UUID) (institucion.Persona, error) {
				return institucion.Persona{}, errors.New("connection refused")
			},
			perfil:     sesiones.PerfilSuperAdmin,
			datosExtra: nil,
			want: sesionutil.Usuario{
				UuidPersona:         personID.String(),
				ArchivoImagenPerfil: sesionutil.SinImagenPerfil,
				Perfil:              sesionutil.PerfilSuperAdmin,
				TenantSchema:        "public",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := buildSessionUserLoader(tt.lookup, nil)
			var result []any
			if tt.datosExtra != nil {
				result = loader(personID, tt.perfil, tt.datosExtra)
			} else {
				result = loader(personID, tt.perfil)
			}

			if len(result) != 1 {
				t.Fatalf("len(result) = %d, want 1", len(result))
			}
			got, ok := result[0].(sesionutil.Usuario)
			if !ok {
				t.Fatalf("result[0] type = %T, want sesionutil.Usuario", result[0])
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
