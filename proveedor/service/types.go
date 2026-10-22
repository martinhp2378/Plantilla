package service

import (
	"PlantillaInstitucional/proveedor"
	"PlantillaInstitucional/pagination"
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/logs"
)

type ProveedorRepositoryProvider interface {
	Consultar(ctx context.Context, criterios pagination.CriteriosPagina[proveedor.Proveedor]) ([]proveedor.Proveedor, int, error)
	ObtenerPorID(ctx context.Context, id uuid.UUID) (proveedor.Proveedor, error)
	Crear(ctx context.Context, item proveedor.Proveedor) error
	Actualizar(ctx context.Context, item proveedor.Proveedor) error
	Eliminar(ctx context.Context, id uuid.UUID) error
}

type ProveedorServiceProvider interface {
	Listar(ctx context.Context, criterios pagination.CriteriosPagina[proveedor.Proveedor]) (pagination.CriteriosPagina[proveedor.Proveedor], error)
	Obtener(ctx context.Context, id uuid.UUID) (proveedor.Proveedor, error)
	Crear(ctx context.Context, item proveedor.Proveedor) error
	Actualizar(ctx context.Context, item proveedor.Proveedor) error
	Eliminar(ctx context.Context, id uuid.UUID) error
}

type ProveedorService struct {
	repo     ProveedorRepositoryProvider
	validate *validator.Validate
	logger   *logs.Logger
}

func NewProveedorService(repo ProveedorRepositoryProvider, validate *validator.Validate, logger *logs.Logger) *ProveedorService {
	return &ProveedorService{
		repo:     repo,
		validate: validate,
		logger:   logger,
	}
}
