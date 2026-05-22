package view

import (
	"PlantillaInstitucional/configuracion/perfil"
	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/shared/pagina"
)

type ParamsPerfiles struct {
	Titulo    string
	Pag       pagina.Pagina
	Criterios pagination.CriteriosPagina[perfil.Perfil]
}
