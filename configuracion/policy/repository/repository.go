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
	return r.RouteAllowedWithMethod(ctx, profile, "*", path)
}

func (r *PolicyRepository) RouteAllowedWithMethod(ctx context.Context, profile string, method string, path string) (bool, error) {
	if method == "" {
		method = "*"
	}
	var allowed bool
	err := r.Tenant.Read(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, routeAllowedQuery, profile, method, path).Scan(&allowed)
	})
	if err != nil {
		return false, err
	}
	return allowed, nil
}

func (r *PolicyRepository) ProfileHasRequiredPolicies(ctx context.Context, profile string) (bool, error) {
	var active bool
	err := r.Tenant.Read(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, profileHasPoliciesQuery, profile).Scan(&active)
	})
	if err != nil {
		return false, err
	}
	return active, nil
}

func (r *PolicyRepository) ListarPoliticas(ctx context.Context, criterios pagination.CriteriosPagina[policy.Politica]) ([]policy.Politica, int, error) {
	var politicas []policy.Politica
	var total int

	err := r.Tenant.Read(ctx, func(tx pgx.Tx) error {
		filtros, args := shared.BuildDynamicWhere(criterios.FilteredBy, shared.FilterConfig{Texto: colsTextoPolicy, Numericas: colsNumericasPolicy, UUID: colsUUIDPolicy}, "p")
		whereClause := ""
		if len(filtros) > 0 {
			whereClause = " WHERE " + strings.Join(filtros, " AND ")
		}

		orden, err := shared.BuildOrderBy(criterios.SortedBy, criterios.Direction, colsOrdenPolicy, "p", "nombre ASC")
		if err != nil {
			return err
		}

		queryArgs := append(args, criterios.Size, criterios.First)
		query := fmt.Sprintf(listPoliticasFiltered, whereClause, orden, len(args)+1, len(args)+2)
		rows, err := tx.Query(ctx, query, queryArgs...)
		if err != nil {
			return fmt.Errorf("ListarPoliticas: error ejecutando query: %w", err)
		}

		politicas, err = pgx.CollectRows(rows, pgx.RowToStructByName[policy.Politica])
		if err != nil {
			return fmt.Errorf("ListarPoliticas: error mapeando resultados: %w", err)
		}

		if len(politicas) == 0 && criterios.First == 0 {
			return nil
		}

		if err := tx.QueryRow(ctx, fmt.Sprintf(countPoliticas, whereClause), args...).Scan(&total); err != nil {
			return fmt.Errorf("ListarPoliticas: error contando registros: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return politicas, total, nil
}

func (r *PolicyRepository) ObtenerPolitica(ctx context.Context, nombre string) (policy.Politica, error) {
	var p policy.Politica
	err := r.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, getPoliticaByNombre, nombre)
		if err != nil {
			return fmt.Errorf("ObtenerPolitica: error de conexión: %w", err)
		}

		p, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[policy.Politica])
		if err != nil {
			if err == pgx.ErrNoRows {
				return fmt.Errorf("política no encontrada (nombre: %s): %w", nombre, err)
			}
			return fmt.Errorf("ObtenerPolitica: %w", err)
		}

		return nil
	})

	if err != nil {
		return policy.Politica{}, err
	}

	return p, nil
}

func (r *PolicyRepository) CrearPolitica(ctx context.Context, p policy.Politica) error {
	return r.Tenant.Write(ctx, func(tx pgx.Tx) error {
		var nuevo string
		if err := tx.QueryRow(ctx, insertPolitica, p.Nombre, p.Descripcion, p.Perfil, p.Recurso, p.Accion, p.Reglas, p.PathPrefix).Scan(&nuevo); err != nil {
			return fmt.Errorf("CrearPolitica: error al insertar: %w", err)
		}
		return nil
	})
}

func (r *PolicyRepository) ActualizarPolitica(ctx context.Context, p policy.Politica, nombreOriginal string) error {
	return r.Tenant.Write(ctx, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, updatePolitica, p.Nombre, p.Descripcion, p.Perfil, p.Recurso, p.Accion, p.Reglas, p.PathPrefix, nombreOriginal)
		if err != nil {
			return fmt.Errorf("ActualizarPolitica: %w", err)
		}
		if ct.RowsAffected() == 0 {
			return fmt.Errorf("no se encontró la política para actualizar (nombre: %s)", nombreOriginal)
		}
		return nil
	})
}

func (r *PolicyRepository) EliminarPolitica(ctx context.Context, nombre string) error {
	return r.Tenant.Write(ctx, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, deletePolitica, nombre)
		if err != nil {
			return fmt.Errorf("EliminarPolitica: %w", err)
		}
		if ct.RowsAffected() == 0 {
			return fmt.Errorf("no se encontró la política para eliminar (nombre: %s)", nombre)
		}
		return nil
	})
}

func (r *PolicyRepository) ListarPerfiles(ctx context.Context) ([]string, error) {
	var items []catalogItem
	err := r.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, listPerfiles)
		if err != nil {
			return err
		}
		items, err = pgx.CollectRows(rows, pgx.RowToStructByName[catalogItem])
		return err
	})
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
	var items []catalogItem
	err := r.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, listRecursos)
		if err != nil {
			return err
		}
		items, err = pgx.CollectRows(rows, pgx.RowToStructByName[catalogItem])
		return err
	})
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
	var items []catalogItem
	err := r.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, listAcciones)
		if err != nil {
			return err
		}
		items, err = pgx.CollectRows(rows, pgx.RowToStructByName[catalogItem])
		return err
	})
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for _, i := range items {
		out = append(out, i.Nombre)
	}
	return out, nil
}
