package repository

import (
	"PlantillaInstitucional/parametros"
	"context"
	"fmt"
	"strconv"
	"time"

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

func (r *ParametrosRepository) CargarParametros() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.params = make(map[string]parametros.Parametro)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var parametrosList []RowParametro
	err := r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, SelectParametros)
		if err != nil {
			return fmt.Errorf("error ejecutando consulta: %w", err)
		}
		parametrosList, err = pgx.CollectRows(rows, pgx.RowToStructByName[RowParametro])
		if err != nil {
			return fmt.Errorf("error mapping results: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, row := range parametrosList {
		p := row.ToParametro()
		r.params[p.Clave] = p
	}

	return nil
}

func (r *ParametrosRepository) ActualizarParametros(param parametros.Parametro) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.params[param.Clave] = param
	fmt.Printf("Parámetro '%s' actualizado.\n", param.Clave)
}

func (r *ParametrosRepository) MonitorParametros(updateChan <-chan string) {
	// Configura un ticker para recargar los parámetros cada 5 minutos.
	// Esto actúa como una medida de seguridad si la señalización falla.
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Bucle infinito que escucha en un 'select'
	for {
		select {
		case <-ticker.C:
			// Recibimos una señal del ticker, es hora de recargar.
			fmt.Println("Sondeo periódico: Recargando todos los parámetros...")
			if err := r.CargarParametros(); err != nil {
				fmt.Printf("Error durante el sondeo de recarga: %v\n", err)
			}
		case key := <-updateChan:
			// Recibimos una señal del canal, lo que indica un cambio en un parámetro.
			fmt.Printf("Señal de actualización recibida para la llave '%s'. Recargando parámetros...\n", key)
			if err := r.CargarParametros(); err != nil {
				fmt.Printf("Error al recargar parámetros por señal: %v\n", err)
			}
		}
	}
}

func (r *ParametrosRepository) GetParam(key string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.params[key].Valor
}

func (r *ParametrosRepository) GetParametro(key string) (parametros.Parametro, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if val, ok := r.params[key]; ok {
		return val, nil
	}
	return parametros.Parametro{}, fmt.Errorf("parámetro no encontrado: %s", key)
}

func (r *ParametrosRepository) GetIntParam(key string) (int, error) {
	param, err := r.GetParametro(key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(param.Valor)
}

func (r *ParametrosRepository) GetBoolParam(key string) (bool, error) {
	param, err := r.GetParametro(key)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(param.Valor)
}

func (r *ParametrosRepository) GetJSONParam(key string) (string, error) {
	param, err := r.GetParametro(key)
	if err != nil {
		return "", err
	}
	return param.Valor, nil
}

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

func (r *ParametrosRepository) ListAllParameters() ([]parametros.Parametro, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Convert the map to a slice for consistent return type
	var allParams []parametros.Parametro
	for _, value := range r.params {
		allParams = append(allParams, value)
	}
	return allParams, nil
}
