package policy

import "PlantillaInstitucional/pagination"

type Politica struct {
	Nombre      string `db:"nombre" validate:"required"`
	Descripcion string `db:"descripcion"`
	Perfil      string `db:"perfil" validate:"required"`
	Recurso     string `db:"recurso" validate:"required"`
	Accion      string `db:"accion" validate:"required"`
	PathPrefix  string `db:"path_prefix" validate:"required"`
	Reglas      string `db:"reglas"`
}

func (p Politica) ToCriteria(page, size int, sortField, sortDir string) pagination.CriteriosPagina[Politica] {
	filtros := make(map[string]any)
	if p.Nombre != "" {
		filtros["nombre"] = p.Nombre
	}
	if p.Perfil != "" {
		filtros["perfil"] = p.Perfil
	}
	if p.Recurso != "" {
		filtros["recurso"] = p.Recurso
	}
	if p.Accion != "" {
		filtros["accion"] = p.Accion
	}

	return pagination.CriteriosPagina[Politica]{
		First:      (page - 1) * size,
		Size:       size,
		Index:      page,
		SortedBy:   sortField,
		Direction:  sortDir,
		FilteredBy: filtros,
	}
}
