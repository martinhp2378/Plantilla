package repository

import (
	bd "PlantillaInstitucional/app/db"
	"PlantillaInstitucional/institucion"
	"database/sql"

	"github.com/gofrs/uuid/v5"
)

type RowPersona struct {
	UuidPersona  string         `db:"uuid_persona"`
	Nombre       string         `db:"nombre"`
	Apellido1    sql.NullString `db:"apellido1"`
	Apellido2    sql.NullString `db:"apellido2"`
	ClaveIssemym sql.NullString `db:"clave_issemym"`
	Curp         sql.NullString `db:"curp"`
	Rfc          sql.NullString `db:"rfc"`
	ImagenPerfil sql.NullString `db:"imagen_perfil"`
	Razon_i      sql.NullString `db:"razon_i"`
	Razon_m      sql.NullString `db:"razon_m"`
	Fecha_i      sql.NullTime   `db:"fecha_i"`
	Fecha_m      sql.NullTime   `db:"fecha_m"`
}

func (r RowPersona) ToPersona() *institucion.Persona {
	return &institucion.Persona{
		UuidPersona:  uuid.FromStringOrNil(r.UuidPersona),
		Nombre:       r.Nombre,
		Apellido1:    r.Apellido1.String,
		Apellido2:    r.Apellido2.String,
		ClaveIssemym: r.ClaveIssemym.String,
		Curp:         r.Curp.String,
		Rfc:          r.Rfc.String,
		ImagenPerfil: r.ImagenPerfil.String,
		Razon_i:      r.Razon_i.String,
		Razon_m:      r.Razon_m.String,
		Fecha_i:      r.Fecha_i.Time.Format("2006-01-02 15:04:05"),
		Fecha_m:      r.Fecha_m.Time.Format("2006-01-02 15:04:05"),
	}
}

type PersonaRepository struct {
	db *bd.AppPools
}

func NewPersonaRepository(db *bd.AppPools) *PersonaRepository {
	return &PersonaRepository{
		db: db,
	}
}
