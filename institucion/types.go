package institucion

import (
	"PlantillaInstitucional/pagination"

	"github.com/gofrs/uuid/v5"
)

type Persona struct {
	UuidPersona  uuid.UUID `json:"uuid_persona"`
	Nombre       string    `json:"nombre"`
	Apellido1    string    `json:"apellido1"`
	Apellido2    string    `json:"apellido2"`
	ClaveIssemym string    `json:"clave_issemym"`
	Curp         string    `json:"curp"`
	Rfc          string    `json:"rfc"`
	ImagenPerfil string    `json:"imagen_perfil"`
	Razon_i      string    `json:"razon_i"`
	Razon_m      string    `json:"razon_m"`
	Fecha_i      string    `json:"fecha_i"`
	Fecha_m      string    `json:"fecha_m"`
}

func (p Persona) GetNombreCompleto() string {
	if p.Apellido2 == "" {
		return p.Apellido1 + " " + p.Nombre
	}
	return p.Apellido1 + " " + p.Apellido2 + p.Nombre
}

// ToCriteria convierte el struct (usado como filtro desde el form)
// en el objeto de criterios de paginación.
func (p Persona) ToCriteria(page, size int, sortField, sortDir string) pagination.CriteriosPagina[Persona] {
	// Inicializamos el mapa
	filtros := make(map[string]any)

	if p.Nombre != "" {
		filtros["nombre"] = p.Nombre
	}
	if p.Apellido1 != "" {
		filtros["apellido1"] = p.Apellido1
	}
	if p.Apellido2 != "" {
		filtros["apellido2"] = p.Apellido2
	}
	if p.Rfc != "" {
		filtros["rfc"] = p.Rfc
	}
	if p.Curp != "" {
		filtros["curp"] = p.Curp
	}
	if p.ClaveIssemym != "" {
		filtros["clave_issemym"] = p.ClaveIssemym
	}

	return pagination.CriteriosPagina[Persona]{
		First:      (page - 1) * size,
		Size:       size,
		Index:      page,
		SortedBy:   sortField, // Corregido typo sortFiels
		Direction:  sortDir,
		FilteredBy: filtros,
	}
}
