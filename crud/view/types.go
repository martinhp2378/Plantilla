package view

import (
	"PlantillaInstitucional/crud"
	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/shared/pagina"
)

type ParamCrud struct {
	Titulo    string
	Pag       pagina.Pagina
	Criterios pagination.CriteriosPagina[crud.Crud]
	BaseURL   string
}
