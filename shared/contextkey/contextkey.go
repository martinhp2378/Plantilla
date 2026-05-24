package contextkey

import "github.com/gofrs/uuid/v5"

// contextKey es un tipo privado para evitar colisiones.
type contextKey string

// EstatusBandejaKey es la clave para inyectar y leer
// el estatus de la bandeja de trabajo desde el contexto.
const EstatusBandejaKey contextKey = "estatusBandeja"

// AuditUUIDKey es la clave para inyectar y leer el UUID del usuario para auditoría desde el contexto.
const AuditUUIDKey contextKey = "auditUUID"

// AplicacionIDKey es la clave para injectar/leer el UUID de aplicación en dashboards operativos.
const AplicacionIDKey contextKey = "aplicacionID"

// SemestreKey es la clave para injectar/leer el filtro de semestre en dashboards operativos.
const SemestreKey contextKey = "semestre"

// AnioKey es la clave para injectar/leer el filtro de año en dashboards ejecutivos.
const AnioKey contextKey = "anio"

// TenantSchemaKey es la clave para inyectar y leer el esquema del tenant desde el contexto.
const TenantSchemaKey contextKey = "tenantSchema"

// CSRFTokenKey es la clave para inyectar y leer el token CSRF desde el contexto.
const CSRFTokenKey contextKey = "csrfToken"

// GetAplicacionID retrieves the aplicacionID from context, returns zero UUID if not set.
func GetAplicacionID(ctx interface{ Value(interface{}) interface{} }) uuid.UUID {
	if v := ctx.Value(AplicacionIDKey); v != nil {
		if id, ok := v.(uuid.UUID); ok {
			return id
		}
	}
	return uuid.Nil
}

// GetSemestre retrieves the semestre from context, returns empty string if not set.
func GetSemestre(ctx interface{ Value(interface{}) interface{} }) string {
	if v := ctx.Value(SemestreKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetAnio retrieves the anio from context, returns empty string if not set.
func GetAnio(ctx interface{ Value(interface{}) interface{} }) string {
	if v := ctx.Value(AnioKey); v != nil {
		if a, ok := v.(string); ok {
			return a
		}
	}
	return ""
}
