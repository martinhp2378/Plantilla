package shared

import (
	"testing"

	"github.com/issemym-repos/archivos"
)

func TestArchivoReporteSelectsPrivateStorageForSensitiveReports(t *testing.T) {
	manejador, err := archivos.NuevoManejador(&archivos.Config{Ambiente: archivos.AmbienteLocal, DirArchivosLocal: t.TempDir()})
	if err != nil {
		t.Fatalf("NuevoManejador() error = %v", err)
	}
	if archivoReporte(manejador, "reportes/personas.pdf", true).EsPublico() {
		t.Fatal("sensitive report must use private storage")
	}
	if !archivoReporte(manejador, "reportes/usuarios.pdf", false).EsPublico() {
		t.Fatal("non-sensitive report must retain public storage")
	}
}
