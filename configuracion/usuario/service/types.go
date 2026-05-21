package service

import (
	"PlantillaInstitucional/configuracion/usuario"
	"PlantillaInstitucional/pagination"
	sharedEvents "PlantillaInstitucional/shared/events"
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/archivos"
	"github.com/issemym-repos/logs"
)

// UsuarioRepositoryProvider define el contrato local con el repositorio de usuarios.
type UsuarioRepositoryProvider interface {
	ConsultarUsuarios(ctx context.Context, criterios pagination.CriteriosPagina[usuario.Usuario]) ([]usuario.Usuario, int, error)
	ConsultarUsuarioPorID(ctx context.Context, uuid_persona uuid.UUID) (usuario.Usuario, error)
	CrearUsuario(ctx context.Context, u usuario.Usuario) error
	ActualizarUsuario(ctx context.Context, u usuario.Usuario) error
	EliminarUsuario(ctx context.Context, uuid_persona uuid.UUID) error

	ActualizarPassword(ctx context.Context, uuid_persona uuid.UUID, password string) error
}

// UsuarioService es la implementación del servicio de usuarios.
type UsuarioService struct {
	repo            UsuarioRepositoryProvider
	validate        *validator.Validate
	logger          *logs.Logger
	archivos        *archivos.ManejadorArchivos
	eventDispatcher *sharedEvents.Dispatcher
}

// UsuarioServiceProvider define el contrato del servicio de usuarios para el handler.
type UsuarioServiceProvider interface {
	ConsultarUsuarios(ctx context.Context, criterios pagination.CriteriosPagina[usuario.Usuario]) (pagination.CriteriosPagina[usuario.Usuario], error)
	ConsultarUsuarioPorID(ctx context.Context, uuid_persona uuid.UUID) (usuario.Usuario, error)
	ObtenerUsuario(ctx context.Context, uuid_persona uuid.UUID) (usuario.Usuario, error)
	ListarUsuarios(ctx context.Context, criterios pagination.CriteriosPagina[usuario.Usuario]) (pagination.CriteriosPagina[usuario.Usuario], error)
	CrearUsuario(ctx context.Context, u usuario.Usuario, rawPassword string) error
	ActualizarUsuario(ctx context.Context, u usuario.Usuario) error
	EliminarUsuario(ctx context.Context, uuid_persona uuid.UUID) error
	ReportePdf(ctx context.Context) (string, error)
	ReporteExcel(ctx context.Context) (string, error)
	ActualizarPassword(ctx context.Context, uuid_persona uuid.UUID, rawPassword string) error
}

// NewUsuarioService inicializa el servicio.
func NewUsuarioService(repo UsuarioRepositoryProvider, validate *validator.Validate, logger *logs.Logger, archivos *archivos.ManejadorArchivos, eventDispatcher *sharedEvents.Dispatcher) *UsuarioService {
	return &UsuarioService{
		repo:            repo,
		validate:        validate,
		logger:          logger,
		archivos:        archivos,
		eventDispatcher: eventDispatcher,
	}
}

// ConsultarUsuarioPorID es un alias para ObtenerUsuario (necesario para implementar la interfaz del handler)
func (s *UsuarioService) ConsultarUsuarioPorID(ctx context.Context, uuid_persona uuid.UUID) (usuario.Usuario, error) {
	return s.ObtenerUsuario(ctx, uuid_persona)
}

// ConsultarUsuarios es un alias para ListarUsuarios (necesario para implementar la interfaz del handler)
func (s *UsuarioService) ConsultarUsuarios(ctx context.Context, criterios pagination.CriteriosPagina[usuario.Usuario]) (pagination.CriteriosPagina[usuario.Usuario], error) {
	return s.ListarUsuarios(ctx, criterios)
}
