package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
)

var (
	colsTexto = map[string]bool{
		"clave": true, "valor": true, "descripcion": true, "tipo_dato": true, "razon_i": true, "razon_m": true,
	}
	colsNumericas = map[string]bool{}
	colsUUID      = map[string]bool{"parametro_id": true}
	// Columnas permitidas estrictamente para ordenamiento
	colsOrden = map[string]bool{
		"parametro_id": true, "clave": true, "valor": true, "descripcion": true, "tipo_dato": true, "is_active": true,
		"razon_i": true, "razon_m": true, "fecha_i": true, "fecha_m": true,
	}
)








func (r *ParametrosRepository) GetJSONParamContext(ctx context.Context, key string) (string, error) {
	var value string
	err := r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, SelectParametroValorByClave, key).Scan(&value)
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

