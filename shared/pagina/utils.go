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
	
	// Páginas legacy o pendientes de migrar a GlobalRutas
	case Regiones:
		return "/regiones/region"
	case Clinicas:
		return "/clinicas/clinica"
	case Dimensiones:
		return "/dimensiones/dimension"
	case Secciones:
		return "/secciones/seccion"
	case EnlaceClinica:
		return "/enlace_clinica/clinica"
	case Indicadores:
		return "/indicadores/indicador"
	case BancoPreguntas:
		return "/banco_preguntas/preguntas"
	case PlantillasSistema:
		return "/plantillas_sistema/plantillas"
	case PlantillasPreguntas:
		return "/plantillas_preguntas/preguntas"
	case GruposOpciones:
		return "/Grupos_Opciones/grupos_opciones"
	case GruposOpcionesDetalle:
		return "/Grupos_Opciones_Detalle/grupos_opciones_detalle"
	case CriteriosEvaluacion:
		return "/criterios_evaluacion/criterio"
	case SolicitudesReapertura:
		return "/reapertura/solicitudes"
	case MonitorReapertura:
		return "/reapertura/monitor"
	case Recomendaciones:
		return "/recomendaciones"
	case Aplicaciones:
		return "/aplicaciones"
	case Dashboard:
		return "/dashboard"
	case DashboardPlantilla:
		return "/dashboard_plantilla/configuracion"
	case DashboardOperativo:
		return "/dashboard-operativo"
	case DashboardEjecutivo:
		return "/dashboard-ejecutivo"
	case ReporteGeneral:
		return "/reporte-general"
	case Mapas:
		return "/mapas/resultados"
	default:
		return "#"
	}
}
