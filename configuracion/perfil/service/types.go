package service

import (
	"PlantillaInstitucional/configuracion/perfil"
	"PlantillaInstitucional/pagination"
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/issemym-repos/archivos"
	"github.com/issemym-repos/logs"
)

// PerfilRepositoryProvider define el contrato local con el repositorio de perfiles.
type PerfilRepositoryProvider interface {
	ConsultarPerfiles(ctx context.Context, criterios pagination.CriteriosPagina[perfil.Perfil]) ([]perfil.Perfil, int, error)
	ConsultarPerfilPorNombre(ctx context.Context, nombre string) (perfil.Perfil, error)
	CrearPerfil(ctx context.Context, p perfil.Perfil) error
	ActualizarPerfil(ctx context.Context, p perfil.Perfil, nombreOld string) error
	EliminarPerfil(ctx context.Context, nombre string) error
}

// PerfilService es la implementación del servicio de perfiles.
type PerfilService struct {
	repo     PerfilRepositoryProvider
	validate *validator.Validate
	logger   *logs.Logger
	archivos *archivos.ManejadorArchivos
}

// PerfilServiceProvider define el contrato del servicio de perfiles para el handler.
type PerfilServiceProvider interface {
	ConsultarPerfiles(ctx context.Context, criterios pagination.CriteriosPagina[perfil.Perfil]) (pagination.CriteriosPagina[perfil.Perfil], error)
	ConsultarPerfilPorNombre(ctx context.Context, nombre string) (perfil.Perfil, error)
	ObtenerPerfil(ctx context.Context, nombre string) (perfil.Perfil, error)
	ListarPerfiles(ctx context.Context, criterios pagination.CriteriosPagina[perfil.Perfil]) (pagination.CriteriosPagina[perfil.Perfil], error)
	CrearPerfil(ctx context.Context, p perfil.Perfil) error
	ActualizarPerfil(ctx context.Context, p perfil.Perfil, nombreOld string) error
	EliminarPerfil(ctx context.Context, nombre string) error
	ReportePdf(ctx context.Context) (string, error)
	ReporteExcel(ctx context.Context) (string, error)
}

// NewPerfilService inicializa el servicio.
func NewPerfilService(repo PerfilRepositoryProvider, validate *validator.Validate, logger *logs.Logger, archivos *archivos.ManejadorArchivos) *PerfilService {
	return &PerfilService{
		repo:     repo,
		validate: validate,
		logger:   logger,
		archivos: archivos,
	}
}

// ConsultarPerfilPorNombre es un alias para ObtenerPerfil (necesario para implementar la interfaz del handler)
func (s *PerfilService) ConsultarPerfilPorNombre(ctx context.Context, nombre string) (perfil.Perfil, error) {
	return s.ObtenerPerfil(ctx, nombre)
}

// ConsultarPerfiles es un alias para ListarPerfiles (necesario para implementar la interfaz del handler)
func (s *PerfilService) ConsultarPerfiles(ctx context.Context, criterios pagination.CriteriosPagina[perfil.Perfil]) (pagination.CriteriosPagina[perfil.Perfil], error) {
	return s.ListarPerfiles(ctx, criterios)
}
