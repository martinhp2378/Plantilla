package pagination

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
)

// CriteriosPagina encapsula toda la información necesaria para paginar, filtrar y ordenar.
type CriteriosPagina[T any] struct {
	Data       []T            `json:"data"`
	Count      int            `json:"count"`     // Total de registros en BD
	Size       int            `json:"size"`      // Registros por página (LIMIT)
	Index      int            `json:"page"`      // Página actual (inicia en 1)
	First      int            `json:"offset"`    // Offset calculado para SQL
	FilteredBy map[string]any `json:"filters"`   // Filtros aplicados
	SortedBy   string         `json:"sort_by"`   // Columna de ordenamiento
	Direction  string         `json:"direction"` // "asc" o "desc"
}

// NewCriteriosPaginaDefault crea una instancia con valores seguros.
func NewCriteriosPaginaDefault[T any]() CriteriosPagina[T] {
	return CriteriosPagina[T]{
		Index:      1,
		Size:       10,
		First:      0,
		FilteredBy: make(map[string]any),
		SortedBy:   "", // Asume que todas tus tablas tienen id, ajusta si no
		Direction:  "asc",
	}
}

// WithSort actualiza inmutablemente el ordenamiento.
// Si la columna es la misma, invierte la dirección. Si es nueva, empieza en ASC.
func (c CriteriosPagina[T]) WithSort(column string) CriteriosPagina[T] {
	if c.SortedBy == column {
		if strings.ToLower(c.Direction) == "asc" {
			c.Direction = "desc"
		} else {
			c.Direction = "asc"
		}
	} else {
		c.SortedBy = column
		c.Direction = "asc"
	}
	return c
}

// WithFilter añade o actualiza un filtro.
// MEJORA: Acepta 'any' para permitir números, bools, etc.
func (c CriteriosPagina[T]) WithFilter(clave string, valor any) CriteriosPagina[T] {
	if c.FilteredBy == nil {
		c.FilteredBy = make(map[string]any)
	}
	// Si el valor es string vacío o nil, podríamos optar por eliminar el filtro
	if v, ok := valor.(string); ok && v == "" {
		delete(c.FilteredBy, clave)
	} else {
		c.FilteredBy[clave] = valor
	}
	return c
}

// WithPage establece la página actual y RECALCULA automáticamente el First (Offset).
// Esta es la función más importante para evitar bugs de datos desalineados.
func (c CriteriosPagina[T]) WithPage(page int) CriteriosPagina[T] {
	if page < 1 {
		page = 1
	}
	c.Index = page
	c.First = (page - 1) * c.Size
	return c
}

// Métodos de Navegación Relativa (Más seguros que pasar argumentos externos)

func (c CriteriosPagina[T]) NextPage() CriteriosPagina[T] {
	return c.WithPage(c.Index + 1)
}

func (c CriteriosPagina[T]) PreviousPage() CriteriosPagina[T] {
	return c.WithPage(c.Index - 1)
}

// --- Helpers para la Vista (Templates) ---

// TotalPages calcula cuántas páginas existen basándose en Count y Size.
func (c CriteriosPagina[T]) TotalPages() int {
	if c.Size == 0 {
		return 0
	}
	return int(math.Ceil(float64(c.Count) / float64(c.Size)))
}

// SortIcon devuelve una indicación visual (útil para tablas simples).
func (c CriteriosPagina[T]) SortIcon(column string) string {
	if c.SortedBy != column {
		return "↕" // O una clase CSS opacity-20
	}
	if strings.ToLower(c.Direction) == "asc" {
		return "▲"
	}
	return "▼"
}

// GetFilterValue devuelve el valor literal del filtro aplicable al campo como string.
func (c CriteriosPagina[T]) GetFilterValue(field string) string {
	if val, ok := c.FilteredBy[field]; ok && val != nil {
		return fmt.Sprintf("%v", val)
	}
	return ""
}

// BuildURL reconstruye la URL actual preservando filtros, orden y paginación.
// Es vital para que los links de "Siguiente", "Ordenar", etc., no pierdan los filtros activos.
func (c CriteriosPagina[T]) BuildURL(baseURL string, targetPage int) string {
	q := url.Values{}

	q.Set("page", strconv.Itoa(targetPage))
	q.Set("size", strconv.Itoa(c.Size))

	if c.SortedBy != "" {
		q.Set("sort", c.SortedBy)
		q.Set("dir", strings.ToLower(c.Direction))
	}

	for k, v := range c.FilteredBy {
		if v != nil {
			q.Set(k, fmt.Sprint(v))
		}
	}

	// Verifica si la baseURL ya tiene query params
	separator := "?"
	if strings.Contains(baseURL, "?") {
		separator = "&"
	}

	return fmt.Sprintf("%s%s%s", baseURL, separator, q.Encode())
}

// FromQuery construye criterios de paginación desde query params de un request HTTP.
// Extrae page, size, sort, dir automáticamente. TODOS los demás params se convierten en filtros.
// Esto permite que los filtros de tabla funcionen correctamente con hx-get + hx-include.
func FromQuery[T any](q url.Values) CriteriosPagina[T] {
	// Parámetros reservados que NO son filtros
	reserved := map[string]bool{
		"page":  true,
		"size":  true,
		"sort":  true,
		"dir":   true,
		"index": true,
	}

	page := 1
	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}

	size := 10
	if v := q.Get("size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			size = n
		}
	}

	filtros := make(map[string]any)
	for key, vals := range q {
		if reserved[key] || len(vals) == 0 {
			continue
		}
		val := vals[0]
		if val == "" {
			continue
		}
		filtros[key] = val
	}

	return CriteriosPagina[T]{
		Index:      page,
		Size:       size,
		First:      (page - 1) * size,
		FilteredBy: filtros,
		SortedBy:   q.Get("sort"),
		Direction:  strings.ToLower(q.Get("dir")),
	}
}

// Helpers para mantener compatibilidad con tu código de vista anterior si lo necesitas:

func (c CriteriosPagina[T]) WithPageFoward(ignored int) CriteriosPagina[T] {
	// Ignoramos el argumento porque es más seguro usar el estado interno
	return c.NextPage()
}

func (c CriteriosPagina[T]) WithPageBackward(ignored int) CriteriosPagina[T] {
	return c.PreviousPage()
}
