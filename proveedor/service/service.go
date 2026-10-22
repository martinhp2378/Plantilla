package service

import (
	"PlantillaInstitucional/proveedor"
	"PlantillaInstitucional/pagination"
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
)

func (s *ProveedorService) Listar(ctx context.Context, criterios pagination.CriteriosPagina[proveedor.Proveedor]) (pagination.CriteriosPagina[proveedor.Proveedor], error) {
	data, total, err := s.repo.Consultar(ctx, criterios)
	if err != nil {
		s.logger.ErrorContext(ctx, "Error listando proveedores", err)
		return criterios, fmt.Errorf("error al consultar proveedores: %w", err)
	}

	criterios.Data = data
	criterios.Count = total
	return criterios, nil
}

func (s *ProveedorService) Obtener(ctx context.Context, id uuid.UUID) (proveedor.Proveedor, error) {
	return s.repo.ObtenerPorID(ctx, id)
}

func (s *ProveedorService) Crear(ctx context.Context, item proveedor.Proveedor) error {
	if err := s.validate.Struct(item); err != nil {
		return err
	}
	return s.repo.Crear(ctx, item)
}

func (s *ProveedorService) Actualizar(ctx context.Context, item proveedor.Proveedor) error {
	if err := s.validate.Struct(item); err != nil {
		return err
	}
	return s.repo.Actualizar(ctx, item)
}

func (s *ProveedorService) Eliminar(ctx context.Context, id uuid.UUID) error {
	return s.repo.Eliminar(ctx, id)
}
