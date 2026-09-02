package errores

import (
	"errors"
	"strings"
)

// Errores centinela para casos comunes
var (
	ErrRegistroNoEncontrado  = errors.New("registro no encontrado")
	ErrConexionBD            = errors.New("error de conexión con la base de datos")
	ErrDatoInvalido          = errors.New("dato inválido")
	ErrCredencialesInvalidas = errors.New("credenciales inválidas")
)

// TraducirError convierte errores técnicos en mensajes amigables para el usuario
func TraducirError(err error) string {
	if err == nil {
		return "Ocurrió un error inesperado"
	}

	errMsg := err.Error()

	// Errores de base de datos
	switch {
	case strings.Contains(errMsg, "no rows"):
		return "El registro que buscas no existe o ya fue eliminado"
	case strings.Contains(errMsg, "connection refused"):
		return "No se pudo conectar con la base de datos. Intenta más tarde"
	case strings.Contains(errMsg, "context deadline exceeded"):
		return "La operación tardó demasiado. Intenta de nuevo"
	case strings.Contains(errMsg, "pq: duplicate key"):
		return "Ya existe un registro con esos datos"
	case strings.Contains(errMsg, "pq: null value"):
		return "Faltan datos requeridos"
	case strings.Contains(errMsg, "SQLSTATE 42703"):
		return "Error interno: columna no encontrada"
	case strings.Contains(errMsg, "SQLSTATE 42883"):
		return "Error interno: operador no válido"
	case strings.Contains(errMsg, "SQLSTATE 42P01"):
		return "Error interno: tabla no encontrada"
	case strings.Contains(errMsg, "SQLSTATE 23503"):
		return "No se puede eliminar porque está relacionado con otros datos"
	case strings.Contains(errMsg, "foreign key"):
		return "No se puede eliminar porque está relacionado con otros datos"
	case strings.Contains(errMsg, "viola la llave foránea"):
		return "No se puede eliminar porque está relacionado con otros datos"
	case strings.Contains(errMsg, "error ejecutando query"):
		return "Error al cargar los datos. Intenta de nuevo"
	case strings.Contains(errMsg, "error mapeando resultados"):
		return "Error al procesar los datos"
	case strings.Contains(errMsg, "error de conexión"):
		return "Error de conexión con la base de datos"
	case strings.Contains(errMsg, "error obteniendo"):
		return "Error al obtener los datos"
	case strings.Contains(errMsg, "no hay resultados completados"):
		return "No hay respuestas completadas para generar el reporte"
	case strings.Contains(errMsg, "10000 row limit"):
		return "El reporte excede el límite de 10,000 registros"
	case strings.Contains(errMsg, "50 MiB file limit"):
		return "El reporte excede el límite de archivo de 50 MiB"
	case strings.Contains(errMsg, "parsing"):
		return "Los datos recibidos tienen un formato inválido"
	case strings.Contains(errMsg, "uuid"):
		return "El identificador no es válido"
	case strings.Contains(errMsg, "syntax error"):
		return "Error en los datos enviados"
	case strings.Contains(errMsg, "invalid character"):
		return "Caracteres inválidos en los datos"
	case strings.Contains(errMsg, "token QR"):
		return "La aplicación no tiene un código QR generado"
	}

	// Errores genéricos del sistema
	switch {
	case strings.Contains(errMsg, "timeout"):
		return "La operación tardó demasiado"
	case strings.Contains(errMsg, "permission denied"):
		return "No tienes permisos para realizar esta acción"
	case strings.Contains(errMsg, "not found"):
		return "El recurso solicitado no existe"
	case strings.Contains(errMsg, "already exists"):
		return "Ya existe un registro con esos datos"
	case strings.Contains(errMsg, "required"):
		return "Faltan datos requeridos"
	case strings.Contains(errMsg, "invalid"):
		return "Los datos proporcionados no son válidos"
	}

	// Si no conocemos el error, mostramos uno genérico amigable
	return "Ocurrió un error. Intenta de nuevo o contacta al administrador"
}

// TraducirErrorExplicit devuelve un mensaje según el tipo de operación
func TraducirErrorExplicit(err error, operacion string) string {
	baseMsg := TraducirError(err)

	// Personalizar según la operación
	switch operacion {
	case "crear":
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "ya existe") {
			return "Ya existe un registro con esos datos"
		}
		return "Error al guardar. " + baseMsg
	case "actualizar":
		if strings.Contains(err.Error(), "no rows") || strings.Contains(err.Error(), "no encontrado") {
			return "El registro que intentas actualizar no existe"
		}
		return "Error al actualizar. " + baseMsg
	case "eliminar":
		if strings.Contains(err.Error(), "foreign key") || strings.Contains(err.Error(), "relacionado") {
			return "No se puede eliminar porque está relacionado con otros datos"
		}
		return "Error al eliminar. " + baseMsg
	case "consultar":
		return "Error al cargar los datos. " + baseMsg
	case "generar":
		return "Error al generar el reporte. " + baseMsg
	default:
		return baseMsg
	}
}
