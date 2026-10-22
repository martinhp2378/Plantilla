package repository

import (
	bd "PlantillaInstitucional/app/db"
	"PlantillaInstitucional/miperfil"
	"context"
	"database/sql"

	"github.com/gofrs/uuid/v5"
)

// RepositoryProvider define el contrato del repositorio para el módulo de Mi Perfil.
type RepositoryProvider interface {
	ObtenerPerfilUsuario(ctx context.Context, uuidPersona uuid.UUID) (miperfil.PerfilUsuario, error)
	ObtenerPasswordHash(ctx context.Context, uuidPersona uuid.UUID) (string, error)
	ActualizarPassword(ctx context.Context, uuidPersona uuid.UUID, passwordHash string) error
	ActualizarImagenPerfil(ctx context.Context, uuidPersona uuid.UUID, nombreArchivo string) error
}

type RowPerfil struct {
	UuidPersona  string         `db:"uuid_persona"`
	Nombre       string         `db:"nombre"`
	Apellido1    sql.NullString `db:"apellido1"`
	Apellido2    sql.NullString `db:"apellido2"`
	ClaveIssemym string         `db:"clave_issemym"`
	Curp         string         `db:"curp"`
	Rfc          string         `db:"rfc"`
	ImagenPerfil sql.NullString `db:"imagen_perfil"`
	Email        sql.NullString `db:"email"`
	Perfil       sql.NullString `db:"perfil"`
}

func (r RowPerfil) ToPerfilUsuario() miperfil.PerfilUsuario {
	return miperfil.PerfilUsuario{
		UuidPersona:  uuid.FromStringOrNil(r.UuidPersona),
		Nombre:       r.Nombre,
		Apellido1:    r.Apellido1.String,
		Apellido2:    r.Apellido2.String,
		ClaveIssemym: r.ClaveIssemym,
		Curp:         r.Curp,
		Rfc:          r.Rfc,
		ImagenPerfil: r.ImagenPerfil.String,
		Email:        r.Email.String,
		Perfil:       r.Perfil.String,
	}
}

// MiPerfilRepository es la implementación con AppPools.
type MiPerfilRepository struct {
	db *bd.AppPools
}

// NewMiPerfilRepository inicializa el repositorio.
func NewMiPerfilRepository(db *bd.AppPools) *MiPerfilRepository {
	return &MiPerfilRepository{db: db}
}
