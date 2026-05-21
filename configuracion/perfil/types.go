package perfil

import (
	"PlantillaInstitucional/pagination"
)

type Perfil struct {
	Nombre string
}

// ToCriteria convierte el struct (usado como filtro desde el form)
// en el objeto de criterios de paginación.
func (p Perfil) ToCriteria(page, size int, sortField, sortDir string) pagination.CriteriosPagina[Perfil] {
	// Inicializamos el mapa
	filtros := make(map[string]any)

	if p.Nombre != "" {
		filtros["nombre"] = p.Nombre
	}

	return pagination.CriteriosPagina[Perfil]{
		First:      (page - 1) * size,
		Size:       size,
		Index:      page,
		SortedBy:   sortField,
		Direction:  sortDir,
		FilteredBy: filtros,
	}
}
