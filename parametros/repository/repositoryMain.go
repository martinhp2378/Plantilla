package repository

import (
	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/parametros"
	shared "PlantillaInstitucional/shared/db"
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *ParametrosRepository) ConsultarParametros(ctx context.Context, criterios pagination.CriteriosPagina[parametros.Parametro]) ([]parametros.Parametro, int, error) {
	// pgxpool handles Acquire/Release internally. It is safer and more efficient.
	db := r.db.ReadPool

	// 1. Dynamic Filter Construction
	config := shared.FilterConfig{
		Texto:     colsTexto,
		Numericas: colsNumericas,
		UUID:      colsUUID,
	}
	filtros, args := shared.BuildDynamicWhere(criterios.FilteredBy, config, "parametro")
	whereClause := ""
	if len(filtros) > 0 {
		whereClause = " WHERE " + strings.Join(filtros, " AND ")
	}

	// 2. Safe Order By Construction
	orden, err := shared.BuildOrderBy(criterios.SortedBy, criterios.Direction, colsOrden, "", "clave ASC")
	if err != nil {
		return nil, 0, err
	}

	// 3. Main Query (Data)
	// Final arguments are: [Filters..., Limit, Offset]
	queryArgs := append(args, criterios.Size, criterios.First)
	query := fmt.Sprintf(SelectParametrosFiltered, whereClause, orden, len(args)+1, len(args)+2)

	rows, err := db.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("ConsultarParametros: error executing query: %w", err)
	}

	// pgx v5: CollectRows iterates, scans and closes rows automatically
	parametrosList, err := pgx.CollectRows(rows, pgx.RowToStructByName[RowParametro])
	if err != nil {
		return nil, 0, fmt.Errorf("ConsultarParametros: error mapping results: %w", err)
	}

	// Optimization: If page 0 and no data, total is 0. Avoid count query.
	if len(parametrosList) == 0 && criterios.First == 0 {
		return nil, 0, nil
	}

	// 4. Secondary Query (Total Count for Pagination)
	var total int
	countQuery := fmt.Sprintf(CountParametros, whereClause)
	if err := db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("ConsultarParametros: error counting records: %w", err)
	}

	return ToParametros(parametrosList), total, nil
}

func (r *ParametrosRepository) ConsultarParametroPorID(ctx context.Context, id uuid.UUID) (parametros.Parametro, error) {
	conn, err := r.db.WritePool.Acquire(ctx)
	if err != nil {
		return parametros.Parametro{}, fmt.Errorf("error obteniendo conexión: %w", err)
	}
	defer conn.Release()

	row, err := conn.Query(ctx, SelectParametroXId, id)
	if err != nil {
		return parametros.Parametro{}, fmt.Errorf("error ejecutando consulta: %w", err)
	}

	respuestaParametro, err := pgx.CollectOneRow(row, pgx.RowToStructByName[RowParametro])
	if err != nil {
		return parametros.Parametro{}, fmt.Errorf("error mapeando resultado: %w", err)
	}

	return respuestaParametro.ToParametro(), nil
}

func (r *ParametrosRepository) ActualizarParametro(ctx context.Context, parametros parametros.Parametro) error {
	conn, err := r.db.WritePool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error obteniendo conexión: %w", err)
	}
	defer conn.Release()

	if err := shared.SetAuditSession(ctx, conn); err != nil {
		return fmt.Errorf("error configurando auditoría: %w", err)
	}

	_, err = conn.Exec(ctx,
		UpdateParametro,
		parametros.Clave,
		parametros.Valor,
		parametros.Descripcion,
		parametros.TipoDato,
		parametros.IsActive,
		parametros.RazonM,
		parametros.ParametroId,
	)

	if err != nil {
		return fmt.Errorf("error ejecutando actualización: %w", err)
	}

	// Actualizar caché
	if err := r.CargarParametros(); err != nil {
		fmt.Printf("Error recargando parámetros tras actualización: %v\n", err)
	}

	return nil
}

func (r *ParametrosRepository) EliminarParametro(ctx context.Context, id uuid.UUID) error {
	conn, err := r.db.WritePool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error obteniendo conexión: %w", err)
	}
	defer conn.Release()

	if err := shared.SetAuditSession(ctx, conn); err != nil {
		return fmt.Errorf("error configurando auditoría: %w", err)
	}

	_, err = conn.Exec(ctx, DeleteParametro, id)
	if err != nil {
		return fmt.Errorf("error eliminando rol: %w", err)
	}

	// Actualizar caché
	if err := r.CargarParametros(); err != nil {
		fmt.Printf("Error recargando parámetros tras eliminación: %v\n", err)
	}

	return nil
}

func (r *ParametrosRepository) CrearParametro(ctx context.Context, parametros parametros.Parametro) error {
	conn, err := r.db.WritePool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error obteniendo conexión: %w", err)
	}
	defer conn.Release()

	if err := shared.SetAuditSession(ctx, conn); err != nil {
		return fmt.Errorf("error configurando auditoría: %w", err)
	}

	_, err = conn.Exec(ctx, InsertParametro,
		parametros.ParametroId,
		parametros.Clave,
		parametros.Valor,
		parametros.Descripcion,
		parametros.TipoDato,
		parametros.IsActive,
		parametros.RazonI,
	)
	if err != nil {
		return fmt.Errorf("error al insertar parametros: %w", err)
	}

	// Actualizar caché
	if err := r.CargarParametros(); err != nil {
		fmt.Printf("Error recargando parámetros tras creación: %v\n", err)
	}

	return nil
}
