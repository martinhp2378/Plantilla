package service

import (
	"PlantillaInstitucional/crud"
	"PlantillaInstitucional/pagination"
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
)

func (s *CrudService) ListarCrud(ctx context.Context, criterios pagination.CriteriosPagina[crud.Crud]) (pagination.CriteriosPagina[crud.Crud], error) {
	itemsList, count, err := s.repo.ConsultarCrud(ctx, criterios)
	if err != nil {
		s.logger.ErrorContext(ctx, "Error consultando repositorio crud", err, "op", "ListarCrud")
		return pagination.CriteriosPagina[crud.Crud]{}, fmt.Errorf("error al listar registros: %w", err)
	}

	items := make([]crud.Crud, 0, len(itemsList))
	for _, item := range itemsList {
		if item != nil {
			items = append(items, *item)
		}
	}

	criterios.Data = items
	criterios.Count = count

	return criterios, nil
}

func (s *CrudService) ObtenerCrud(ctx context.Context, id uuid.UUID) (crud.Crud, error) {
	item, err := s.repo.ConsultarCrudPorID(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "Error obteniendo registro crud", err, "id", id, "op", "ObtenerCrud")
		return crud.Crud{}, err
	}

	return *item, nil
}

func (s *CrudService) CrearCrud(ctx context.Context, c crud.Crud) error {
	// Validación con validator
	if err := s.validate.Struct(c); err != nil {
		s.logger.WarnContext(ctx, "Validación fallida al crear crud", "error", err, "op", "CrearCrud")
		return fmt.Errorf("datos inválidos: %w", err)
	}

	newId, err := uuid.NewV4()
	if err != nil {
		s.logger.ErrorContext(ctx, "Error generando UUID", err, "op", "CrearCrud")
		return fmt.Errorf("error al generar identificador: %w", err)
	}
	c.Id = newId

	if err := s.repo.CrearCrud(ctx, c); err != nil {
		s.logger.ErrorContext(ctx, "Error persistiendo registro crud", err, "id", c.Id, "op", "CrearCrud")
		return fmt.Errorf("no se pudo crear el registro: %w", err)
	}

	s.logger.InfoContext(ctx, "Registro crud creado exitosamente", "id", c.Id, "op", "CrearCrud")
	return nil
}

func (s *CrudService) ActualizarCrud(ctx context.Context, c crud.Crud) error {
	if err := s.validate.Struct(c); err != nil {
		s.logger.WarnContext(ctx, "Validación fallida al actualizar crud", "error", err, "id", c.Id, "op", "ActualizarCrud")
		return fmt.Errorf("datos inválidos: %w", err)
	}

	if err := s.repo.ActualizarCrud(ctx, c); err != nil {
		s.logger.ErrorContext(ctx, "Error actualizando repositorio crud", err, "id", c.Id, "op", "ActualizarCrud")
		return fmt.Errorf("no se pudo actualizar el registro: %w", err)
	}

	s.logger.InfoContext(ctx, "Registro crud actualizado exitosamente", "id", c.Id, "op", "ActualizarCrud")
	return nil
}

func (s *CrudService) EliminarCrud(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.EliminarCrud(ctx, id); err != nil {
		s.logger.ErrorContext(ctx, "Error eliminando registro crud", err, "id", id, "op", "EliminarCrud")
		return fmt.Errorf("no se pudo eliminar el registro: %w", err)
	}

	s.logger.InfoContext(ctx, "Registro crud eliminado exitosamente", "id", id, "op", "EliminarCrud")
	return nil
}

func (s *CrudService) ObtenerTodos(ctx context.Context) ([]crud.Crud, error) {
	items, err := s.repo.ObtenerTodos(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Error consultando todos los registros crud", err, "op", "ObtenerTodos")
		return nil, fmt.Errorf("error al listar todos los registros: %w", err)
	}
	return items, nil
}
