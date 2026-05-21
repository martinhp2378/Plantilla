package usuario

import (
	"PlantillaInstitucional/pagination"

	"github.com/gofrs/uuid/v5"
)

type Usuario struct {
	UuidPersona    uuid.UUID `db:"uuid_persona" json:"uuid_persona"`
	Email          string    `db:"email" json:"email"`
	Password       string    `db:"password" json:"password"`
	Perfil         string    `db:"perfil" json:"perfil"`
	Razon_i        string    `db:"razon_i" json:"razon_i"`
	Razon_m        string    `db:"razon_m" json:"razon_m"`
	Fecha_i        string    `db:"fecha_i" json:"fecha_i"`
	Fecha_m        string    `db:"fecha_m" json:"fecha_m"`
	NombreCompleto string    `db:"nombre_completo" json:"nombre_completo"`
}

// ToCriteria convierte el struct (usado como filtro desde el form)
// en el objeto de criterios de paginación.
func (u Usuario) ToCriteria(page, size int, sortField, sortDir string) pagination.CriteriosPagina[Usuario] {
	// Inicializamos el mapa
	filtros := make(map[string]any)

	if u.Email != "" {
		filtros["email"] = u.Email
	}
	if u.Perfil != "" {
		filtros["perfil"] = u.Perfil
	}
	if u.NombreCompleto != "" {
		filtros["nombre_completo"] = u.NombreCompleto
	}

	return pagination.CriteriosPagina[Usuario]{
		First:      (page - 1) * size,
		Size:       size,
		Index:      page,
		SortedBy:   sortField,
		Direction:  sortDir,
		FilteredBy: filtros,
	}
}
