package shared

import (
	"PlantillaInstitucional/shared/entorno"
	"PlantillaInstitucional/shared/qrcode"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/issemym-repos/archivos"
	"github.com/johnfercher/maroto/pkg/color"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
)

// GenerarReportePDFMaroto genera un reporte PDF usando la librería Maroto
// Devuelve la URL pública del archivo almacenado en la librería de archivos
func GenerarReportePDFMaroto[T any](
	ctx context.Context,
	archivosMgr *archivos.ManejadorArchivos,
	datos []T,
	encabezados []string,
	rutasCampos []string,
	nombreArchivo string,
	titulo string,
	generadoPor string,
	renderFuncs map[string]func(any) string,
) (string, error) {

	if len(datos) == 0 {
		return "", fmt.Errorf("lista vacía")
	}

	m := pdf.NewMaroto(consts.Landscape, consts.A4)
	m.SetPageMargins(10, 10, 10)

	// --- Encabezado ---
	m.RegisterHeader(func() {
		m.Row(25, func() {
			m.Col(12, func() {
				env := entorno.EnvDeContexto(ctx)
				dirArchivosLocal := "./static"
				if env != nil {
					dirArchivosLocal = env.DirArchivosLocal
				}
				logoPath := filepath.Join(dirArchivosLocal, "public", filepath.FromSlash("img/Encabezado.hj.ByN.png"))
				if _, err := os.Stat(logoPath); err == nil {
					_ = m.FileImage(logoPath, props.Rect{
						Center:  true,
						Percent: 100,
					})
				}
			})
		})

		m.Row(15, func() {
			m.Col(9, func() {
				m.Text(titulo, props.Text{
					Top:   3,
					Size:  14,
					Style: consts.Bold,
					Align: consts.Center,
				})
			})

			m.Col(3, func() {
				m.Text("Generado:", props.Text{
					Size:  8,
					Style: consts.Bold,
					Align: consts.Right,
					Top:   3,
				})
				m.Text(time.Now().Format("02/01/2006 15:04"), props.Text{
					Size:  8,
					Align: consts.Right,
					Top:   8,
				})
			})
		})
	})

	// --- Tabla ---
	var tableRows [][]string
	headerRow := encabezados

	for _, item := range datos {
		var row []string
		for _, ruta := range rutasCampos {
			val := ObtenerValorCampo(item, ruta, renderFuncs)
			row = append(row, val)
		}
		tableRows = append(tableRows, row)
	}

	colWidths := calculateColumnWidths(headerRow, tableRows)

	m.TableList(headerRow, tableRows, props.TableList{
		HeaderProp: props.TableListContent{
			Size:      8,
			GridSizes: colWidths,
			Style:     consts.Bold,
			Color:     color.NewBlack(),
			Family:    consts.Arial,
		},
		ContentProp: props.TableListContent{
			Size:      7,
			GridSizes: colWidths,
			Family:    consts.Arial,
		},
		Align:                consts.Center,
		HeaderContentSpace:   1,
		AlternatedBackground: &color.Color{Red: 245, Green: 245, Blue: 245},
	})

	// --- Footer ---
	m.RegisterFooter(func() {
		m.Row(20, func() {
			m.Col(10, func() {
				m.Text(fmt.Sprintf("Página %d", m.GetCurrentPage()), props.Text{
					Align: consts.Left,
					Size:  8,
					Top:   15,
				})
			})
			m.Col(2, func() {
				qrContent := fmt.Sprintf("Generado por: %s\nFecha: %s", generadoPor, time.Now().Format(time.RFC3339))
				qrBytes, err := qrcode.GenerateQR(qrContent, 128)
				if err == nil {
					qrBase64 := base64.StdEncoding.EncodeToString(qrBytes)
					_ = m.Base64Image(qrBase64, consts.Png, props.Rect{
						Center:  false,
						Percent: 100,
						Left:    0,
					})
				}
			})
		})
	})

	// Generar en archivo temporal y luego leer
	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s_%d.pdf", nombreArchivo, time.Now().UnixNano()))
	if err := m.OutputFileAndClose(tempFile); err != nil {
		return "", fmt.Errorf("error generando PDF Maroto: %v", err)
	}

	// Leer el archivo temporal
	data, err := os.ReadFile(tempFile)
	if err != nil {
		return "", fmt.Errorf("error leyendo PDF temporal: %v", err)
	}

	// Eliminar archivo temporal
	if err := os.Remove(tempFile); err != nil {
		log.Printf("Advertencia: no se pudo eliminar archivo temporal: %v", err)
	}

	// Guardar en la librería de archivos
	timestamp := time.Now().Local().Format("20060102-150405")
	storagePath := fmt.Sprintf("reportes/%s_%s_maroto.pdf", nombreArchivo, timestamp)

	archivoPublico := archivosMgr.ArchivoPublico(storagePath)
	reader := bytes.NewReader(data)

	if err := archivoPublico.Guardar(ctx, reader); err != nil {
		return "", fmt.Errorf("error guardando PDF Maroto en storage: %v", err)
	}

	url, err := archivoPublico.GetURL(ctx)
	if err != nil {
		return "", fmt.Errorf("error obteniendo URL del PDF Maroto: %v", err)
	}

	return url, nil
}

func calculateColumnWidths(headers []string, rows [][]string) []uint {
	numCols := len(headers)
	if numCols == 0 {
		return []uint{}
	}

	maxLengths := make([]int, numCols)
	for i, h := range headers {
		maxLengths[i] = len(h)
	}

	for _, row := range rows {
		for i, val := range row {
			if i < numCols {
				if len(val) > maxLengths[i] {
					maxLengths[i] = len(val)
				}
			}
		}
	}

	totalChars := 0
	for _, l := range maxLengths {
		totalChars += l
	}

	widths := make([]uint, numCols)
	usedGrid := 0

	for i := 0; i < numCols; i++ {
		ratio := float64(maxLengths[i]) / float64(totalChars)
		gridUnits := uint(ratio * 12.0)
		if gridUnits < 1 {
			gridUnits = 1
		}
		widths[i] = gridUnits
		usedGrid += int(gridUnits)
	}

	if usedGrid > 12 {
		diff := usedGrid - 12
		for diff > 0 {
			maxIdx := -1
			maxVal := uint(0)
			for i, w := range widths {
				if w > 1 && w > maxVal {
					maxVal = w
					maxIdx = i
				}
			}
			if maxIdx != -1 {
				widths[maxIdx]--
				diff--
			} else {
				break
			}
		}
	} else if usedGrid < 12 {
		diff := 12 - usedGrid
		for diff > 0 {
			maxIdx := -1
			maxVal := -1
			for i, l := range maxLengths {
				if l > maxVal {
					maxVal = l
					maxIdx = i
				}
			}
			if maxIdx == -1 {
				maxIdx = 0
			}

			widths[maxIdx]++
			diff--
			maxLengths[maxIdx] = 0
		}
	}

	return widths
}
