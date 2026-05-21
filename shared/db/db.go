package db

import (
	contextkeys "PlantillaInstitucional/shared/contextkey"
	"context"

	"github.com/jackc/pgx/v5/pgconn"
)

// Executor es una interfaz que abstrae pgx.Conn y pgx.Tx para ejecutar queries
type Executor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// SetAuditSession establece la variable de sesión de PostgreSQL para auditoría
// con el UUID del usuario extraído del contexto.
func SetAuditSession(ctx context.Context, exec Executor) error {
	uuidVal := ctx.Value(contextkeys.AuditUUIDKey)
	var uuid string
	if uuidVal != nil {
		if s, ok := uuidVal.(string); ok {
			uuid = s
		}
	}
	// Si no hay UUID (usuario no autenticado), se establece como cadena vacía
	// La función de auditoría en la DB maneja NULL gracefully
	_, err := exec.Exec(ctx, "SELECT set_config('session_variables.uuid_persona', $1, false)", uuid)
	return err
}
