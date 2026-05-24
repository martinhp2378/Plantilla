package shared

import (
	"PlantillaInstitucional/shared/entorno"
	"bytes"
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
	fuenteTabla         = "Arial"
	fuenteBold          = "B"
	fuenteRegular       = ""
)

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
// Devuelve la URL pública del archivo almacenado en la librería de archivos
func GenerarReportePDF[T any](
	ctx context.Context,
	archivosMgr *archivos.ManejadorArchivos,
	datos []T,
	encabezados []string,
	rutasCampos []string,
	nombreArchivo string,
	titulo string,
	renderFuncs map[string]func(any) string,
) (string, error) {

	if len(datos) == 0 {
		return "", fmt.Errorf("lista vacía")
	}

	pdf := gofpdf.New("L", "mm", "A4", "")
	// Comentamos Gotham OTF temporalmente ya que causa panic en SplitLines por falta de soporte UTF8 nativo para OTF
	// pdf.AddUTF8Font(fuenteTabla, fuenteRegular, "Gotham-Book.otf")
	// pdf.AddUTF8Font(fuenteTabla, fuenteBold, "Gotham-Bold.otf")
	
	// Usamos Arial (built-in) para validar la estructura del reporte
	pdf.SetFont("Arial", fuenteRegular, 12)
	pdf.AddPage()

	pageWidth, _ := pdf.GetPageSize()
	leftMargin, _, rightMargin, _ := pdf.GetMargins()
	usableWidth := pageWidth - leftMargin - rightMargin

	// Bloque de la primera página (con imagen y título)
	env := entorno.EnvDeContexto(ctx)
	dirArchivosLocal := "./static"
	if env != nil {
		dirArchivosLocal = env.DirArchivosLocal
	}
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

	pdf.SetFont(fuenteTabla, fuenteBold, 18)
	tw := pdf.GetStringWidth(titulo)
	pdf.Text((pageWidth-tw)/2, yPos+imageH_mm+5, titulo)

	pdf.SetFont(fuenteTabla, fuenteRegular, 12)
	fecha := time.Now().Format("02/01/2006 15:04")
	pdf.Text((pageWidth-pdf.GetStringWidth("Generado el:"))/2, yPos+imageH_mm+15, "Generado el:")
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
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return "", fmt.Errorf("error generando PDF: %v", err)
	}

	// Guardar en la librería de archivos
	timestamp := time.Now().Local().Format("20060102-150405")
	storagePath := fmt.Sprintf("reportes/%s_%s.pdf", nombreArchivo, timestamp)

	archivoPublico := archivosMgr.ArchivoPublico(storagePath)
	reader := bytes.NewReader(buf.Bytes())

	if err := archivoPublico.Guardar(ctx, reader); err != nil {
		return "", fmt.Errorf("error guardando PDF en storage: %v", err)
	}

	url, err := archivoPublico.GetURL(ctx)
	if err != nil {
		return "", fmt.Errorf("error obteniendo URL del PDF: %v", err)
	}

	return url, nil
}

func CalcularAnchosColumnas[T any](pdf *gofpdf.Fpdf, encabezados []string, datos []T, rutasCampos []string, renderFuncs map[string]func(any) string, maxAncho float64) []float64 {
	pdf.SetFont(fuenteTabla, fuenteBold, 11)
	colAnchos := make([]float64, len(encabezados))
	for i, h := range encabezados {
		colAnchos[i] = pdf.GetStringWidth(h)
	}
	pdf.SetFont(fuenteTabla, fuenteRegular, 9)
	for _, item := range datos {
		for i, ruta := range rutasCampos {
			val := ObtenerValorCampo(item, ruta, renderFuncs)
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
	valores := make([]string, len(rutasCampos))
	altos := make([]float64, len(rutasCampos))
	for i, ruta := range rutasCampos {
		val := ObtenerValorCampo(item, ruta, renderFuncs)
		valores[i] = val
		if val == "" {
			altos[i] = altoLineaTexto
		} else {
			lines := pdf.SplitLines([]byte(val), colAnchos[i])
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
	pdf.SetX(startX)
	pdf.SetFillColor(86, 33, 47)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(fuenteTabla, fuenteBold, 10)
	for i, h := range encabezados {
		pdf.CellFormat(colAnchos[i], alturaEncabezado, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetFont(fuenteTabla, fuenteRegular, 9)
	pdf.SetTextColor(0, 0, 0)
}

func ObtenerValorCampo[T any](item T, ruta string, renderFuncs map[string]func(any) string) string {
	if f, ok := renderFuncs[ruta]; ok {
		return f(item)
	}
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	valRef := v
	for _, p := range strings.Split(ruta, ".") {
		if valRef.Kind() == reflect.Ptr && valRef.IsNil() {
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
