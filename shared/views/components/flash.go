package components

import (
	"github.com/a-h/templ"
	"github.com/issemym-repos/sesiones"
)

func renderFlashMessage(msg sesiones.MensajeFlash) templ.Component {
	return MensajeFlashDirecto(normalizarVarianteFlash(string(msg.Tipo)), msg.Contenido)
}

func normalizarVarianteFlash(variante string) string {
	switch variante {
	case "exito", "success":
		return "success"
	case "error":
		return "error"
	case "warning", "advertencia":
		return "warning"
	case "info", "informacion":
		return "info"
	default:
		return "info"
	}
}

func claseToast(variante string) string {
	switch variante {
	case "success":
		return ""
	case "error":
		return ""
	case "warning":
		return ""
	default:
		return ""
	}
}

func claseIconoToast(variante string) string {
	switch variante {
	case "success":
		return "text-green-500 bg-green-100 dark:bg-green-800 dark:text-green-200"
	case "error":
		return "text-red-500 bg-red-100 dark:bg-red-800 dark:text-red-200"
	case "warning":
		return "text-yellow-500 bg-yellow-100 dark:bg-yellow-800 dark:text-yellow-200"
	default:
		return "text-blue-500 bg-blue-100 dark:bg-blue-800 dark:text-blue-200"
	}
}

func iconoToast(variante string) string {
	switch variante {
	case "success":
		return "check_circle"
	case "error":
		return "highlight_off"
	case "warning":
		return "warning"
	default:
		return "info"
	}
}
