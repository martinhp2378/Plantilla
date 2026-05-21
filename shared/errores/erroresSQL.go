package errores

import (
	"net/http"

	"github.com/issemym-repos/logs"
	"github.com/lib/pq"
)

type SQLErrorHandler struct {
	Logger *logs.Logger
}

func NewSQLErrorHandler(logger *logs.Logger) *SQLErrorHandler {

	return &SQLErrorHandler{Logger: logger}
}

func (h *SQLErrorHandler) Handle(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}

	// Usar el traductor de errores para mensajes amigables
	return http.StatusInternalServerError, TraducirError(err)
}

func (h *SQLErrorHandler) HandleWithOperation(err error, operacion string) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}

	// Usar el traductor con operación específica
	return http.StatusInternalServerError, TraducirErrorExplicit(err, operacion)
}

func (h *SQLErrorHandler) handlePostgresError(err *pq.Error) (int, string) {
	// Traducir errores PostgreSQL comunes
	switch err.Code {
	case "23505": // Unique violation
		return http.StatusConflict, "Ya existe un registro con estos datos"
	case "23503": // Foreign key violation
		return http.StatusBadRequest, "No se puede modificar porque está relacionado con otros datos"
	case "42601": // Syntax error
		return http.StatusBadRequest, "Error en los datos enviados"
	case "42703": // Invalid column
		return http.StatusInternalServerError, "Error interno del sistema"
	case "23502": // Not null violation
		return http.StatusBadRequest, "Falta completar un campo requerido"
	case "22001": // String data right truncation
		return http.StatusBadRequest, "Los datos ingresados son demasiado largos"
	case "42501": // Insufficient privilege
		return http.StatusForbidden, "No tienes permisos para esta acción"
	case "28P01": // Invalid password
		return http.StatusUnauthorized, "Credenciales incorrectas"
	case "22003": // Numeric value out of range
		return http.StatusBadRequest, "El valor numérico está fuera del rango permitido"
	case "22P02": // Invalid text representation
		return http.StatusBadRequest, "El formato de los datos es inválido"
	case "22007": // Invalid datetime format
		return http.StatusBadRequest, "El formato de fecha/hora es inválido"
	case "22012": // Division by zero
		return http.StatusBadRequest, "Error de cálculo: división por cero"
	case "40001": // Serialization failure
		return http.StatusConflict, "La operación no pudo completarse. Intenta de nuevo"
	case "40P01": // Deadlock detected
		return http.StatusConflict, "El sistema está ocupado. Intenta de nuevo"
	case "53300": // Too many connections
		return http.StatusServiceUnavailable, "El sistema está temporalmente saturado"
	case "53200": // Out of memory
		return http.StatusServiceUnavailable, "Memoria insuficiente"
	case "53100": // Disk full
		return http.StatusServiceUnavailable, "Espacio de almacenamiento insuficiente"
	case "42P01": // undefined_table
		return http.StatusInternalServerError, "Error interno del sistema"
	case "42883": // undefined_function
		return http.StatusInternalServerError, "Error interno del sistema"
	default:
		return http.StatusInternalServerError, TraducirError(err)
	}
}
