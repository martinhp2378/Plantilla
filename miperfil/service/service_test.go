package service

import (
	"PlantillaInstitucional/miperfil"
	"PlantillaInstitucional/shared/password"
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/gofrs/uuid/v5"
)

type mockRepo struct {
	perfil       miperfil.PerfilUsuario
	passwordHash string
	updatedPass  string
	updatedImg   string
	err          error
}

func (m *mockRepo) ObtenerPerfilUsuario(ctx context.Context, uuidPersona uuid.UUID) (miperfil.PerfilUsuario, error) {
	if m.err != nil {
		return miperfil.PerfilUsuario{}, m.err
	}
	return m.perfil, nil
}

func (m *mockRepo) ObtenerPasswordHash(ctx context.Context, uuidPersona uuid.UUID) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.passwordHash, nil
}

func (m *mockRepo) ActualizarPassword(ctx context.Context, uuidPersona uuid.UUID, passwordHash string) error {
	if m.err != nil {
		return m.err
	}
	m.updatedPass = passwordHash
	return nil
}

func (m *mockRepo) ActualizarImagenPerfil(ctx context.Context, uuidPersona uuid.UUID, nombreArchivo string) error {
	if m.err != nil {
		return m.err
	}
	m.updatedImg = nombreArchivo
	return nil
}

func TestObtenerPerfil(t *testing.T) {
	uID := uuid.Must(uuid.NewV4())
	repo := &mockRepo{
		perfil: miperfil.PerfilUsuario{
			UuidPersona:  uID,
			Nombre:       "Rosendo",
			Apellido1:    "Mercado",
			Email:        "rosendo@issemym.gob.mx",
			Perfil:       "superadmin",
			ImagenPerfil: "placeholder.png",
		},
	}
	svc := NewMiPerfilService(repo, nil, nil)

	ctx := context.Background()
	p, err := svc.ObtenerPerfil(ctx, uID)
	if err != nil {
		t.Fatalf("ObtenerPerfil failed: %v", err)
	}
	if p.Nombre != "Rosendo" || p.NombreCompleto() != "Rosendo Mercado" {
		t.Errorf("Unexpected perfil data: %+v", p)
	}

	_, errNil := svc.ObtenerPerfil(ctx, uuid.Nil)
	if errNil == nil {
		t.Errorf("Expected error for nil UUID, got nil")
	}
}

func TestCambiarPassword(t *testing.T) {
	uID := uuid.Must(uuid.NewV4())
	currentPass := "OriginalPassword123!"
	currentHash, _ := password.HashPassword(currentPass)

	t.Run("successful password update", func(t *testing.T) {
		repo := &mockRepo{passwordHash: currentHash}
		svc := NewMiPerfilService(repo, nil, nil)

		err := svc.CambiarPassword(context.Background(), uID, miperfil.CambioPasswordForm{
			PasswordActual:       currentPass,
			PasswordNueva:        "NewSecurePassword456!",
			PasswordConfirmacion: "NewSecurePassword456!",
		})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if repo.updatedPass == "" || !password.CheckPasswordHash("NewSecurePassword456!", repo.updatedPass) {
			t.Errorf("Password was not properly hashed and updated")
		}
	})

	t.Run("wrong current password", func(t *testing.T) {
		repo := &mockRepo{passwordHash: currentHash}
		svc := NewMiPerfilService(repo, nil, nil)

		err := svc.CambiarPassword(context.Background(), uID, miperfil.CambioPasswordForm{
			PasswordActual:       "IncorrectPassword!",
			PasswordNueva:        "NewSecurePassword456!",
			PasswordConfirmacion: "NewSecurePassword456!",
		})
		if !errors.Is(err, ErrPasswordActualIncorrecta) {
			t.Errorf("Expected ErrPasswordActualIncorrecta, got: %v", err)
		}
	})

	t.Run("new password too short", func(t *testing.T) {
		repo := &mockRepo{passwordHash: currentHash}
		svc := NewMiPerfilService(repo, nil, nil)

		err := svc.CambiarPassword(context.Background(), uID, miperfil.CambioPasswordForm{
			PasswordActual:       currentPass,
			PasswordNueva:        "short",
			PasswordConfirmacion: "short",
		})
		if !errors.Is(err, ErrPasswordCorta) {
			t.Errorf("Expected ErrPasswordCorta, got: %v", err)
		}
	})

	t.Run("confirmation mismatch", func(t *testing.T) {
		repo := &mockRepo{passwordHash: currentHash}
		svc := NewMiPerfilService(repo, nil, nil)

		err := svc.CambiarPassword(context.Background(), uID, miperfil.CambioPasswordForm{
			PasswordActual:       currentPass,
			PasswordNueva:        "NewSecurePassword456!",
			PasswordConfirmacion: "DifferentPassword456!",
		})
		if !errors.Is(err, ErrPasswordNoCoincide) {
			t.Errorf("Expected ErrPasswordNoCoincide, got: %v", err)
		}
	})

	t.Run("new password equals current", func(t *testing.T) {
		repo := &mockRepo{passwordHash: currentHash}
		svc := NewMiPerfilService(repo, nil, nil)

		err := svc.CambiarPassword(context.Background(), uID, miperfil.CambioPasswordForm{
			PasswordActual:       currentPass,
			PasswordNueva:        currentPass,
			PasswordConfirmacion: currentPass,
		})
		if !errors.Is(err, ErrPasswordIgual) {
			t.Errorf("Expected ErrPasswordIgual, got: %v", err)
		}
	})
}

func TestCambiarImagen(t *testing.T) {
	t.Cleanup(func() {
		_ = os.RemoveAll("static")
	})
	uID := uuid.Must(uuid.NewV4())

	t.Run("successful image update with png", func(t *testing.T) {
		repo := &mockRepo{}
		svc := NewMiPerfilService(repo, nil, nil)

		fakePNG := []byte("\x89PNG\r\n\x1a\nfakeimagecontent")
		name, err := svc.CambiarImagen(context.Background(), uID, "avatar.png", "image/png", bytes.NewReader(fakePNG), int64(len(fakePNG)))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !strings.HasSuffix(name, ".png") || len(name) > 40 {
			t.Errorf("Generated filename unexpected or exceeds 40 chars: %s (len: %d)", name, len(name))
		}
		if repo.updatedImg != name {
			t.Errorf("Repository was not updated with filename %s", name)
		}
	})

	t.Run("reject invalid extension", func(t *testing.T) {
		repo := &mockRepo{}
		svc := NewMiPerfilService(repo, nil, nil)

		content := []byte("some executable or text")
		_, err := svc.CambiarImagen(context.Background(), uID, "script.sh", "application/x-sh", bytes.NewReader(content), int64(len(content)))
		if !errors.Is(err, ErrFormatoImagenInvalido) {
			t.Errorf("Expected ErrFormatoImagenInvalido, got: %v", err)
		}
	})

	t.Run("reject oversized file", func(t *testing.T) {
		repo := &mockRepo{}
		svc := NewMiPerfilService(repo, nil, nil)

		_, err := svc.CambiarImagen(context.Background(), uID, "photo.jpg", "image/jpeg", bytes.NewReader([]byte{}), MaxImageSizeBytes+1)
		if !errors.Is(err, ErrImagenDemasiadoGrande) {
			t.Errorf("Expected ErrImagenDemasiadoGrande, got: %v", err)
		}
	})
}
