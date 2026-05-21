package parametros

import (
	"PlantillaInstitucional/pagination"
	"time"

	"github.com/gofrs/uuid/v5"
)

type Parametro struct {
	ParametroId uuid.UUID
	Clave       string `validate:"required,min=3,max=100"`
	Valor       string `validate:"required,min=1"`
	Descripcion string
	TipoDato    string `validate:"required,oneof=STRING INT BOOLEAN JSON"`
	IsActive    bool
	RazonI      string
	RazonM      string
	FechaI      time.Time
	FechaM      time.Time
}

// ToCriteria converts the struct (used as filter from form)
// into the pagination criteria object.
func (p Parametro) ToCriteria(page, size int, sortField, sortDir string) pagination.CriteriosPagina[Parametro] {
	// Initialize map
	filtros := make(map[string]any)

	if p.Clave != "" {
		filtros["clave"] = p.Clave
	}
	if p.Valor != "" {
		filtros["valor"] = p.Valor
	}
	if p.Descripcion != "" {
		filtros["descripcion"] = p.Descripcion
	}
	if p.TipoDato != "" {
		filtros["tipo_dato"] = p.TipoDato
	}
	// For booleans or other types, we might need to handle them carefully.
	// Assuming IsActive is handled if needed, but usually filters are text-based or specific.
	// If IsActive is part of the filter form as a checkbox/select:
	// if p.IsActive { ... } - but here it's a bool in struct.
	// Often filters use strings or pointers to handle "not set" vs "false".
	// For now, let's stick to the text fields as per previous implementation, adding Descripcion and TipoDato.

	return pagination.CriteriosPagina[Parametro]{
		First:      (page - 1) * size,
		Size:       size,
		Index:      page,
		SortedBy:   sortField,
		Direction:  sortDir,
		FilteredBy: filtros,
	}
}
