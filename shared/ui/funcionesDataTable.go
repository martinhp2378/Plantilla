package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
)

// ConvertToString convierte varios tipos a string
func ConvertToString[T any](value T) string {
	switch v := any(value).(type) {
	case int:
		return strconv.Itoa(v)
	case uuid.UUID:
		return v.String()
	case bool:
		if v {
			return "true"
		}
		return "false"
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

// SafeString extrae y convierte un valor de un mapa genérico
func SafeString[T any](m map[string]T, key string) string {
	if v, ok := m[key]; ok {
		return ConvertToString(v)
	}
	return ""
}
