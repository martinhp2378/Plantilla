package repository

import (
	bd "PlantillaInstitucional/app/db"
	"PlantillaInstitucional/crud"
	"database/sql"

	"github.com/gofrs/uuid/v5"
)

type RowCrud struct {
	Id         string          `db:"id"`
	TextoCorto string          `db:"texto_corto"`
	TextoLargo sql.NullString  `db:"texto_largo"`
	Entero     sql.NullInt64   `db:"entero"`
	Decimal    sql.NullFloat64 `db:"decimal"`
	Booleano   sql.NullBool    `db:"booleano"`
	Fecha      sql.NullString  `db:"fecha"`
	Hora       sql.NullString  `db:"hora"`
	FechaHora  sql.NullTime    `db:"fecha_hora"`
	Seleccion  string          `db:"seleccion"`
	Correo     string          `db:"correo"`
	Telefono   sql.NullString  `db:"telefono"`
	URL        sql.NullString  `db:"url"`
	Color      sql.NullString  `db:"color"`
	TenantId   sql.NullString  `db:"tenant_id"`
	Razon_i    sql.NullString  `db:"razon_i"`
	Razon_m    sql.NullString  `db:"razon_m"`
	Fecha_i    sql.NullTime    `db:"fecha_i"`
	Fecha_m    sql.NullTime    `db:"fecha_m"`
}

func (r RowCrud) ToCrud() *crud.Crud {
	var fechaHora, fechaI, fechaM string
	if r.FechaHora.Valid {
		fechaHora = r.FechaHora.Time.Format("2006-01-02T15:04")
	}
	if r.Fecha_i.Valid {
		fechaI = r.Fecha_i.Time.Format("2006-01-02 15:04:05")
	}
	if r.Fecha_m.Valid {
		fechaM = r.Fecha_m.Time.Format("2006-01-02 15:04:05")
	}

	horaStr := r.Hora.String
	if len(horaStr) >= 5 {
		horaStr = horaStr[:5]
	}

	colorVal := r.Color.String
	if colorVal == "" {
		colorVal = "#691C32"
	}

	return &crud.Crud{
		Id:         uuid.FromStringOrNil(r.Id),
		TextoCorto: r.TextoCorto,
		TextoLargo: r.TextoLargo.String,
		Entero:     int(r.Entero.Int64),
		Decimal:    r.Decimal.Float64,
		Booleano:   r.Booleano.Bool,
		Fecha:      r.Fecha.String,
		Hora:       horaStr,
		FechaHora:  fechaHora,
		Seleccion:  r.Seleccion,
		Correo:     r.Correo,
		Telefono:   r.Telefono.String,
		URL:        r.URL.String,
		Color:      colorVal,
		TenantId:   uuid.FromStringOrNil(r.TenantId.String),
		Razon_i:    r.Razon_i.String,
		Razon_m:    r.Razon_m.String,
		Fecha_i:    fechaI,
		Fecha_m:    fechaM,
	}
}

type CrudRepository struct {
	db *bd.AppPools
}

func NewCrudRepository(db *bd.AppPools) *CrudRepository {
	return &CrudRepository{db: db}
}
