package shared

import "github.com/issemym-repos/archivos"

func archivoReporte(manejador *archivos.ManejadorArchivos, path string, privado bool) *archivos.Archivo {
	if privado {
		return manejador.ArchivoPrivado(path)
	}
	return manejador.ArchivoPublico(path)
}
