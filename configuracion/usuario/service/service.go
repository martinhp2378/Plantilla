package service

import (
	"PlantillaInstitucional/configuracion/usuario"
	"PlantillaInstitucional/pagination"
	shared "PlantillaInstitucional/shared/password"
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
)

// ListarUsuarios maneja la lógica de filtrado y paginación.
func (s *UsuarioService) ListarUsuarios(ctx context.Context, criterios pagination.CriteriosPagina[usuario.Usuario]) (pagination.CriteriosPagina[usuario.Usuario], error) {
	usuariosList, count, err := s.repo.ConsultarUsuarios(ctx, criterios)
	if err != nil {
		s.logger.ErrorContext(ctx, "Error consultando repositorio", err, "op", "ListarUsuarios")
		return pagination.CriteriosPagina[usuario.Usuario]{}, fmt.Errorf("error al listar usuarios: %w", err)
	}

	criterios.Data = usuariosList
	criterios.Count = count

	return criterios, nil
}

func (s *UsuarioService) ObtenerUsuario(ctx context.Context, uuid_persona uuid.UUID) (usuario.Usuario, error) {
	b, err := s.repo.ConsultarUsuarioPorID(ctx, uuid_persona)
	if err != nil {
		s.logger.ErrorContext(ctx, "Error obteniendo usuario", err, "id", uuid_persona, "op", "ObtenerUsuario")
		return usuario.Usuario{}, err // El error ya viene envuelto desde el repo
	}

	return b, nil
}

func (s *UsuarioService) CrearUsuario(ctx context.Context, b usuario.Usuario, rawPassword string) error {
	// 1. Validación de estructura (Business Rules básicas)
	if err := s.validate.Struct(b); err != nil {
		s.logger.WarnContext(ctx, "Validación fallida al crear", "error", err, "op", "CrearUsuario")
		return fmt.Errorf("datos inválidos: %w", err)
	}

	// 2. Persistencia
	if b.UuidPersona == uuid.Nil {
		return fmt.Errorf("debe seleccionar una persona para asignarle el usuario")
	}

	// Si no se proporciona password, generar uno
	if rawPassword == "" {
		var err error
		rawPassword, err = shared.GenerateRandomPassword()
		if err != nil {
			s.logger.ErrorContext(ctx, "Generacion de password fallida", err, "email", b.Email, "op", "CrearUsuario")
			return fmt.Errorf("error generando contraseña: %w", err)
		}
	}

	hashed, err := shared.HashPassword(rawPassword)
	if err != nil {
		s.logger.ErrorContext(ctx, "Hasheo de password fallido", err, "email", b.Email, "op", "CrearUsuario")
		return fmt.Errorf("error asegurando la contraseña: %w", err)
	}

	b.Password = hashed

	if err := s.repo.CrearUsuario(ctx, b); err != nil {
		s.logger.ErrorContext(ctx, "Error persistiendo usuario", err, "email", b.Email, "op", "CrearUsuario")
		return fmt.Errorf("no se pudo crear el usuario: %w", err)
	}

	s.logger.InfoContext(ctx, "Usuario creado exitosamente", "id", b.UuidPersona, "op", "CrearUsuario")
	return nil
}

func (s *UsuarioService) ActualizarUsuario(ctx context.Context, b usuario.Usuario) error {
	if err := s.validate.Struct(b); err != nil {
		s.logger.WarnContext(ctx, "Validación fallida al actualizar", "error", err, "id", b.UuidPersona, "op", "ActualizarUsuario")
		return fmt.Errorf("datos inválidos: %w", err)
	}

	bExistente, err := s.ObtenerUsuario(ctx, b.UuidPersona)
	if err == nil && b.Password == "" {
		b.Password = bExistente.Password
	}

	// CORRECCIÓN IMPORTANTE: Capturamos el error explícitamente
	if err := s.repo.ActualizarUsuario(ctx, b); err != nil {
		s.logger.ErrorContext(ctx, "Error actualizando repositorio", err, "id", b.UuidPersona, "op", "ActualizarUsuario")
		return fmt.Errorf("no se pudo actualizar el usuario: %w", err)
	}

	s.logger.InfoContext(ctx, "Usuario actualizado exitosamente", "id", b.UuidPersona, "op", "ActualizarUsuario")
	return nil
}

func (s *UsuarioService) EliminarUsuario(ctx context.Context, uuid_persona uuid.UUID) error {
	if err := s.repo.EliminarUsuario(ctx, uuid_persona); err != nil {
		s.logger.ErrorContext(ctx, "Error eliminando en repositorio", err, "id", uuid_persona, "op", "EliminarUsuario")
		return fmt.Errorf("no se pudo eliminar el usuario: %w", err)
	}

	s.logger.InfoContext(ctx, "Usuario eliminado exitosamente", "id", uuid_persona, "op", "EliminarUsuario")
	return nil
}

func (s *UsuarioService) ActualizarPassword(ctx context.Context, uuid_persona uuid.UUID, rawPassword string) error {
	hashed, err := shared.HashPassword(rawPassword)
	if err != nil {
		s.logger.ErrorContext(ctx, "Hasheo de password fallido", err, "id", uuid_persona, "op", "ActualizarPassword")
		return fmt.Errorf("error asegurando la contraseña: %w", err)
	}

	if err := s.repo.ActualizarPassword(ctx, uuid_persona, hashed); err != nil {
		s.logger.ErrorContext(ctx, "Error actualizando password en repositorio", err, "id", uuid_persona, "op", "ActualizarPassword")
		return fmt.Errorf("no se pudo actualizar la contraseña: %w", err)
	}
	s.logger.InfoContext(ctx, "Password actualizado exitosamente", "id", uuid_persona, "op", "ActualizarPassword")
	return nil
}
