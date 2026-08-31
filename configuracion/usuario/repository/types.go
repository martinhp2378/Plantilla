package repository

import (
	bd "PlantillaInstitucional/app/db"
	"PlantillaInstitucional/configuracion/usuario"
	"database/sql"

	"github.com/gofrs/uuid/v5"
)

type RowUsuario struct {
	UuidPersona    string         `db:"uuid_persona"`
	Email          string         `db:"email"`
	Password       string         `db:"password"`
	Perfil         string         `db:"perfil"`
	Razon_i        sql.NullString `db:"razon_i"`
	Razon_m        sql.NullString `db:"razon_m"`
	Fecha_i        sql.NullTime   `db:"fecha_i"`
	Fecha_m        sql.NullTime   `db:"fecha_m"`
	NombreCompleto string         `db:"nombre_completo"`
}

func (r RowUsuario) ToUsuario() usuario.Usuario {
	var fechaI, fechaM string
	if r.Fecha_i.Valid {
		fechaI = r.Fecha_i.Time.Format("2006-01-02 15:04:05")
	}
	if r.Fecha_m.Valid {
		fechaM = r.Fecha_m.Time.Format("2006-01-02 15:04:05")
	}
	return usuario.Usuario{
		UuidPersona:    uuid.FromStringOrNil(r.UuidPersona),
		Email:          r.Email,
		Password:       r.Password,
		Perfil:         r.Perfil,
		Razon_i:        r.Razon_i.String,
		Razon_m:      	r.Razon_m.String,
		Fecha_i:        fechaI,
		Fecha_m:        fechaM,
		NombreCompleto: r.NombreCompleto,
	}
}

// UsuarioRepository es la implementación del repositorio de usuarios.
type UsuarioRepository struct {
	db *bd.AppPools
}

func NewUsuarioRepository(db *bd.AppPools) *UsuarioRepository {
	return &UsuarioRepository{
		db: db,
	}
}
