package shared

import (
	"testing"

	"github.com/phpdave11/gofpdf"
)

type sampleReportItem struct {
	Nombre      string
	Codigo      string
	Descripcion string
}

func TestPDFAccentsTranslation(t *testing.T) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetFont(fuenteTabla, fuenteRegular, 12)
	pdf.AddPage()

	headers := []string{"Nombre", "Código", "Descripción", "Estatus"}
	items := []sampleReportItem{
		{Nombre: "García", Codigo: "COD-01", Descripcion: "Descripción válida con acentos: áéíóú ñ"},
	}
	fields := []string{"Nombre", "Codigo", "Descripcion"}

	colAnchos := CalcularAnchosColumnas(pdf, headers, items, fields, nil, 200.0)
	if len(colAnchos) != len(headers) {
		t.Fatalf("expected %d colAnchos, got %d", len(headers), len(colAnchos))
	}

	valores, alt := PrepararFila(pdf, items[0], fields, colAnchos, nil)
	if alt <= 0 {
		t.Fatalf("expected positive row height, got %f", alt)
	}
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	if valores[0] != tr("García") {
		t.Errorf("expected translated value %q, got %q", tr("García"), valores[0])
	}
	if valores[2] != tr("Descripción válida con acentos: áéíóú ñ") {
		t.Errorf("expected translated value %q, got %q", tr("Descripción válida con acentos: áéíóú ñ"), valores[2])
	}

	DibujarEncabezados(pdf, headers, colAnchos, 10.0)
}
