package crud

import (
	"PlantillaInstitucional/pagination"

	"github.com/gofrs/uuid/v5"
)

// Crud representa el modelo de datos de ejemplo que reúne un campo representativo
// de cada tipo de dato común utilizado en aplicaciones web institucionales.
type Crud struct {
	Id         uuid.UUID `db:"id" json:"id"`
	TextoCorto string    `db:"texto_corto" json:"texto_corto" validate:"required,max=150"`
	TextoLargo string    `db:"texto_largo" json:"texto_largo"`
	Entero     int       `db:"entero" json:"entero"`
	Decimal    float64   `db:"decimal" json:"decimal"`
	Booleano   bool      `db:"booleano" json:"booleano"`
	Fecha      string    `db:"fecha" json:"fecha" validate:"required"`
	Hora       string    `db:"hora" json:"hora" validate:"required"`
	FechaHora  string    `db:"fecha_hora" json:"fecha_hora"`
	Seleccion  string    `db:"seleccion" json:"seleccion" validate:"required"`
	Correo     string    `db:"correo" json:"correo" validate:"required,email"`
	Telefono   string    `db:"telefono" json:"telefono"`
	URL        string    `db:"url" json:"url"`
	Color      string    `db:"color" json:"color"`
	TenantId   uuid.UUID `db:"tenant_id" json:"tenant_id"`
	Razon_i    string    `db:"razon_i" json:"razon_i"`
	Razon_m    string    `db:"razon_m" json:"razon_m"`
	Fecha_i    string    `db:"fecha_i" json:"fecha_i"`
	Fecha_m    string    `db:"fecha_m" json:"fecha_m"`
}

// ToCriteria convierte el struct (usado como filtro desde el formulario)
// en el objeto genérico de criterios de paginación.
func (c Crud) ToCriteria(page, size int, sortField, sortDir string) pagination.CriteriosPagina[Crud] {
	filtros := make(map[string]any)

	if c.TextoCorto != "" {
		filtros["texto_corto"] = c.TextoCorto
	}
	if c.Seleccion != "" {
		filtros["seleccion"] = c.Seleccion
	}
	if c.Correo != "" {
		filtros["correo"] = c.Correo
	}

	return pagination.CriteriosPagina[Crud]{
		First:      (page - 1) * size,
		Size:       size,
		Index:      page,
		SortedBy:   sortField,
		Direction:  sortDir,
		FilteredBy: filtros,
	}
}
