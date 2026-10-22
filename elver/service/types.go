package service

import (
	"PlantillaInstitucional/elver"
	"PlantillaInstitucional/pagination"
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/archivos"
	"github.com/issemym-repos/logs"
)

type ElverRepositoryProvider interface {
	Consultar(ctx context.Context, criterios pagination.CriteriosPagina[elver.Elver]) ([]elver.Elver, int, error)
	ObtenerPorID(ctx context.Context, id uuid.UUID) (elver.Elver, error)
	ObtenerTodos(ctx context.Context) ([]elver.Elver, error)
	Crear(ctx context.Context, item elver.Elver) error
	Actualizar(ctx context.Context, item elver.Elver) error
	Eliminar(ctx context.Context, id uuid.UUID) error
}

type ElverServiceProvider interface {
	Listar(ctx context.Context, criterios pagination.CriteriosPagina[elver.Elver]) (pagination.CriteriosPagina[elver.Elver], error)
	Obtener(ctx context.Context, id uuid.UUID) (elver.Elver, error)
	Crear(ctx context.Context, item elver.Elver) error
	Actualizar(ctx context.Context, item elver.Elver) error
	Eliminar(ctx context.Context, id uuid.UUID) error
	ReportePdf(ctx context.Context) (string, error)
	ReporteExcel(ctx context.Context) (string, error)
}

type ElverService struct {
	repo     ElverRepositoryProvider
	validate *validator.Validate
	logger   *logs.Logger
	archivos *archivos.ManejadorArchivos
}

func NewElverService(repo ElverRepositoryProvider, validate *validator.Validate, logger *logs.Logger, archivos *archivos.ManejadorArchivos) *ElverService {
	return &ElverService{
		repo:     repo,
		validate: validate,
		logger:   logger,
		archivos: archivos,
	}
}
