package service

import (
	"PlantillaInstitucional/crud"
	"PlantillaInstitucional/pagination"
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/archivos"
	"github.com/issemym-repos/logs"
)

type CrudRepositoryProvider interface {
	ConsultarCrud(ctx context.Context, criterios pagination.CriteriosPagina[crud.Crud]) ([]*crud.Crud, int, error)
	ConsultarCrudPorID(ctx context.Context, id uuid.UUID) (*crud.Crud, error)
	CrearCrud(ctx context.Context, c crud.Crud) error
	CrearCrudMasivo(ctx context.Context, items []crud.Crud) error
	ActualizarCrud(ctx context.Context, c crud.Crud) error
	EliminarCrud(ctx context.Context, id uuid.UUID) error
	ObtenerTodos(ctx context.Context) ([]crud.Crud, error)
}

type CrudService struct {
	repo     CrudRepositoryProvider
	logger   *logs.Logger
	validate *validator.Validate
	archivos *archivos.ManejadorArchivos
}

func NewCrudService(repo CrudRepositoryProvider, validate *validator.Validate, logger *logs.Logger, archivos *archivos.ManejadorArchivos) *CrudService {
	return &CrudService{
		repo:     repo,
		logger:   logger,
		validate: validate,
		archivos: archivos,
	}
}
