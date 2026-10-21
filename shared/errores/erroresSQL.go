package errores

import (
	"net/http"

	"github.com/issemym-repos/logs"
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

