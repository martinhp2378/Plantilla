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
	pdf.SetFont(fuenteFallback, fuenteRegular, 12)
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

func TestPDFAccentsWithGotham(t *testing.T) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	fuente, esUTF8 := ConfigurarFuentesPDF(pdf, "../../static")
	if !esUTF8 || fuente != fuenteInstitucional {
		t.Fatalf("expected Gotham UTF-8 font to be configured, got fuente=%s, esUTF8=%v", fuente, esUTF8)
	}
	pdf.SetFont(fuente, fuenteRegular, 12)
	pdf.AddPage()

	headers := []string{"Nombre", "Código", "Descripción", "Estatus"}
	items := []sampleReportItem{
		{Nombre: "García", Codigo: "COD-01", Descripcion: "Descripción válida con acentos: áéíóú ñ ÁÉÍÓÚ Ñ"},
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
	if valores[0] != "García" {
		t.Errorf("expected native UTF-8 value %q, got %q", "García", valores[0])
	}
	if valores[2] != "Descripción válida con acentos: áéíóú ñ ÁÉÍÓÚ Ñ" {
		t.Errorf("expected native UTF-8 value, got %q", valores[2])
	}

	DibujarEncabezados(pdf, headers, colAnchos, 10.0)

	// Verify complete PDF buffer generation
	buf := newLimitedBuffer(MaxReportSize)
	if err := pdf.Output(buf); err != nil {
		t.Fatalf("pdf.Output failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatalf("expected non-empty PDF buffer")
	}
}

func TestConfigurarFuentesPDFFallback(t *testing.T) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	fuente, esUTF8 := ConfigurarFuentesPDF(pdf, "/non/existent/path")
	if fuente != fuenteFallback {
		t.Errorf("expected fallback font %s, got %s", fuenteFallback, fuente)
	}
	if esUTF8 {
		t.Errorf("expected esUTF8 to be false on fallback")
	}
}

func TestAsegurarFuenteValida(t *testing.T) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	// Do not set any font; calling PrepararFila should not panic
	headers := []string{"Nombre"}
	items := []sampleReportItem{{Nombre: "Prueba"}}
	fields := []string{"Nombre"}

	colAnchos := CalcularAnchosColumnas(pdf, headers, items, fields, nil, 200.0)
	valores, alt := PrepararFila(pdf, items[0], fields, colAnchos, nil)
	if alt <= 0 || len(valores) == 0 {
		t.Fatalf("expected valid row calculation without panic")
	}
}

func TestGothamFontLoadingAndSplitting(t *testing.T) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	fontDir := "../../static/public/font/Gotham"
	pdf.SetFontLocation(fontDir)
	pdf.AddUTF8Font("Gotham", "", "Gotham-Book.ttf")
	pdf.AddUTF8Font("Gotham", "B", "Gotham-Bold.ttf")

	if err := pdf.Error(); err != nil {
		t.Fatalf("AddUTF8Font failed: %v", err)
	}

	pdf.SetFont("Gotham", "", 12)
	pdf.AddPage()

	// Test accents and text width
	text := "Reporte Institucional con acentos: á, é, í, ó, ú, ñ, Á, É, Í, Ó, Ú, Ñ"
	w := pdf.GetStringWidth(text)
	if w <= 0 {
		t.Fatalf("expected positive width for UTF-8 text, got %f", w)
	}

	// Test SplitText vs SplitLines
	lines := pdf.SplitText(text, 50.0)
	if len(lines) <= 1 {
		t.Errorf("expected text to wrap into multiple lines, got %d lines: %v", len(lines), lines)
	}
	splitLinesResult := pdf.SplitLines([]byte(text), 50.0)
	if len(splitLinesResult) == 0 {
		t.Errorf("expected splitLinesResult to have lines")
	}

	pdf.SetXY(10, 20)
	pdf.MultiCell(50.0, 5.0, text, "", "L", false)

	descGotham := pdf.GetFontDesc("", "")
	t.Logf("Gotham Desc: %+v", descGotham)

	pdfArial := gofpdf.New("L", "mm", "A4", "")
	pdfArial.SetFont("Arial", "", 12)
	descArial := pdfArial.GetFontDesc("", "")
	t.Logf("Arial Desc: %+v", descArial)

	// Verify SetFont with empty family preserves current font family
	pdf.SetFont("", "B", 10)
	descBold := pdf.GetFontDesc("", "")
	if descBold.Ascent == 0 {
		t.Fatalf("expected Gotham Bold to preserve font family, got ascent 0")
	}

	// Test Output buffer generation
	var buf []byte
	pdf.SetFont("Gotham", "B", 14)
	pdf.Cell(0, 10, "Título en Negrita")
	if err := pdf.Output(&testBuffer{buf: &buf}); err != nil {
		t.Fatalf("pdf.Output failed: %v", err)
	}
	if len(buf) == 0 {
		t.Fatalf("expected non-empty PDF output buffer")
	}
}

type testBuffer struct {
	buf *[]byte
}

func (b *testBuffer) Write(p []byte) (n int, err error) {
	*b.buf = append(*b.buf, p...)
	return len(p), nil
}
