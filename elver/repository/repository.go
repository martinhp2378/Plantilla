package repository

import (
	"PlantillaInstitucional/elver"
	"PlantillaInstitucional/pagination"
	shared "PlantillaInstitucional/shared/db"
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *ElverRepository) Consultar(ctx context.Context, criterios pagination.CriteriosPagina[elver.Elver]) ([]elver.Elver, int, error) {
	config := shared.FilterConfig{
		Texto:     colsTexto,
		Numericas: colsNumericas,
		UUID:      colsUUID,
	}
	filtros, args := shared.BuildDynamicWhere(criterios.FilteredBy, config, "elver")
	whereClause := ""
	if len(filtros) > 0 {
		whereClause = " WHERE " + strings.Join(filtros, " AND ")
	}

	orden, err := shared.BuildOrderBy(criterios.SortedBy, criterios.Direction, colsOrden, "", "nombre ASC")
	if err != nil {
		return nil, 0, err
	}

	queryArgs := append(args, criterios.Size, criterios.First)
	query := fmt.Sprintf(SelectElverFiltered, whereClause, orden, len(args)+1, len(args)+2)

	var rowsList []RowElver
	var total int

	err = r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, queryArgs...)
		if err != nil {
			return fmt.Errorf("error ejecutando consulta: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var row RowElver
			var count int
			if err := rows.Scan(&row.ID, &row.Nombre, &row.Codigo, &row.Descripcion, &row.IsActive, &row.FechaI, &row.FechaM, &count); err != nil {
				return err
			}
			rowsList = append(rowsList, row)
			total = count
		}
		return rows.Err()
	})
	if err != nil {
		return nil, 0, err
	}

	entities := make([]elver.Elver, len(rowsList))
	for i, row := range rowsList {
		entities[i] = row.ToEntity()
	}
	return entities, total, nil
}

func (r *ElverRepository) ObtenerPorID(ctx context.Context, id uuid.UUID) (elver.Elver, error) {
	var row RowElver
	err := r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, SelectElverByID, id)
		if err != nil {
			return err
		}
		row, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[RowElver])
		return err
	})
	if err != nil {
		return elver.Elver{}, err
	}
	return row.ToEntity(), nil
}

func (r *ElverRepository) ObtenerTodos(ctx context.Context) ([]elver.Elver, error) {
	var items []elver.Elver
	err := r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, SelectAllElver)
		if err != nil {
			return fmt.Errorf("error ejecutando consulta: %w", err)
		}

		rowsList, err := pgx.CollectRows(rows, pgx.RowToStructByName[RowElver])
		if err != nil {
			return fmt.Errorf("error mapeando resultados: %w", err)
		}

		items = make([]elver.Elver, len(rowsList))
		for i, row := range rowsList {
			items[i] = row.ToEntity()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ElverRepository) Crear(ctx context.Context, item elver.Elver) error {
	return r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		var id uuid.UUID
		return tx.QueryRow(ctx, InsertElver, item.Nombre, item.Codigo, item.Descripcion, item.IsActive).Scan(&id)
	})
}

func (r *ElverRepository) Actualizar(ctx context.Context, item elver.Elver) error {
	return r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, UpdateElver, item.Nombre, item.Codigo, item.Descripcion, item.IsActive, item.ID)
		if err != nil {
			return err
		}
		if ct.RowsAffected() == 0 {
			return fmt.Errorf("registro no encontrado para actualizar")
		}
		return nil
	})
}

func (r *ElverRepository) Eliminar(ctx context.Context, id uuid.UUID) error {
	return r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, DeleteElver, id)
		if err != nil {
			return err
		}
		if ct.RowsAffected() == 0 {
			return fmt.Errorf("registro no encontrado para eliminar")
		}
		return nil
	})
}
