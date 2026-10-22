package handler

import (
	"PlantillaInstitucional/miperfil"
	"PlantillaInstitucional/shared/sesionutil"
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/logs"
	"github.com/issemym-repos/sesiones"
)

type mockService struct {
	perfil    miperfil.PerfilUsuario
	newPassErr error
	newImgErr  error
	newImgName string
}

func (m *mockService) ObtenerPerfil(ctx context.Context, uuidPersona uuid.UUID) (miperfil.PerfilUsuario, error) {
	return m.perfil, nil
}

func (m *mockService) CambiarPassword(ctx context.Context, uuidPersona uuid.UUID, form miperfil.CambioPasswordForm) error {
	return m.newPassErr
}

func (m *mockService) CambiarImagen(ctx context.Context, uuidPersona uuid.UUID, filename string, contentType string, reader io.Reader, size int64) (string, error) {
	if m.newImgErr != nil {
		return "", m.newImgErr
	}
	return m.newImgName, nil
}

func setupTestSession(req *http.Request, uID uuid.UUID) (*http.Request, *sesiones.ConstructorSesion) {
	reg := sesiones.NuevoRegistroTipos()
	_, _ = reg.RegistrarTipo(sesionutil.Usuario{})
	cs := sesiones.NuevoConstructor(nil, &sesiones.ConfigSesion{}, reg, nil)
	ses := cs.NuevaSesion(req)
	ses.Perfil = "superadmin"
	ses.UuidPersona = &uID
	_ = cs.AgregarInstancia(ses, sesionutil.Usuario{
		UuidPersona:  uID.String(),
		Nombre:       "Rosendo",
		Perfil:       "superadmin",
		TenantSchema: "local",
	})
	ctx := context.WithValue(req.Context(), sesiones.LlaveContexto, ses)
	return req.WithContext(ctx), cs
}

func TestHandleVerPerfil(t *testing.T) {
	uID := uuid.Must(uuid.NewV4())
	logger := logs.NuevoLogger(&logs.Config{Nivel: "DEBUG"})
	svc := &mockService{
		perfil: miperfil.PerfilUsuario{
			UuidPersona:  uID,
			Nombre:       "Rosendo",
			Email:        "rosendo@issemym.gob.mx",
			Perfil:       "superadmin",
			ImagenPerfil: "placeholder.png",
		},
	}
	h := NewMiPerfilHandler(svc, logger, nil)

	t.Run("redirect when unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/perfil", nil)
		rr := httptest.NewRecorder()
		h.HandleVerPerfil(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Errorf("Expected 303 redirect, got: %d", rr.Code)
		}
	})

	t.Run("success when authenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/perfil", nil)
		req, _ = setupTestSession(req, uID)
		rr := httptest.NewRecorder()
		h.HandleVerPerfil(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got: %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Mi Perfil") {
			t.Errorf("Body does not contain 'Mi Perfil'")
		}
	})
}

func TestHandleCambiarPassword(t *testing.T) {
	uID := uuid.Must(uuid.NewV4())
	logger := logs.NuevoLogger(&logs.Config{Nivel: "DEBUG"})

	t.Run("success feedback", func(t *testing.T) {
		svc := &mockService{}
		h := NewMiPerfilHandler(svc, logger, nil)

		form := url.Values{
			"password_actual":       {"CurrentPass123!"},
			"password_nueva":        {"NewSecurePass456!"},
			"password_confirmacion": {"NewSecurePass456!"},
		}
		req := httptest.NewRequest(http.MethodPost, "/perfil/password", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req, _ = setupTestSession(req, uID)

		rr := httptest.NewRecorder()
		h.HandleCambiarPassword(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got: %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "actualizada exitosamente") {
			t.Errorf("Expected success message, got: %s", rr.Body.String())
		}
	})

	t.Run("error feedback when service rejects", func(t *testing.T) {
		svc := &mockService{newPassErr: errors.New("la contraseña actual no es correcta")}
		h := NewMiPerfilHandler(svc, logger, nil)

		form := url.Values{
			"password_actual":       {"WrongPass!"},
			"password_nueva":        {"NewPass456!"},
			"password_confirmacion": {"NewPass456!"},
		}
		req := httptest.NewRequest(http.MethodPost, "/perfil/password", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req, _ = setupTestSession(req, uID)

		rr := httptest.NewRecorder()
		h.HandleCambiarPassword(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected 200 OK for HTMX error feedback, got: %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "la contraseña actual no es correcta") {
			t.Errorf("Expected error message in feedback, got: %s", rr.Body.String())
		}
	})
}

func TestHandleCambiarImagen(t *testing.T) {
	uID := uuid.Must(uuid.NewV4())
	logger := logs.NuevoLogger(&logs.Config{Nivel: "DEBUG"})

	t.Run("success upload image", func(t *testing.T) {
		svc := &mockService{newImgName: uID.String() + "_123456.png"}
		h := NewMiPerfilHandler(svc, logger, nil)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("imagen", "avatar.png")
		_, _ = part.Write([]byte("fake-image-bytes"))
		_ = writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/perfil/imagen", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req, _ = setupTestSession(req, uID)

		rr := httptest.NewRecorder()
		h.HandleCambiarImagen(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got: %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Foto de perfil actualizada exitosamente") {
			t.Errorf("Expected success message, got: %s", rr.Body.String())
		}
	})

	t.Run("missing file error feedback", func(t *testing.T) {
		svc := &mockService{}
		h := NewMiPerfilHandler(svc, logger, nil)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/perfil/imagen", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req, _ = setupTestSession(req, uID)

		rr := httptest.NewRecorder()
		h.HandleCambiarImagen(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got: %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Debe seleccionar un archivo") {
			t.Errorf("Expected error feedback, got: %s", rr.Body.String())
		}
	})
}
