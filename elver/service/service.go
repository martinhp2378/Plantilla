package service

import (
	"PlantillaInstitucional/elver"
	"PlantillaInstitucional/pagination"
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
)

func (s *ElverService) Listar(ctx context.Context, criterios pagination.CriteriosPagina[elver.Elver]) (pagination.CriteriosPagina[elver.Elver], error) {
	data, total, err := s.repo.Consultar(ctx, criterios)
	if err != nil {
		s.logger.ErrorContext(ctx, "Error listando elveres", err)
		return criterios, fmt.Errorf("error al consultar elveres: %w", err)
	}

	criterios.Data = data
	criterios.Count = total
	return criterios, nil
}

func (s *ElverService) Obtener(ctx context.Context, id uuid.UUID) (elver.Elver, error) {
	return s.repo.ObtenerPorID(ctx, id)
}

func (s *ElverService) Crear(ctx context.Context, item elver.Elver) error {
	if err := s.validate.Struct(item); err != nil {
		return err
	}
	return s.repo.Crear(ctx, item)
}

func (s *ElverService) Actualizar(ctx context.Context, item elver.Elver) error {
	if err := s.validate.Struct(item); err != nil {
		return err
	}
	return s.repo.Actualizar(ctx, item)
}

func (s *ElverService) Eliminar(ctx context.Context, id uuid.UUID) error {
	return s.repo.Eliminar(ctx, id)
}
