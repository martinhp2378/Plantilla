package cargamasiva

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ErrorFila describe un error puntual de validación en una fila y columna del archivo.
type ErrorFila struct {
	Fila    int    `json:"fila"`
	Columna string `json:"columna,omitempty"`
	Valor   string `json:"valor,omitempty"`
	Mensaje string `json:"mensaje"`
}

func (e ErrorFila) Error() string {
	if e.Columna != "" {
		return fmt.Sprintf("Línea %d, columna '%s' (valor '%s'): %s", e.Fila, e.Columna, e.Valor, e.Mensaje)
	}
	return fmt.Sprintf("Línea %d: %s", e.Fila, e.Mensaje)
}

// ResultadoCarga contiene el reporte detallado tras procesar el archivo.
type ResultadoCarga[T any] struct {
	TotalLeidos  int         `json:"total_leidos"`
	TotalValidos int         `json:"total_validos"`
	TotalErrores int         `json:"total_errores"`
	Registros    []T         `json:"registros"`
	Errores      []ErrorFila `json:"errores"`
}

// EsExitoso indica si se procesaron registros y no hubo ningún error.
func (r *ResultadoCarga[T]) EsExitoso() bool {
	return len(r.Errores) == 0 && r.TotalValidos > 0
}

// DefinicionColumna especifica las reglas de parseo y validación de una columna.
type DefinicionColumna[T any] struct {
	Nombre      string
	Requerido   bool
	Descripcion string
	Ejemplo     string
	Parser      func(valor string, entidad *T) error
}

// ConfigCarga encapsula las opciones para procesar el CSV.
type ConfigCarga[T any] struct {
	Separador         rune
	MaxFilas          int
	ColumnasEsperadas []DefinicionColumna[T]
}

// ProcesarCSV lee, valida encabezados y contenido fila por fila.
func ProcesarCSV[T any](r io.Reader, cfg ConfigCarga[T]) (*ResultadoCarga[T], error) {
	if len(cfg.ColumnasEsperadas) == 0 {
		return nil, errors.New("la configuración no contiene columnas esperadas")
	}

	rawBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error al leer contenido del archivo: %w", err)
	}

	// Quitar BOM UTF-8 si existe (típico de exportaciones de Excel)
	rawBytes = bytes.TrimPrefix(rawBytes, []byte("\xef\xbb\xbf"))

	sep := cfg.Separador
	if sep == 0 {
		sep = ','
	}

	csvReader := csv.NewReader(bytes.NewReader(rawBytes))
	csvReader.Comma = sep
	csvReader.FieldsPerRecord = -1 // Permite validar número de campos manualmente con mejor mensaje
	csvReader.TrimLeadingSpace = true

	// 1. Validar Encabezado
	header, err := csvReader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, errors.New("el archivo CSV está vacío")
		}
		return nil, fmt.Errorf("error al leer encabezado del CSV: %w", err)
	}

	headerMap := make(map[string]int)
	for idx, col := range header {
		normalized := strings.ToLower(strings.TrimSpace(col))
		if normalized != "" {
			headerMap[normalized] = idx
		}
	}

	// Validar que todas las columnas requeridas existan en el encabezado
	var columnasFaltantes []string
	for _, col := range cfg.ColumnasEsperadas {
		colKey := strings.ToLower(strings.TrimSpace(col.Nombre))
		if _, ok := headerMap[colKey]; !ok && col.Requerido {
			columnasFaltantes = append(columnasFaltantes, col.Nombre)
		}
	}

	if len(columnasFaltantes) > 0 {
		return nil, fmt.Errorf("faltan columnas obligatorias en el encabezado del archivo: %s", strings.Join(columnasFaltantes, ", "))
	}

	resultado := &ResultadoCarga[T]{
		Registros: make([]T, 0),
		Errores:   make([]ErrorFila, 0),
	}

	maxFilas := cfg.MaxFilas
	if maxFilas <= 0 {
		maxFilas = 10000 // Límite de seguridad por defecto
	}

	numFila := 1 // La fila 1 fue el encabezado
	for {
		record, err := csvReader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			numFila++
			resultado.Errores = append(resultado.Errores, ErrorFila{
				Fila:    numFila,
				Mensaje: fmt.Sprintf("formato de fila inválido: %v", err),
			})
			continue
		}

		numFila++
		resultado.TotalLeidos++

		if resultado.TotalLeidos > maxFilas {
			return nil, fmt.Errorf("el archivo excede el límite máximo permitido de %d filas", maxFilas)
		}

		// Validar si la fila está completamente vacía y saltarla
		if filaEstaVacia(record) {
			resultado.TotalLeidos-- // No contar fila vacía
			continue
		}

		entidad, erroresFila := procesarFila(record, headerMap, cfg.ColumnasEsperadas, numFila)
		if len(erroresFila) > 0 {
			resultado.Errores = append(resultado.Errores, erroresFila...)
			resultado.TotalErrores++
		} else {
			resultado.Registros = append(resultado.Registros, entidad)
			resultado.TotalValidos++
		}
	}

	return resultado, nil
}

func filaEstaVacia(record []string) bool {
	for _, val := range record {
		if strings.TrimSpace(val) != "" {
			return false
		}
	}
	return true
}

func procesarFila[T any](record []string, headerMap map[string]int, columnas []DefinicionColumna[T], numFila int) (T, []ErrorFila) {
	var entidad T
	var erroresFila []ErrorFila

	for _, colDef := range columnas {
		colKey := strings.ToLower(strings.TrimSpace(colDef.Nombre))
		idx, existe := headerMap[colKey]

		var valor string
		if existe && idx < len(record) {
			valor = strings.TrimSpace(record[idx])
		}

		if valor == "" {
			if colDef.Requerido {
				erroresFila = append(erroresFila, ErrorFila{
					Fila:    numFila,
					Columna: colDef.Nombre,
					Valor:   "",
					Mensaje: "Campo obligatorio no proporcionado",
				})
			}
			continue
		}

		if colDef.Parser != nil {
			if parseErr := colDef.Parser(valor, &entidad); parseErr != nil {
				erroresFila = append(erroresFila, ErrorFila{
					Fila:    numFila,
					Columna: colDef.Nombre,
					Valor:   valor,
					Mensaje: parseErr.Error(),
				})
			}
		}
	}

	return entidad, erroresFila
}

// GenerarPlantillaCSV genera los bytes de un archivo CSV con el encabezado y una fila de ejemplo.
func GenerarPlantillaCSV[T any](cfg ConfigCarga[T]) ([]byte, error) {
	if len(cfg.ColumnasEsperadas) == 0 {
		return nil, errors.New("no hay columnas definidas")
	}

	var buf bytes.Buffer
	sep := cfg.Separador
	if sep == 0 {
		sep = ','
	}

	writer := csv.NewWriter(&buf)
	writer.Comma = sep

	headers := make([]string, len(cfg.ColumnasEsperadas))
	ejemplos := make([]string, len(cfg.ColumnasEsperadas))

	for i, col := range cfg.ColumnasEsperadas {
		headers[i] = col.Nombre
		ejemplos[i] = col.Ejemplo
	}

	if err := writer.Write(headers); err != nil {
		return nil, err
	}
	if err := writer.Write(ejemplos); err != nil {
		return nil, err
	}
	writer.Flush()

	return buf.Bytes(), writer.Error()
}

// --- Parsers y Validadores Auxiliares Reutilizables ---

// ParseEntero valida y convierte a int dentro de un rango opcional.
func ParseEntero(valor string, min, max int) (int, error) {
	n, err := strconv.Atoi(valor)
	if err != nil {
		return 0, errors.New("debe ser un número entero válido")
	}
	if min != 0 || max != 0 {
		if n < min || n > max {
			return 0, fmt.Errorf("debe estar entre %d y %d", min, max)
		}
	}
	return n, nil
}

// ParseDecimal valida y convierte a float64 dentro de un rango opcional.
func ParseDecimal(valor string, min, max float64) (float64, error) {
	f, err := strconv.ParseFloat(valor, 64)
	if err != nil {
		return 0, errors.New("debe ser un valor numérico decimal válido")
	}
	if min != 0 || max != 0 {
		if f < min || f > max {
			return 0, fmt.Errorf("debe estar entre %.2f y %.2f", min, max)
		}
	}
	return f, nil
}

// ParseBooleano convierte valores booleanos en español o estándar.
func ParseBooleano(valor string) (bool, error) {
	v := strings.ToLower(strings.TrimSpace(valor))
	switch v {
	case "true", "1", "si", "sí", "s", "verdadero", "activo":
		return true, nil
	case "false", "0", "no", "n", "falso", "inactivo":
		return false, nil
	default:
		return false, errors.New("debe ser booleano ('Sí', 'No', 'true', 'false', '1', '0')")
	}
}

// ParseFecha valida una fecha en formato YYYY-MM-DD.
func ParseFecha(valor string) (string, error) {
	v := strings.TrimSpace(valor)
	_, err := time.Parse("2006-01-02", v)
	if err != nil {
		return "", errors.New("formato de fecha inválido, debe ser AAAA-MM-DD (ej: 2026-10-21)")
	}
	return v, nil
}

// ParseHora valida una hora en formato HH:MM (24h).
func ParseHora(valor string) (string, error) {
	v := strings.TrimSpace(valor)
	// Si viene con segundos (HH:MM:SS), normalizar a HH:MM
	if len(v) == 8 && v[2] == ':' && v[5] == ':' {
		v = v[:5]
	}
	_, err := time.Parse("15:04", v)
	if err != nil {
		return "", errors.New("formato de hora inválido, debe ser HH:MM en formato 24 horas (ej: 14:30)")
	}
	return v, nil
}

// ParseEmail valida sintaxis de correo electrónico.
func ParseEmail(valor string) (string, error) {
	addr, err := mail.ParseAddress(strings.TrimSpace(valor))
	if err != nil || !strings.Contains(addr.Address, "@") || !strings.Contains(addr.Address, ".") {
		return "", errors.New("correo electrónico inválido")
	}
	return strings.ToLower(addr.Address), nil
}

// ParseURL valida una URL bien formada.
func ParseURL(valor string) (string, error) {
	v := strings.TrimSpace(valor)
	u, err := url.ParseRequestURI(v)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", errors.New("URL inválida, debe incluir protocolo (ej: https://www.ejemplo.gob.mx)")
	}
	return v, nil
}

var hexColorRegex = regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)

// ParseColor valida un color hexadecimal (#RGB o #RRGGBB).
func ParseColor(valor string) (string, error) {
	v := strings.TrimSpace(valor)
	if !hexColorRegex.MatchString(v) {
		return "", errors.New("color inválido, debe ser formato hexadecimal (ej: #691C32)")
	}
	return strings.ToUpper(v), nil
}

// ParseEnum valida que el valor pertenezca a un conjunto de opciones permitidas.
func ParseEnum(valor string, opcionesValidas []string) (string, error) {
	v := strings.TrimSpace(valor)
	for _, opt := range opcionesValidas {
		if strings.EqualFold(v, opt) {
			return opt, nil
		}
	}
	return "", fmt.Errorf("valor no permitido. Opciones válidas: %s", strings.Join(opcionesValidas, ", "))
}
