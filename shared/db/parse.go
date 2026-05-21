package db

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofrs/uuid/v5"
)

// ParseUUID valida y convierte un string a UUID
// NOTA: Si el parámetro está vacío, retorna uuid.Nil sin error.
// Llama a esta función solo cuando el UUID es OPCIONAL en el contexto.
// Para UUIDs requeridos, verifica el resultado con uuid.Nil.
func ParseUUID(param string) (uuid.UUID, error) {
	if param == "" {
		return uuid.Nil, nil // Empty = no filter provided
	}

	id, err := uuid.FromString(param)
	if err != nil {
		return uuid.Nil, errors.New("ID con formato inválido")
	}

	return id, nil
}

func ParseIntOrDefault(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil && v > 0 {
		return v
	}
	return def
}

// ParseTime convierte una cadena de texto en un objeto time.Time.
// Asume el formato "YYYY-MM-DD".
func ParseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil // Retorna un valor de tiempo cero si la cadena está vacía
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("formato de fecha inválido: %w", err)
	}
	return t, nil
}

// ParseInt convierte una cadena de texto a un entero (int).
func ParseInt(s string) (int, error) {
	if s == "" {
		return 0, nil // Retorna 0 si la cadena está vacía
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("valor de número inválido: %w", err)
	}
	return i, nil
}

// ParseFloat convierte una cadena de texto a un valor decimal.
func ParseFloat64(s string) (float64, error) {
	if s == "" {
		return 0.0, nil
	}

	// strconv.ParseFloat convierte una cadena a float64
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, fmt.Errorf("valor de número flotante inválido: %w", err)
	}

	return f, nil
}
