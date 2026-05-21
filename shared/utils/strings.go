package utils

// ValueOrZero devuelve el valor en string o "0" si está vacío.
func ValueOrZero(v string) string {
	if v == "" {
		return "0"
	}
	return v
}
