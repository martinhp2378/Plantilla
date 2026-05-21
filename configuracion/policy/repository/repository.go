package repository

import (
	"PlantillaInstitucional/configuracion/policy"
	"PlantillaInstitucional/pagination"
	shared "PlantillaInstitucional/shared/db"
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

var (
	colsTextoPolicy     = map[string]bool{"nombre": true, "perfil": true, "recurso": true, "accion": true}
	colsNumericasPolicy = map[string]bool{}
	colsUUIDPolicy      = map[string]bool{}
	colsOrdenPolicy     = map[string]bool{"nombre": true, "perfil": true, "recurso": true, "accion": true}
)

func (r *PolicyRepository) RouteAllowed(ctx context.Context, profile string, path string) (bool, error) {
	var allowed bool
	err := r.ReadPool.QueryRow(ctx, routeAllowedQuery, profile, path).Scan(&allowed)
	if err != nil {
		return false, err
	}
	return allowed, nil
}

func (r *PolicyRepository) ProfileHasRequiredPolicies(ctx context.Context, profile string) (bool, error) {
	var active bool
	err := r.ReadPool.QueryRow(ctx, profileHasPoliciesQuery, profile).Scan(&active)
	if err != nil {
		return false, err
	}
	return active, nil
}

func (r *PolicyRepository) ListarPoliticas(ctx context.Context, criterios pagination.CriteriosPagina[policy.Politica]) ([]policy.Politica, int, error) {
	filtros, args := shared.BuildDynamicWhere(criterios.FilteredBy, shared.FilterConfig{Texto: colsTextoPolicy, Numericas: colsNumericasPolicy, UUID: colsUUIDPolicy}, "p")
	whereClause := ""
	if len(filtros) > 0 {
		whereClause = " WHERE " + strings.Join(filtros, " AND ")
	}

	orden, err := shared.BuildOrderBy(criterios.SortedBy, criterios.Direction, colsOrdenPolicy, "p", "nombre ASC")
	if err != nil {
		return nil, 0, err
	}

	queryArgs := append(args, criterios.Size, criterios.First)
	query := fmt.Sprintf(listPoliticasFiltered, whereClause, orden, len(args)+1, len(args)+2)
	rows, err := r.ReadPool.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("ListarPoliticas: error ejecutando query: %w", err)
	}

	politicas, err := pgx.CollectRows(rows, pgx.RowToStructByName[policy.Politica])
	if err != nil {
		return nil, 0, fmt.Errorf("ListarPoliticas: error mapeando resultados: %w", err)
	}

	if len(politicas) == 0 && criterios.First == 0 {
		return politicas, 0, nil
	}

	var total int
	if err := r.ReadPool.QueryRow(ctx, fmt.Sprintf(countPoliticas, whereClause), args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("ListarPoliticas: error contando registros: %w", err)
	}

	return politicas, total, nil
}

func (r *PolicyRepository) ObtenerPolitica(ctx context.Context, nombre string) (policy.Politica, error) {
	rows, err := r.ReadPool.Query(ctx, getPoliticaByNombre, nombre)
	if err != nil {
		return policy.Politica{}, fmt.Errorf("ObtenerPolitica: error de conexión: %w", err)
	}

	p, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[policy.Politica])
	if err != nil {
		if err == pgx.ErrNoRows {
			return policy.Politica{}, fmt.Errorf("política no encontrada (nombre: %s): %w", nombre, err)
		}
		return policy.Politica{}, fmt.Errorf("ObtenerPolitica: %w", err)
	}

	return p, nil
}

func (r *PolicyRepository) CrearPolitica(ctx context.Context, p policy.Politica) error {
	conn, err := r.WritePool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error obteniendo conexión: %w", err)
	}
	defer conn.Release()

	if err := shared.SetAuditSession(ctx, conn); err != nil {
		return fmt.Errorf("error configurando auditoría: %w", err)
	}

	var nuevo string
	if err := conn.QueryRow(ctx, insertPolitica, p.Nombre, p.Descripcion, p.Perfil, p.Recurso, p.Accion, p.PathPrefix).Scan(&nuevo); err != nil {
		return fmt.Errorf("CrearPolitica: error al insertar: %w", err)
	}

	return nil
}

func (r *PolicyRepository) ActualizarPolitica(ctx context.Context, p policy.Politica, nombreOriginal string) error {
	conn, err := r.WritePool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error obteniendo conexión: %w", err)
	}
	defer conn.Release()

	if err := shared.SetAuditSession(ctx, conn); err != nil {
		return fmt.Errorf("error configurando auditoría: %w", err)
	}

	ct, err := conn.Exec(ctx, updatePolitica, p.Nombre, p.Descripcion, p.Perfil, p.Recurso, p.Accion, p.PathPrefix, nombreOriginal)
	if err != nil {
		return fmt.Errorf("ActualizarPolitica: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("no se encontró la política para actualizar (nombre: %s)", nombreOriginal)
	}

	return nil
}

func (r *PolicyRepository) EliminarPolitica(ctx context.Context, nombre string) error {
	conn, err := r.WritePool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error obteniendo conexión: %w", err)
	}
	defer conn.Release()

	if err := shared.SetAuditSession(ctx, conn); err != nil {
		return fmt.Errorf("error configurando auditoría: %w", err)
	}

	ct, err := conn.Exec(ctx, deletePolitica, nombre)
	if err != nil {
		return fmt.Errorf("EliminarPolitica: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("no se encontró la política para eliminar (nombre: %s)", nombre)
	}

	return nil
}

func (r *PolicyRepository) ListarPerfiles(ctx context.Context) ([]string, error) {
	rows, err := r.ReadPool.Query(ctx, listPerfiles)
	if err != nil {
		return nil, err
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[catalogItem])
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for _, i := range items {
		out = append(out, i.Nombre)
	}
	return out, nil
}

func (r *PolicyRepository) ListarRecursos(ctx context.Context) ([]string, error) {
	rows, err := r.ReadPool.Query(ctx, listRecursos)
	if err != nil {
		return nil, err
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[catalogItem])
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for _, i := range items {
		out = append(out, i.Nombre)
	}
	return out, nil
}

func (r *PolicyRepository) ListarAcciones(ctx context.Context) ([]string, error) {
	rows, err := r.ReadPool.Query(ctx, listAcciones)
	if err != nil {
		return nil, err
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[catalogItem])
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for _, i := range items {
		out = append(out, i.Nombre)
	}
	return out, nil
}
