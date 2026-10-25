package db

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
)

// DateRange estructura auxiliar para filtrar rangos de fecha
type DateRange struct {
	Start time.Time
	End   time.Time
}

// FilterConfig define qué columnas soportan qué tipo de filtro
type FilterConfig struct {
	Texto     map[string]bool
	Numericas map[string]bool
	UUID      map[string]bool
	Fecha     map[string]bool
	Boolean   map[string]bool
	TextoCast map[string]bool // Columnas no-texto (decimal, hora, etc.) filtradas con ILIKE mediante cast a ::text
}

// BuildDynamicWhere construye las cláusulas WHERE y argumentos para queries dinámicos.
// Soporta:
// - Texto: ILIKE %valor%
// - Númerico: Igualdad exacta
// - UUID: Igualdad exacta
// - Fecha: Rango (DateRange) o Igualdad exacta (time.Time)
// - Booleano: Igualdad exacta
func BuildDynamicWhere(filtros map[string]any, config FilterConfig, tableAlias string) ([]string, []any) {
	var condiciones []string
	var args []any
	argID := 1
	aliasPrefix := ""
	if tableAlias != "" {
		aliasPrefix = tableAlias + "."
	}

	for campo, valor := range filtros {
		if valor == nil {
			continue
		}

		// ILIKE para texto
		if config.Texto[campo] {
			strVal := fmt.Sprint(valor)
			if strVal != "" {
				condiciones = append(condiciones, fmt.Sprintf("%s%s ILIKE $%d", aliasPrefix, campo, argID))
				args = append(args, "%"+strVal+"%")
				argID++
			}
			continue
		}

		// Igualdad para números
		if config.Numericas[campo] {
			if numVal, ok := ToInt(valor); ok && numVal != 0 {
				condiciones = append(condiciones, fmt.Sprintf("%s%s = $%d", aliasPrefix, campo, argID))
				args = append(args, numVal)
				argID++
			}
			continue
		}

		// Igualdad para UUIDs
		if config.UUID[campo] {
			if uuidVal, ok := ToUUID(valor); ok && uuidVal != uuid.Nil {
				condiciones = append(condiciones, fmt.Sprintf("%s%s = $%d", aliasPrefix, campo, argID))
				args = append(args, uuidVal)
				argID++
			}
			continue
		}

		// Rango de fechas o igualdad simple
		if config.Fecha[campo] {
			if cond, filterArgs, inc, ok := buildFechaFilter(campo, aliasPrefix, valor, argID); ok {
				condiciones = append(condiciones, cond)
				args = append(args, filterArgs...)
				argID += inc
			}
			continue
		}

		// Igualdad para booleanos
		if config.Boolean[campo] {
			if boolVal, ok := ToBool(valor); ok {
				condiciones = append(condiciones, fmt.Sprintf("%s%s = $%d", aliasPrefix, campo, argID))
				args = append(args, boolVal)
				argID++
			}
			continue
		}

		// ILIKE con cast para columnas que no son tipo texto nativo (decimal, hora, etc.)
		if config.TextoCast[campo] {
			strVal := strings.TrimSpace(fmt.Sprint(valor))
			if strVal != "" {
				condiciones = append(condiciones, fmt.Sprintf("%s%s::text ILIKE $%d", aliasPrefix, campo, argID))
				args = append(args, "%"+strVal+"%")
				argID++
			}
			continue
		}
	}
	return condiciones, args
}

// Conversión segura de tipos (Helpers Exportados)
func ToInt(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case float64:
		return int(val), true
	case string:
		if val == "" {
			return 0, false
		}
		i, err := strconv.Atoi(val)
		return i, err == nil
	default:
		return 0, false
	}
}

func ToUUID(v any) (uuid.UUID, bool) {
	switch val := v.(type) {
	case uuid.UUID:
		return val, true
	case string:
		if val == "" {
			return uuid.Nil, false
		}
		uid, err := ParseUUID(val)
		return uid, err == nil
	default:
		return uuid.Nil, false
	}
}

func ToBool(v any) (bool, bool) {
	switch val := v.(type) {
	case bool:
		return val, true
	case string:
		s := strings.TrimSpace(strings.ToLower(val))
		if s == "si" || s == "sí" || s == "true" || s == "1" || s == "t" || s == "verdadero" {
			return true, true
		}
		if s == "no" || s == "false" || s == "0" || s == "f" || s == "falso" {
			return false, true
		}
		b, err := strconv.ParseBool(s)
		return b, err == nil
	case int:
		if val == 1 {
			return true, true
		}
		if val == 0 {
			return false, true
		}
		return false, false
	default:
		return false, false
	}
}

// BuildOrderBy construye la cláusula ORDER BY de forma segura.
// Parámetros:
//   - sortedBy: columna por la cual ordenar (si está vacía, usa defaultOrder)
//   - direction: "ASC" o "DESC" (si está vacía, usa ASC por defecto)
//   - allowedCols: mapa de columnas permitidas para ordenamiento (whitelist)
//   - tableAlias: alias de la tabla (ej: "p", "u") - puede estar vacío
//   - defaultOrder: orden por defecto si sortedBy está vacío (ej: "nombre ASC")
//
// Retorna:
//   - String con la cláusula ORDER BY (ej: "p.nombre DESC")
//   - Error si la columna no está en la lista blanca
func BuildOrderBy(sortedBy, direction string, allowedCols map[string]bool, tableAlias, defaultOrder string) (string, error) {
	// Si no hay columna especificada, devolver el orden por defecto
	if sortedBy == "" {
		return defaultOrder, nil
	}

	// Normalizar: verificar que la columna está permitida (whitelist)
	col := strings.ToLower(sortedBy)
	if !allowedCols[col] {
		return "", fmt.Errorf("columna de ordenamiento no permitida: %s", sortedBy)
	}

	// Normalizar dirección
	dir := "ASC"
	if strings.ToUpper(direction) == "DESC" {
		dir = "DESC"
	}

	// Construir con alias si existe
	if tableAlias != "" {
		return fmt.Sprintf("%s.%s %s", tableAlias, col, dir), nil
	}
	return fmt.Sprintf("%s %s", col, dir), nil
}

func buildFechaFilter(campo, aliasPrefix string, valor any, argID int) (string, []any, int, bool) {
	if dateRange, ok := valor.(DateRange); ok {
		if !dateRange.Start.IsZero() && !dateRange.End.IsZero() {
			return fmt.Sprintf("%s%s BETWEEN $%d AND $%d", aliasPrefix, campo, argID, argID+1),
				[]any{dateRange.Start, dateRange.End}, 2, true
		}
		if !dateRange.Start.IsZero() {
			return fmt.Sprintf("%s%s >= $%d", aliasPrefix, campo, argID),
				[]any{dateRange.Start}, 1, true
		}
		if !dateRange.End.IsZero() {
			return fmt.Sprintf("%s%s <= $%d", aliasPrefix, campo, argID),
				[]any{dateRange.End}, 1, true
		}
	} else if t, ok := valor.(time.Time); ok && !t.IsZero() {
		return fmt.Sprintf("%s%s = $%d", aliasPrefix, campo, argID),
			[]any{t}, 1, true
	} else if s, ok := valor.(string); ok && strings.TrimSpace(s) != "" {
		s = strings.TrimSpace(s)
		return fmt.Sprintf("%s%s::text ILIKE $%d", aliasPrefix, campo, argID),
			[]any{"%" + s + "%"}, 1, true
	}
	return "", nil, 0, false
}
