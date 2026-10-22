package repository

import (
	"PlantillaInstitucional/proveedor"
	"PlantillaInstitucional/pagination"
	shared "PlantillaInstitucional/shared/db"
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *ProveedorRepository) Consultar(ctx context.Context, criterios pagination.CriteriosPagina[proveedor.Proveedor]) ([]proveedor.Proveedor, int, error) {
	config := shared.FilterConfig{
		Texto:     colsTexto,
		Numericas: colsNumericas,
		UUID:      colsUUID,
	}
	filtros, args := shared.BuildDynamicWhere(criterios.FilteredBy, config, "proveedor")
	whereClause := ""
	if len(filtros) > 0 {
		whereClause = " WHERE " + strings.Join(filtros, " AND ")
	}

	orden, err := shared.BuildOrderBy(criterios.SortedBy, criterios.Direction, colsOrden, "", "nombre ASC")
	if err != nil {
		return nil, 0, err
	}

	queryArgs := append(args, criterios.Size, criterios.First)
	query := fmt.Sprintf(SelectProveedorFiltered, whereClause, orden, len(args)+1, len(args)+2)

	var rowsList []RowProveedor
	var total int

	err = r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, queryArgs...)
		if err != nil {
			return fmt.Errorf("error ejecutando consulta: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var row RowProveedor
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

	entities := make([]proveedor.Proveedor, len(rowsList))
	for i, row := range rowsList {
		entities[i] = row.ToEntity()
	}
	return entities, total, nil
}

func (r *ProveedorRepository) ObtenerPorID(ctx context.Context, id uuid.UUID) (proveedor.Proveedor, error) {
	var row RowProveedor
	err := r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, SelectProveedorByID, id)
		if err != nil {
			return err
		}
		row, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[RowProveedor])
		return err
	})
	if err != nil {
		return proveedor.Proveedor{}, err
	}
	return row.ToEntity(), nil
}

func (r *ProveedorRepository) Crear(ctx context.Context, item proveedor.Proveedor) error {
	return r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		var id uuid.UUID
		return tx.QueryRow(ctx, InsertProveedor, item.Nombre, item.Codigo, item.Descripcion, item.IsActive).Scan(&id)
	})
}

func (r *ProveedorRepository) Actualizar(ctx context.Context, item proveedor.Proveedor) error {
	return r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, UpdateProveedor, item.Nombre, item.Codigo, item.Descripcion, item.IsActive, item.ID)
		if err != nil {
			return err
		}
		if ct.RowsAffected() == 0 {
			return fmt.Errorf("registro no encontrado para actualizar")
		}
		return nil
	})
}

func (r *ProveedorRepository) Eliminar(ctx context.Context, id uuid.UUID) error {
	return r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, DeleteProveedor, id)
		if err != nil {
			return err
		}
		if ct.RowsAffected() == 0 {
			return fmt.Errorf("registro no encontrado para eliminar")
		}
		return nil
	})
}
