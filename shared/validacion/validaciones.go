package validacion

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	errores "PlantillaInstitucional/shared/errores"
)

// Estructura de respuesta JSON
type ErrorResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Extrae los errores en un mapa para plantillas
func ValidationErrorsMap(err error) map[string]string {
	errorsMap := make(map[string]string)

	if verrs, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range verrs {
			errorsMap[strings.ToLower(fe.Field())] = getValidationMessage(fe)
		}
	}

	return errorsMap
}

// Envía los errores como JSON (API)
func RespondWithValidationErrors(w http.ResponseWriter, err error) {
	details := ValidationErrorsMap(err)

	w.WriteHeader(http.StatusUnprocessableEntity)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Code:    http.StatusUnprocessableEntity,
		Message: "Errores de validación",
		Details: details,
	})
}

// Traducción de mensajes de validación
func getValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return errores.MensajeRequeridoCampo
	case "email":
		return "Formato de email inválido"
	case "uuid":
		return "Debe ser un UUID válido"
	case "oneof":
		return fmt.Sprintf("Valor permitido: %s", fe.Param())
	case "min":
		return fmt.Sprintf("Mínimo %s caracteres", fe.Param())
	case "max":
		return fmt.Sprintf("Máximo %s caracteres", fe.Param())
	case "alphanum":
		return "Solo caracteres alfanuméricos"
	case "alpha":
		return "Solo caracteres alfa"
	case "fechaNoPasada":
		return "La fecha no puede ser anterior a hoy"
	default:
		return "Valor inválido"
	}
}

// Registra las validaciones personalizadas en el validador
func RegisterCustomValidators(v *validator.Validate) {
	_ = v.RegisterValidation("fechaNoPasada", fechaNoPasada)
}

// Validador: la fecha no puede ser anterior a hoy
func fechaNoPasada(fl validator.FieldLevel) bool {
	fecha, ok := fl.Field().Interface().(time.Time)
	if !ok {
		return false
	}

	hoy := time.Now().Truncate(24 * time.Hour)
	fechaVal := fecha.Truncate(24 * time.Hour)

	return !fechaVal.Before(hoy) // La fecha debe ser hoy o en el futuro
}
