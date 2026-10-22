package service

import (
	"PlantillaInstitucional/miperfil"
	"PlantillaInstitucional/miperfil/repository"
	"context"
	"io"

	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/archivos"
	"github.com/issemym-repos/logs"
)

// ServiceProvider define el contrato de lógica de negocio para Mi Perfil.
type ServiceProvider interface {
	ObtenerPerfil(ctx context.Context, uuidPersona uuid.UUID) (miperfil.PerfilUsuario, error)
	CambiarPassword(ctx context.Context, uuidPersona uuid.UUID, form miperfil.CambioPasswordForm) error
	CambiarImagen(ctx context.Context, uuidPersona uuid.UUID, filename string, contentType string, reader io.Reader, size int64) (string, error)
}

// MiPerfilService implementa ServiceProvider.
type MiPerfilService struct {
	repo     repository.RepositoryProvider
	archivos *archivos.ManejadorArchivos
	logger   *logs.Logger
}

// NewMiPerfilService construye la instancia del servicio.
func NewMiPerfilService(repo repository.RepositoryProvider, archivos *archivos.ManejadorArchivos, logger *logs.Logger) *MiPerfilService {
	return &MiPerfilService{
		repo:     repo,
		archivos: archivos,
		logger:   logger,
	}
}
