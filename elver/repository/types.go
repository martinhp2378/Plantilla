package repository

import (
	bd "PlantillaInstitucional/app/db"
	"PlantillaInstitucional/elver"
	"time"

	"github.com/gofrs/uuid/v5"
)

var (
	colsTexto     = map[string]bool{"nombre": true, "codigo": true, "descripcion": true}
	colsNumericas = map[string]bool{}
	colsUUID      = map[string]bool{"id": true}
	colsOrden     = map[string]bool{"id": true, "nombre": true, "codigo": true, "fecha_i": true, "is_active": true}
)

type RowElver struct {
	ID          uuid.UUID `db:"id"`
	Nombre      string    `db:"nombre"`
	Codigo      string    `db:"codigo"`
	Descripcion string    `db:"descripcion"`
	IsActive    bool      `db:"is_active"`
	FechaI      time.Time `db:"fecha_i"`
	FechaM      time.Time `db:"fecha_m"`
}

func (r RowElver) ToEntity() elver.Elver {
	return elver.Elver{
		ID:          r.ID,
		Nombre:      r.Nombre,
		Codigo:      r.Codigo,
		Descripcion: r.Descripcion,
		IsActive:    r.IsActive,
		FechaI:      r.FechaI,
		FechaM:      r.FechaM,
	}
}

type ElverRepository struct {
	db *bd.AppPools
}

func NewElverRepository(db *bd.AppPools) *ElverRepository {
	return &ElverRepository{db: db}
}
