package shared

import (
	"PlantillaInstitucional/shared/entorno"
	"bytes"
	"context"
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/issemym-repos/archivos"
	"github.com/xuri/excelize/v2"
)

// GenerarReporteExcel crea un archivo Excel y lo almacena en la librería de archivos
// Devuelve la URL pública del archivo
func GenerarReporteExcel[T any](
	ctx context.Context,
	archivosMgr *archivos.ManejadorArchivos,
	datos []T,
	encabezados []string,
	rutasCampos []string,
	nombreArchivo string,
	titulo string,
	renderFuncs map[string]func(any) string,
	privado bool,
) (string, error) {
	if len(datos) == 0 {
		return "", fmt.Errorf("lista vacía")
	}

	f := excelize.NewFile()
	sheet := "Hoja1"
	_ = f.SetSheetName("Sheet1", sheet)

	const headerRow = 12

	// Estilo para el Título
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Font: &excelize.Font{
			Bold:  true,
			Size:  14,
			Color: "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"56212F"},
			Pattern: 1,
		},
	})

	// Estilo para la Fecha
	fechaStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Font: &excelize.Font{
			Italic: true,
			Size:   12,
			Color:  "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"56212F"},
			Pattern: 1,
		},
	})

	// Estilo para Encabezados
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"56212F"},
			Pattern: 1,
		},
	})

	lastCol, _ := excelize.ColumnNumberToName(len(encabezados))

	for i := 1; i <= 7; i++ {
		_ = f.SetRowHeight(sheet, i, 15)
	}

	imageEndCell := fmt.Sprintf("%s7", lastCol)
	if err := f.MergeCell(sheet, "A1", imageEndCell); err != nil {
		log.Printf("Advertencia: No se pudo fusionar las celdas para la imagen: %v", err)
	}

	env := entorno.EnvDeContexto(ctx)
	dirArchivosLocal := "./static"
	if env != nil {
		dirArchivosLocal = env.DirArchivosLocal
	}
	logoPath := filepath.Join(dirArchivosLocal, "public", filepath.FromSlash("img/Encabezado.hj.ByN.png"))
	opts := &excelize.GraphicOptions{
		Positioning: "oneCell",
		OffsetX:     5,
		OffsetY:     5,
	}
	if err := f.AddPicture(sheet, "A1", logoPath, opts); err != nil {
		log.Printf("Advertencia: No se pudo agregar la imagen al Excel: %v", err)
	}

	titleEndCell := fmt.Sprintf("%s9", lastCol)
	_ = f.SetCellValue(sheet, "A9", titulo)
	_ = f.MergeCell(sheet, "A9", titleEndCell)
	_ = f.SetCellStyle(sheet, "A9", titleEndCell, titleStyle)

	fechaStr := time.Now().Format("02/01/2006 15:04")
	fechaEndCell := fmt.Sprintf("%s10", lastCol)
	_ = f.SetCellValue(sheet, "A10", "Generado el: "+fechaStr)
	_ = f.MergeCell(sheet, "A10", fechaEndCell)
	_ = f.SetCellStyle(sheet, "A10", fechaEndCell, fechaStyle)

	// Escribir Encabezados
	for i, h := range encabezados {
		colName, err := excelize.ColumnNumberToName(i + 1)
		if err != nil {
			log.Fatalf("Error al obtener nombre de columna: %v", err)
		}

		cell := fmt.Sprintf("%s%d", colName, headerRow)

		_ = f.SetCellValue(sheet, cell, h)
		_ = f.SetColWidth(sheet, colName, colName, 20)
		_ = f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	// Escribir Datos
	for rowIdx, item := range datos {
		v := reflect.ValueOf(item)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		currentRow := headerRow + 1 + rowIdx

		for colIdx, ruta := range rutasCampos {
			var val string

			colName, err := excelize.ColumnNumberToName(colIdx + 1)
			if err != nil {
				log.Printf("Advertencia: Error al obtener nombre de columna en fila %d: %v", currentRow, err)
				continue
			}
			cell := fmt.Sprintf("%s%d", colName, currentRow)

			if fRender, ok := renderFuncs[ruta]; ok {
				val = fRender(item)
			} else {
				valRef := v
				for _, p := range strings.Split(ruta, ".") {
					if valRef.Kind() == reflect.Ptr && valRef.IsNil() {
						valRef = reflect.Zero(valRef.Type().Elem())
						break
					}
					valRef = reflect.Indirect(valRef).FieldByName(p)
				}

				if valRef.IsValid() && valRef.CanInterface() {
					val = fmt.Sprintf("%v", valRef.Interface())
				} else {
					val = ""
				}
			}

			_ = f.SetCellValue(sheet, cell, val)
		}
	}

	// Aplicar Filtros
	topLeft := fmt.Sprintf("A%d", headerRow)
	bottomRight := fmt.Sprintf("%s%d", lastCol, headerRow+len(datos))
	if err := f.AutoFilter(sheet, topLeft+":"+bottomRight, nil); err != nil {
		return "", fmt.Errorf("error aplicando autofiltros: %v", err)
	}

	// Guardar en memoria
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return "", fmt.Errorf("error generando Excel en memoria: %v", err)
	}

	// Guardar en la librería de archivos
	timestamp := time.Now().Local().Format("20060102-150405")
	storagePath := fmt.Sprintf("reportes/%s_%s.xlsx", nombreArchivo, timestamp)

	archivo := archivoReporte(archivosMgr, storagePath, privado)
	reader := bytes.NewReader(buf.Bytes())

	if err := archivo.Guardar(ctx, reader); err != nil {
		return "", fmt.Errorf("error guardando Excel en storage: %v", err)
	}

	url, err := archivo.GetURL(ctx)
	if err != nil {
		return "", fmt.Errorf("error obteniendo URL del Excel: %v", err)
	}

	return url, nil
}
