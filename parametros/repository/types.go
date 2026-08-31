package repository

import (
	conexionBd "PlantillaInstitucional/app/db"
	"PlantillaInstitucional/parametros"
	"database/sql"

	"github.com/gofrs/uuid/v5"
)

// ParametrosRepository es la implementación del repositorio de parámetros.
type ParametrosRepository struct {
	db     *conexionBd.AppPools
}

func NewRepository(db *conexionBd.AppPools) *ParametrosRepository {
	return &ParametrosRepository{
		db:     db,
	}
}

type RowParametro struct {
	ParametroId uuid.UUID
	Clave       string
	Valor       string
	Descripcion sql.NullString
	TipoDato    string
	IsActive    bool
	RazonI      sql.NullString
	RazonM      sql.NullString
	FechaI      sql.NullTime
	FechaM      sql.NullTime
}

func (r RowParametro) ToParametro() parametros.Parametro {
	return parametros.Parametro{
		ParametroId: r.ParametroId,
		Clave:       r.Clave,
		Valor:       r.Valor,
		Descripcion: r.Descripcion.String,
		TipoDato:    r.TipoDato,
		IsActive:    r.IsActive,
		RazonI:      r.RazonI.String,
		RazonM:      r.RazonM.String,
		FechaI:      r.FechaI.Time,
		FechaM:      r.FechaM.Time,
	}
}

func ToParametros(rows []RowParametro) []parametros.Parametro {
	params := make([]parametros.Parametro, len(rows))
	for i, row := range rows {
		params[i] = row.ToParametro()
	}
	return params
}
