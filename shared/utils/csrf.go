package utils

import (
	"PlantillaInstitucional/shared/contextkey"
	"context"
)

// GetCSRFToken recupera el token CSRF del contexto de forma segura.
func GetCSRFToken(ctx context.Context) string {
	if token, ok := ctx.Value(contextkey.CSRFTokenKey).(string); ok {
		return token
	}
	return ""
}
