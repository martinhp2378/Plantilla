package pagina

import (
	"fmt"
)

// Rutas es una estructura que centraliza todas las rutas del sistema.
type Rutas struct {
	Prefix string
	Base   string
	Tabla  string
	Nuevo  string
	Crear  string
	Editar string
	Update string
	Borrar string
	Export      RutasExport
	CargaMasiva RutasCargaMasiva
}

type RutasCargaMasiva struct {
	Modal     string
	Subir     string
	Plantilla string
}

type RutasExport struct {
	PDF          string
	Excel        string
	Notificacion string
}

// GlobalRutas contiene las rutas base para cada módulo.
var GlobalRutas = struct {
	Login      Rutas
	Dashboard  Rutas
	Usuarios   Rutas
	Perfiles   Rutas
	Politicas  Rutas
	Parametros Rutas
	Personas   Rutas
	Crud       Rutas
	MiPerfil   Rutas
}{
	Login:      NuevaRuta("/login"),
	Dashboard:  NuevaRuta("/dashboard"),
	Usuarios:   NuevaRuta("/configuracion/usuarios"),
	Perfiles:   NuevaRuta("/configuracion/perfiles"),
	Politicas:  NuevaRuta("/configuracion/politicas"),
	Parametros: NuevaRuta("/configuracion/parametros"),
	Personas:   NuevaRuta("/institucion/personas"),
	Crud:       NuevaRuta("/crud"),
	MiPerfil:   NuevaRuta("/perfil"),
}

// NuevaRuta genera la estructura estándar de rutas para un módulo a partir de su path base.
func NuevaRuta(base string) Rutas {
	// Limpieza para asegurar consistencia
	if base == "" || base == "/" {
		return Rutas{Prefix: "/", Base: "/"}
	}

	return Rutas{
		Prefix: base,
		Base:   base + "/",
		Tabla:  base + "/tabla",
		Nuevo:  base + "/modales/nuevo",
		Crear:  base + "/",
		Editar: base + "/modales/editar/%s",
		Update: base + "/%s",
		Borrar: base + "/%s",
		Export: RutasExport{
			PDF:          base + "/reportes/pdf",
			Excel:        base + "/reportes/excel",
			Notificacion: base + "/reportes/notificacion-exito",
		},
		CargaMasiva: RutasCargaMasiva{
			Modal:     base + "/modales/carga-masiva",
			Subir:     base + "/carga-masiva",
			Plantilla: base + "/carga-masiva/plantilla",
		},
	}
}

// GetEditar devuelve la ruta formateada para editar un registro.
func (r Rutas) GetEditar(id any) string {
	if id == nil {
		return "#"
	}
	return fmt.Sprintf(r.Editar, id)
}

// GetUpdate devuelve la ruta formateada para actualizar un registro.
func (r Rutas) GetUpdate(id any) string {
	if id == nil {
		return "#"
	}
	return fmt.Sprintf(r.Update, id)
}

// GetBorrar devuelve la ruta formateada para eliminar un registro.
func (r Rutas) GetBorrar(id any) string {
	if id == nil {
		return "#"
	}
	return fmt.Sprintf(r.Borrar, id)
}
