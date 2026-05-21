package view

import (
	"PlantillaInstitucional/configuracion/policy"
	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/shared/pagina"
)

type ParamsPoliticas struct {
	Titulo    string
	Pag       pagina.Pagina
	Criterios pagination.CriteriosPagina[policy.Politica]
}
