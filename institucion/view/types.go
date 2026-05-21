package view

import (
	"PlantillaInstitucional/institucion"
	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/shared/pagina"
)

type ParamPersona struct {
	Titulo    string
	Pag       pagina.Pagina
	Criterios pagination.CriteriosPagina[institucion.Persona]
	BaseURL   string
}
