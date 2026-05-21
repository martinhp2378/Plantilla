package view

import (
	"PlantillaInstitucional/configuracion/usuario"
	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/shared/pagina"
)

type ParamsUsuario struct {
	Titulo    string
	Pag       pagina.Pagina
	Criterios pagination.CriteriosPagina[usuario.Usuario]
}
