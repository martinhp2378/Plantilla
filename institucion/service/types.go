package service

import (
	"PlantillaInstitucional/institucion"
	"PlantillaInstitucional/pagination"
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/archivos"
	"github.com/issemym-repos/logs"
)

// PersonaRepositoryProvider define el contrato del repositorio requerido por el servicio.
type PersonaRepositoryProvider interface {
	ConsultarPersonas(ctx context.Context, criterios pagination.CriteriosPagina[institucion.Persona]) ([]*institucion.Persona, int, error)
	ConsultarPersonaPorID(ctx context.Context, uuid_persona uuid.UUID) (*institucion.Persona, error)
	CrearPersona(ctx context.Context, p institucion.Persona) error
	ActualizarPersona(ctx context.Context, p institucion.Persona) error
	EliminarPersona(ctx context.Context, uuid_persona uuid.UUID) error
	ObtenerTodasCorto(ctx context.Context) ([]*institucion.Persona, error)
	ActualizarImagenPerfil(ctx context.Context, uuid_persona uuid.UUID, imagen string) error
}

type PersonaService struct {
	repo     PersonaRepositoryProvider
	validate *validator.Validate
	logger   *logs.Logger
	archivos *archivos.ManejadorArchivos
}

func NewPersonaService(repo PersonaRepositoryProvider, validate *validator.Validate, logger *logs.Logger, archivos *archivos.ManejadorArchivos) *PersonaService {
	return &PersonaService{
		repo:     repo,
		validate: validate,
		logger:   logger,
		archivos: archivos,
	}
}
