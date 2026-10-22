package repository

import (
	"PlantillaInstitucional/miperfil"
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

// ObtenerPerfilUsuario recupera los datos personales y de cuenta del usuario.
func (r *MiPerfilRepository) ObtenerPerfilUsuario(ctx context.Context, uuidPersona uuid.UUID) (miperfil.PerfilUsuario, error) {
	var row RowPerfil
	err := r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, SelectPerfilUsuario, uuidPersona)
		if err != nil {
			return fmt.Errorf("ObtenerPerfilUsuario query: %w", err)
		}
		row, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[RowPerfil])
		if err != nil {
			return fmt.Errorf("ObtenerPerfilUsuario collect: %w", err)
		}
		return nil
	})
	if err != nil {
		return miperfil.PerfilUsuario{}, fmt.Errorf("ObtenerPerfilUsuario: %w", err)
	}
	return row.ToPerfilUsuario(), nil
}

// ObtenerPasswordHash obtiene el hash actual de la contraseña del usuario.
func (r *MiPerfilRepository) ObtenerPasswordHash(ctx context.Context, uuidPersona uuid.UUID) (string, error) {
	var hash string
	err := r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, SelectPasswordHash, uuidPersona).Scan(&hash)
	})
	if err != nil {
		return "", fmt.Errorf("ObtenerPasswordHash: %w", err)
	}
	return hash, nil
}

// ActualizarPassword modifica la contraseña hasheada del usuario.
func (r *MiPerfilRepository) ActualizarPassword(ctx context.Context, uuidPersona uuid.UUID, passwordHash string) error {
	return r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, UpdatePasswordUsuario, passwordHash, uuidPersona)
		if err != nil {
			return fmt.Errorf("ActualizarPassword: %w", err)
		}
		if ct.RowsAffected() == 0 {
			return fmt.Errorf("no se encontró usuario con id: %s", uuidPersona)
		}
		return nil
	})
}

// ActualizarImagenPerfil actualiza el nombre del archivo de la foto de perfil en persona.
func (r *MiPerfilRepository) ActualizarImagenPerfil(ctx context.Context, uuidPersona uuid.UUID, nombreArchivo string) error {
	return r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, UpdateImagenPersona, nombreArchivo, uuidPersona)
		if err != nil {
			return fmt.Errorf("ActualizarImagenPerfil: %w", err)
		}
		if ct.RowsAffected() == 0 {
			return fmt.Errorf("no se encontró persona con id: %s", uuidPersona)
		}
		return nil
	})
}
