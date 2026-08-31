package service

import (
	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/parametros"
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/archivos"
	"github.com/issemym-repos/logs"
)

// ParametrosRepositoryProvider define el contrato local con el repositorio de parámetros.
type ParametrosRepositoryProvider interface {
	ConsultarParametros(ctx context.Context, criterios pagination.CriteriosPagina[parametros.Parametro]) ([]parametros.Parametro, int, error)
	ConsultarParametroPorID(ctx context.Context, id uuid.UUID) (parametros.Parametro, error)
	ActualizarParametro(ctx context.Context, parametros parametros.Parametro) error
	EliminarParametro(ctx context.Context, id uuid.UUID) error
	CrearParametro(ctx context.Context, parametros parametros.Parametro) error
	GetJSONParamContext(ctx context.Context, key string) (string, error)
}

// ParametrosService es la implementación del servicio de parámetros.
type ParametrosService struct {
	logger   *logs.Logger
	repo     ParametrosRepositoryProvider
	validate *validator.Validate
	archivos *archivos.ManejadorArchivos
}

// ParametrosServiceProvider define el contrato del servicio de parámetros para el handler.
type ParametrosServiceProvider interface {
	GetParameter(ctx context.Context, key string) string

	ListarParametros(ctx context.Context) (pagination.CriteriosPagina[parametros.Parametro], error)
	ListarParametrosFiltrados(ctx context.Context, criterios pagination.CriteriosPagina[parametros.Parametro]) (pagination.CriteriosPagina[parametros.Parametro], error)
	ObtenerParametro(ctx context.Context, id string) (parametros.Parametro, error)
	CrearParametro(ctx context.Context, p parametros.Parametro) error
	ActualizarParametro(ctx context.Context, p parametros.Parametro) error
	EliminarParametro(ctx context.Context, id string) error
	ReportePdf(ctx context.Context) (string, error)
	ReporteExcel(ctx context.Context) (string, error)
}

// NewService crea una nueva instancia del servicio de parámetros.
func NewService(logger *logs.Logger, repo ParametrosRepositoryProvider, validate *validator.Validate, archivos *archivos.ManejadorArchivos) *ParametrosService {
	return &ParametrosService{
		logger:   logger,
		repo:     repo,
		validate: validate,
		archivos: archivos,
	}
}
