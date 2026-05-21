package service

import (
	"PlantillaInstitucional/configuracion/perfil"
	"PlantillaInstitucional/pagination"
	"context"
	"fmt"
)

// ListarPerfiles maneja la lógica de filtrado y paginación.
// Nota: Fusioné ListarPerfiles y ListarPerfilesFiltrados porque hacían lo mismo.
// Si no hay criterios, el controlador envía criterios vacíos, no el servicio.
func (s *PerfilService) ListarPerfiles(ctx context.Context, criterios pagination.CriteriosPagina[perfil.Perfil]) (pagination.CriteriosPagina[perfil.Perfil], error) {
	perfilesList, count, err := s.repo.ConsultarPerfiles(ctx, criterios)
	if err != nil {
		s.logger.ErrorContext(ctx, "Error consultando repositorio", err, "op", "ListarPerfiles")
		return pagination.CriteriosPagina[perfil.Perfil]{}, fmt.Errorf("error al listar perfiles: %w", err)
	}

	criterios.Data = perfilesList
	criterios.Count = count

	return criterios, nil
}

func (s *PerfilService) ObtenerPerfil(ctx context.Context, nombre string) (perfil.Perfil, error) {
	b, err := s.repo.ConsultarPerfilPorNombre(ctx, nombre)
	if err != nil {
		s.logger.ErrorContext(ctx, "Error obteniendo perfil", err, "nombre", nombre, "op", "ObtenerPerfil")
		return perfil.Perfil{}, err // El error ya viene envuelto desde el repo
	}

	return b, nil
}

func (s *PerfilService) CrearPerfil(ctx context.Context, p perfil.Perfil) error {
	// 1. Validación de estructura (Business Rules básicas)
	if err := s.validate.Struct(p); err != nil {
		s.logger.WarnContext(ctx, "Validación fallida al crear", "error", err, "op", "CrearPerfil")
		return fmt.Errorf("datos inválidos: %w", err)
	}

	// 2. Persistencia
	if err := s.repo.CrearPerfil(ctx, p); err != nil {
		s.logger.ErrorContext(ctx, "Error persistiendo perfil", err, "nombre", p.Nombre, "op", "CrearPerfil")
		return fmt.Errorf("no se pudo crear el perfil: %w", err)
	}

	s.logger.InfoContext(ctx, "Perfil creado exitosamente", "nombre", p.Nombre, "op", "CrearPerfil")
	return nil
}

func (s *PerfilService) ActualizarPerfil(ctx context.Context, p perfil.Perfil, nombreOld string) error {
	if err := s.validate.Struct(p); err != nil {
		s.logger.WarnContext(ctx, "Validación fallida al actualizar", "error", err, "nombre", p.Nombre, "op", "ActualizarPerfil")
		return fmt.Errorf("datos inválidos: %w", err)
	}

	// CORRECCIÓN IMPORTANTE: Capturamos el error explícitamente
	if err := s.repo.ActualizarPerfil(ctx, p, nombreOld); err != nil {
		s.logger.ErrorContext(ctx, "Error actualizando repositorio", err, "nombre", p.Nombre, "op", "ActualizarPerfil")
		return fmt.Errorf("no se pudo actualizar el perfil: %w", err)
	}

	s.logger.InfoContext(ctx, "Perfil actualizado exitosamente", "nombre", p.Nombre, "op", "ActualizarPerfil")
	return nil
}

func (s *PerfilService) EliminarPerfil(ctx context.Context, nombre string) error {
	if err := s.repo.EliminarPerfil(ctx, nombre); err != nil {
		s.logger.ErrorContext(ctx, "Error eliminando en repositorio", err, "nombre", nombre, "op", "EliminarPerfil")
		return fmt.Errorf("no se pudo eliminar el perfil: %w", err)
	}

	s.logger.InfoContext(ctx, "Perfil eliminado exitosamente", "nombre", nombre, "op", "EliminarPerfil")
	return nil
}
