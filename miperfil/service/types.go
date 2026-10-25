package service

import (
	"PlantillaInstitucional/miperfil"
	"context"
	"io"

	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/archivos"
	"github.com/issemym-repos/logs"
)

// RepositoryProvider define el puerto de persistencia requerido por el servicio de Mi Perfil.
type RepositoryProvider interface {
	ObtenerPerfilUsuario(ctx context.Context, uuidPersona uuid.UUID) (miperfil.PerfilUsuario, error)
	ObtenerPasswordHash(ctx context.Context, uuidPersona uuid.UUID) (string, error)
	ActualizarPassword(ctx context.Context, uuidPersona uuid.UUID, passwordHash string) error
	ActualizarImagenPerfil(ctx context.Context, uuidPersona uuid.UUID, nombreArchivo string) error
}

// ServiceProvider define el contrato de lógica de negocio para Mi Perfil.
type ServiceProvider interface {
	ObtenerPerfil(ctx context.Context, uuidPersona uuid.UUID) (miperfil.PerfilUsuario, error)
	CambiarPassword(ctx context.Context, uuidPersona uuid.UUID, form miperfil.CambioPasswordForm) error
	CambiarImagen(ctx context.Context, uuidPersona uuid.UUID, filename string, contentType string, reader io.Reader, size int64) (string, error)
}

// MiPerfilService implementa ServiceProvider.
type MiPerfilService struct {
	repo     RepositoryProvider
	archivos *archivos.ManejadorArchivos
	logger   *logs.Logger
}

// NewMiPerfilService construye la instancia del servicio.
func NewMiPerfilService(repo RepositoryProvider, archivos *archivos.ManejadorArchivos, logger *logs.Logger) *MiPerfilService {
	return &MiPerfilService{
		repo:     repo,
		archivos: archivos,
		logger:   logger,
	}
}
