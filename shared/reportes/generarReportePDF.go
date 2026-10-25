package shared

import (
	"PlantillaInstitucional/shared/entorno"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/issemym-repos/archivos"
	"github.com/phpdave11/gofpdf"
)

// Constantes para el diseño del PDF
const (
	alturaEncabezado    = 10.0
	altoLineaTexto      = 5.0
	margenPrimeraPagina = 60.0
	margenOtrasPaginas  = 20.0
	margenInferiorMax   = 185.0
	margenLateralMin    = 20.0
	margenTextoCelda    = 6.0
	fuenteInstitucional = "Gotham"
	fuenteFallback      = "Arial"
	fuenteTabla         = fuenteInstitucional
	fuenteBold          = "B"
	fuenteRegular       = ""
)

// ConfigurarFuentesPDF registra la tipografía institucional Gotham si los archivos TTF
// existen en el directorio estático. Si no están disponibles o falla la carga, recurre
// de forma segura a Arial para garantizar que el reporte siempre se genere.
func ConfigurarFuentesPDF(pdf *gofpdf.Fpdf, dirArchivosLocal string) (fuente string, esUTF8 bool) {
	fontDir := filepath.Join(dirArchivosLocal, "public", "font", "Gotham")
	bookPath := filepath.Join(fontDir, "Gotham-Book.ttf")
	boldPath := filepath.Join(fontDir, "Gotham-Bold.ttf")

	if _, err := os.Stat(bookPath); err != nil {
		log.Printf("Advertencia: No se encontró Gotham-Book.ttf en %s, usando fallback %s: %v", fontDir, fuenteFallback, err)
		return fuenteFallback, false
	}
	if _, err := os.Stat(boldPath); err != nil {
		log.Printf("Advertencia: No se encontró Gotham-Bold.ttf en %s, usando fallback %s: %v", fontDir, fuenteFallback, err)
		return fuenteFallback, false
	}

	pdf.SetFontLocation(fontDir)
	pdf.AddUTF8Font(fuenteInstitucional, fuenteRegular, "Gotham-Book.ttf")
	pdf.AddUTF8Font(fuenteInstitucional, fuenteBold, "Gotham-Bold.ttf")

	if pdf.Err() {
		log.Printf("Advertencia: Error registrando fuentes Gotham UTF-8 (%v), usando fallback %s", pdf.Error(), fuenteFallback)
		pdf.ClearError()
		return fuenteFallback, false
	}

	return fuenteInstitucional, true
}

// esFuenteActualUTF8 detecta si la fuente activa en el documento PDF es UTF-8 nativa
// (como Gotham) o estándar de 8 bits (como Arial o Helvetica).
func esFuenteActualUTF8(pdf *gofpdf.Fpdf) bool {
	desc := pdf.GetFontDesc("", "")
	return desc.Ascent != 0 || desc.CapHeight != 0
}

// asegurarFuenteValida verifica que el PDF tenga una fuente válida activa;
// si no tiene ninguna o está en error, establece Arial como salvaguarda.
func asegurarFuenteValida(pdf *gofpdf.Fpdf) {
	if pdf.Err() {
		pdf.ClearError()
		pdf.SetFont(fuenteFallback, fuenteRegular, 12)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			pdf.ClearError()
			pdf.SetFont(fuenteFallback, fuenteRegular, 12)
		}
	}()
	_ = pdf.GetStringWidth(" ")
}

// obtenerTraductorTexto retorna la función para adaptar cadenas al conjunto de caracteres
// de la fuente activa: para fuentes UTF-8 nativas preserva el texto tal cual;
// para fuentes estándar de 8 bits (Arial) traduce a CP-1252.
func obtenerTraductorTexto(pdf *gofpdf.Fpdf, esUTF8 bool) func(string) string {
	if esUTF8 {
		return func(s string) string { return s }
	}
	return pdf.UnicodeTranslatorFromDescriptor("")
}

// dividirLineasTexto divide un texto según el ancho especificado respetando el tipo de fuente.
// Para fuentes UTF-8 utiliza SplitText; para fuentes de 8 bits utiliza SplitLines.
func dividirLineasTexto(pdf *gofpdf.Fpdf, esUTF8 bool, texto string, ancho float64) []string {
	if esUTF8 {
		return pdf.SplitText(texto, ancho)
	}
	lineasBytes := pdf.SplitLines([]byte(texto), ancho)
	lineas := make([]string, len(lineasBytes))
	for i, l := range lineasBytes {
		lineas[i] = string(l)
	}
	return lineas
}

// Función auxiliar para obtener las dimensiones de una imagen
func GetImageDimensions(imagePath string) (width, height int, err error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return 0, 0, fmt.Errorf("no se pudo abrir la imagen: %w", err)
	}
	defer func() { _ = file.Close() }()
	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, fmt.Errorf("no se pudo decodificar la configuración de la imagen: %w", err)
	}
	return img.Width, img.Height, nil
}

// Genera un reporte PDF genérico para cualquier tipo de datos T
// Returns an authenticated application URL for the private stored file.
func GenerarReportePDF[T any](
	ctx context.Context,
	archivosMgr *archivos.ManejadorArchivos,
	datos []T,
	encabezados []string,
	rutasCampos []string,
	nombreArchivo string,
	titulo string,
	renderFuncs map[string]func(any) string,
	_ bool,
) (string, error) {

	if len(datos) == 0 {
		return "", fmt.Errorf("lista vacía")
	}
	if err := validateReportRows(len(datos)); err != nil {
		return "", err
	}

	pdf := gofpdf.New("L", "mm", "A4", "")

	// Configuración de fuentes institucional (Gotham) con fallback seguro a Arial
	env := entorno.EnvDeContexto(ctx)
	dirArchivosLocal := "./static"
	if env != nil && env.DirArchivosLocal != "" {
		dirArchivosLocal = env.DirArchivosLocal
	}

	fuente, esUTF8 := ConfigurarFuentesPDF(pdf, dirArchivosLocal)
	tr := obtenerTraductorTexto(pdf, esUTF8)

	pdf.SetFont(fuente, fuenteRegular, 12)
	pdf.AddPage()

	pageWidth, _ := pdf.GetPageSize()
	leftMargin, _, rightMargin, _ := pdf.GetMargins()
	usableWidth := pageWidth - leftMargin - rightMargin

	// Bloque de la primera página (con imagen y título)
	logoPath := filepath.Join(dirArchivosLocal, "public", filepath.FromSlash("img/Encabezado.hj.ByN.png"))
	imgW_px, imgH_px, err := GetImageDimensions(logoPath)
	if err != nil {
		log.Printf("Advertencia: No se pudo obtener las dimensiones de la imagen, usando valores por defecto: %v", err)
		imgW_px, imgH_px = 100, 100
	}
	desiredImageWidth_mm := 50.0
	ratio := float64(imgH_px) / float64(imgW_px)
	imageH_mm := desiredImageWidth_mm * ratio
	xPos := leftMargin + (usableWidth-desiredImageWidth_mm)/2
	yPos := 10.0
	pdf.Image(logoPath, xPos, yPos, desiredImageWidth_mm, imageH_mm, false, "", 0, "")

	pdf.SetFont(fuente, fuenteBold, 18)
	tituloTr := tr(titulo)
	tw := pdf.GetStringWidth(tituloTr)
	pdf.Text((pageWidth-tw)/2, yPos+imageH_mm+5, tituloTr)

	pdf.SetFont(fuente, fuenteRegular, 12)
	fecha := time.Now().Format("02/01/2006 15:04")
	generadoEl := tr("Generado el:")
	pdf.Text((pageWidth-pdf.GetStringWidth(generadoEl))/2, yPos+imageH_mm+15, generadoEl)
	pdf.Text((pageWidth-pdf.GetStringWidth(fecha))/2, yPos+imageH_mm+22, fecha)

	colAnchos := CalcularAnchosColumnas(pdf, encabezados, datos, rutasCampos, renderFuncs, usableWidth)
	startX := leftMargin + (usableWidth-Suma(colAnchos))/2
	pdf.SetY(margenPrimeraPagina)
	DibujarEncabezados(pdf, encabezados, colAnchos, startX)

	// Dibujar filas una por una
	for _, item := range datos {
		valores, alturaFila := PrepararFila(pdf, item, rutasCampos, colAnchos, renderFuncs)

		// Salto de página
		if pdf.GetY()+alturaFila > margenInferiorMax {
			pdf.AddPage()
			pdf.SetY(margenOtrasPaginas)
			DibujarEncabezados(pdf, encabezados, colAnchos, startX)
		}

		x, y := startX, pdf.GetY()
		for i, val := range valores {
			pdf.SetXY(x, y)
			pdf.MultiCell(colAnchos[i], altoLineaTexto, val, "", "L", false)
			x += colAnchos[i]
		}

		x = startX
		for _, ancho := range colAnchos {
			pdf.Rect(x, y, ancho, alturaFila, "D")
			x += ancho
		}
		pdf.SetY(y + alturaFila)
	}

	// Generar el PDF en memoria
	buf := newLimitedBuffer(MaxReportSize)
	if err := pdf.Output(buf); err != nil {
		return "", fmt.Errorf("error generando PDF: %v", err)
	}

	url, err := StorePrivateReport(ctx, archivosMgr, buf.Bytes(), nombreArchivo, ".pdf")
	if err != nil {
		return "", fmt.Errorf("error guardando PDF en storage: %v", err)
	}
	return url, nil
}

func CalcularAnchosColumnas[T any](pdf *gofpdf.Fpdf, encabezados []string, datos []T, rutasCampos []string, renderFuncs map[string]func(any) string, maxAncho float64) []float64 {
	asegurarFuenteValida(pdf)
	esUTF8 := esFuenteActualUTF8(pdf)
	tr := obtenerTraductorTexto(pdf, esUTF8)
	pdf.SetFont("", fuenteBold, 11)
	colAnchos := make([]float64, len(encabezados))
	for i, h := range encabezados {
		colAnchos[i] = pdf.GetStringWidth(tr(h))
	}
	pdf.SetFont("", fuenteRegular, 9)
	for _, item := range datos {
		for i, ruta := range rutasCampos {
			val := tr(ObtenerValorCampo(item, ruta, renderFuncs))
			w := pdf.GetStringWidth(val)
			if w > colAnchos[i] {
				colAnchos[i] = w
			}
		}
	}
	for i := range colAnchos {
		colAnchos[i] += margenTextoCelda
		if colAnchos[i] < margenLateralMin {
			colAnchos[i] = margenLateralMin
		}
	}
	total := Suma(colAnchos)
	if total > maxAncho {
		factor := maxAncho / total
		for i := range colAnchos {
			colAnchos[i] *= factor
		}
	}
	return colAnchos
}

func PrepararFila[T any](pdf *gofpdf.Fpdf, item T, rutasCampos []string, colAnchos []float64, renderFuncs map[string]func(any) string) ([]string, float64) {
	asegurarFuenteValida(pdf)
	esUTF8 := esFuenteActualUTF8(pdf)
	tr := obtenerTraductorTexto(pdf, esUTF8)
	valores := make([]string, len(rutasCampos))
	altos := make([]float64, len(rutasCampos))
	for i, ruta := range rutasCampos {
		val := tr(ObtenerValorCampo(item, ruta, renderFuncs))
		valores[i] = val
		if val == "" {
			altos[i] = altoLineaTexto
		} else {
			lines := dividirLineasTexto(pdf, esUTF8, val, colAnchos[i])
			altos[i] = float64(len(lines)) * altoLineaTexto
		}
	}
	alturaFila := altoLineaTexto
	for _, h := range altos {
		if h > alturaFila {
			alturaFila = h
		}
	}
	return valores, alturaFila
}

func DibujarEncabezados(pdf *gofpdf.Fpdf, encabezados []string, colAnchos []float64, startX float64) {
	asegurarFuenteValida(pdf)
	esUTF8 := esFuenteActualUTF8(pdf)
	tr := obtenerTraductorTexto(pdf, esUTF8)
	pdf.SetX(startX)
	pdf.SetFillColor(86, 33, 47)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("", fuenteBold, 10)
	for i, h := range encabezados {
		pdf.CellFormat(colAnchos[i], alturaEncabezado, tr(h), "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetFont("", fuenteRegular, 9)
	pdf.SetTextColor(0, 0, 0)
}

func ObtenerValorCampo[T any](item T, ruta string, renderFuncs map[string]func(any) string) string {
	if f, ok := renderFuncs[ruta]; ok {
		return f(item)
	}
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	valRef := v
	for _, p := range strings.Split(ruta, ".") {
		if valRef.Kind() == reflect.Pointer && valRef.IsNil() {
			return ""
		}
		valRef = reflect.Indirect(valRef).FieldByName(p)
	}
	if valRef.IsValid() && valRef.CanInterface() {
		return fmt.Sprintf("%v", valRef.Interface())
	}
	return ""
}

func Suma(floats []float64) float64 {
	total := 0.0
	for _, v := range floats {
		total += v
	}
	return total
}
