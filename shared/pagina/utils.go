package pagina

func (p Pagina) URL() string {
	switch p {
	case Inicio:
		return GlobalRutas.Dashboard.Base
	case Usuarios:
		return GlobalRutas.Usuarios.Base
	case Perfil:
		return GlobalRutas.Perfiles.Base
	case Politicas:
		return GlobalRutas.Politicas.Base
	case Parametros:
		return GlobalRutas.Parametros.Base
	case Personas:
		return GlobalRutas.Personas.Base
	case Crud:
		return GlobalRutas.Crud.Base

	case Dashboard:
		return "/dashboard"
	default:
		return "#"
	}
}
