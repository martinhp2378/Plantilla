package service

import (
	"PlantillaInstitucional/institucion"
	"PlantillaInstitucional/pagination"
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
)

func (s *PersonaService) ListarPersonas(ctx context.Context, criterios pagination.CriteriosPagina[institucion.Persona]) (pagination.CriteriosPagina[institucion.Persona], error) {
	personasList, count, err := s.repo.ConsultarPersonas(ctx, criterios)
	if err != nil {
		s.logger.ErrorContext(ctx, "Error consultando repositorio", err, "op", "ListarPersonas")
		return pagination.CriteriosPagina[institucion.Persona]{}, fmt.Errorf("error al listar personas: %w", err)
	}

	var personasValores []institucion.Persona
	for _, p := range personasList {
		if p != nil {
			personasValores = append(personasValores, *p)
		}
	}

	criterios.Data = personasValores
	criterios.Count = count

	return criterios, nil
}

func (s *PersonaService) ObtenerPersona(ctx context.Context, idParam uuid.UUID) (institucion.Persona, error) {
	b, err := s.repo.ConsultarPersonaPorID(ctx, idParam)
	if err != nil {
		s.logger.ErrorContext(ctx, "Error obteniendo persona", err, "id", idParam, "op", "ObtenerPersona")
		return institucion.Persona{}, err // El error ya viene envuelto desde el repo
	}

	return *b, nil
}

func (s *PersonaService) CrearPersona(ctx context.Context, b institucion.Persona) error {
	// 1. Validación de estructura (Business Rules básicas)
	if err := s.validate.Struct(b); err != nil {
		s.logger.WarnContext(ctx, "Validación fallida al crear", "error", err, "op", "CrearPersona")
		return fmt.Errorf("datos inválidos: %w", err)
	}

	// 2. Persistencia
	newId, err := uuid.NewV4()
	if err != nil {
		s.logger.ErrorContext(ctx, "Creacion de uuid fallida", err, "persona", b.Nombre, "op", "CrearPersona")
		return fmt.Errorf("error al generar nuevoUUID %w", err)
	}

	b.UuidPersona = newId

	if err := s.repo.CrearPersona(ctx, b); err != nil {
		s.logger.ErrorContext(ctx, "Error persistiendo persona", err, "persona", b.Nombre, "op", "CrearPersona")
		return fmt.Errorf("no se pudo crear la persona: %w", err)
	}

	s.logger.InfoContext(ctx, "Persona creado exitosamente", "id", b.UuidPersona, "op", "CrearPersona")
	return nil
}

func (s *PersonaService) ActualizarPersona(ctx context.Context, b institucion.Persona) error {
	if err := s.validate.Struct(b); err != nil {
		s.logger.WarnContext(ctx, "Validación fallida al actualizar", "error", err, "id", b.UuidPersona, "op", "ActualizarPersona")
		return fmt.Errorf("datos inválidos: %w", err)
	}

	// CORRECCIÓN IMPORTANTE: Capturamos el error explícitamente
	if err := s.repo.ActualizarPersona(ctx, b); err != nil {
		s.logger.ErrorContext(ctx, "Error actualizando repositorio", err, "id", b.UuidPersona, "op", "ActualizarPersona")
		return fmt.Errorf("no se pudo actualizar la persona: %w", err)
	}

	s.logger.InfoContext(ctx, "Persona actualizado exitosamente", "id", b.UuidPersona, "op", "ActualizarPersona")
	return nil
}

func (s *PersonaService) EliminarPersona(ctx context.Context, idParam uuid.UUID) error {
	if err := s.repo.EliminarPersona(ctx, idParam); err != nil {
		s.logger.ErrorContext(ctx, "Error eliminando en repositorio", err, "id", idParam, "op", "EliminarPersona")
		return fmt.Errorf("no se pudo eliminar la persona: %w", err)
	}

	s.logger.InfoContext(ctx, "Persona eliminado exitosamente", "id", idParam, "op", "EliminarPersona")
	return nil
}

func (s *PersonaService) ObtenerTodasCorto(ctx context.Context) ([]institucion.Persona, error) {
	personasList, err := s.repo.ObtenerTodasCorto(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Error consultando repositorio", err, "op", "ObtenerTodasCorto")
		return nil, fmt.Errorf("error al listar todas las personas: %w", err)
	}

	var personasValores []institucion.Persona
	for _, p := range personasList {
		if p != nil {
			personasValores = append(personasValores, *p)
		}
	}

	return personasValores, nil
}

func (s *PersonaService) ActualizarImagenPerfil(ctx context.Context, uuid_persona uuid.UUID, imagen string) error {
	if err := s.repo.ActualizarImagenPerfil(ctx, uuid_persona, imagen); err != nil {
		s.logger.ErrorContext(ctx, "Error actualizando imagen en repositorio", err, "id", uuid_persona, "op", "ActualizarImagenPerfil")
		return fmt.Errorf("no se pudo actualizar la imagen de la persona: %w", err)
	}

	s.logger.InfoContext(ctx, "Imagen de persona actualizada exitosamente", "id", uuid_persona, "op", "ActualizarImagenPerfil")
	return nil
}
