package repository

import (
	"PlantillaInstitucional/crud"
	"PlantillaInstitucional/pagination"
	shared "PlantillaInstitucional/shared/db"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

var (
	colsTexto = map[string]bool{
		"texto_corto": true,
		"texto_largo": true,
		"seleccion":   true,
		"correo":      true,
		"telefono":    true,
		"url":         true,
	}
	colsNumericas = map[string]bool{
		"entero": true,
	}
	colsUUID = map[string]bool{
		"id": true,
	}
	colsBoolean = map[string]bool{
		"booleano": true,
	}
	colsFecha = map[string]bool{
		"fecha": true,
	}
	colsTextoCast = map[string]bool{
		"decimal": true,
		"hora":    true,
	}
	colsOrden = map[string]bool{
		"id":          true,
		"texto_corto": true,
		"entero":      true,
		"decimal":     true,
		"booleano":    true,
		"fecha":       true,
		"hora":        true,
		"seleccion":   true,
		"correo":      true,
		"fecha_i":     true,
	}
)

func (r *CrudRepository) ConsultarCrud(ctx context.Context, criterios pagination.CriteriosPagina[crud.Crud]) ([]*crud.Crud, int, error) {
	// 1. Construcción de filtros dinámicos
	filterCfg := shared.FilterConfig{
		Texto:     colsTexto,
		Numericas: colsNumericas,
		UUID:      colsUUID,
		Fecha:     colsFecha,
		Boolean:   colsBoolean,
		TextoCast: colsTextoCast,
	}
	filtros, args := shared.BuildDynamicWhere(criterios.FilteredBy, filterCfg, "c")
	whereClause := ""
	if len(filtros) > 0 {
		whereClause = " WHERE " + strings.Join(filtros, " AND ")
	}

	// 2. Construcción de ordenamiento seguro
	orden, err := shared.BuildOrderBy(criterios.SortedBy, criterios.Direction, colsOrden, "", "fecha_i DESC")
	if err != nil {
		return nil, 0, err
	}

	// 3. Query Principal (Datos paginados)
	queryArgs := append(args, criterios.Size, criterios.First)
	query := fmt.Sprintf(SelectCrudFiltered, whereClause, orden, len(args)+1, len(args)+2)

	var items []*crud.Crud
	var total int

	err = r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, queryArgs...)
		if err != nil {
			return fmt.Errorf("ConsultarCrud: error ejecutando query: %w", err)
		}

		rowsList, err := pgx.CollectRows(rows, pgx.RowToStructByName[RowCrud])
		if err != nil {
			return fmt.Errorf("ConsultarCrud: error mapeando resultados: %w", err)
		}

		if len(rowsList) == 0 && criterios.First == 0 {
			total = 0
			return nil
		}

		items = make([]*crud.Crud, len(rowsList))
		for i, row := range rowsList {
			items[i] = row.ToCrud()
		}

		// 4. Conteo Total para Paginación
		countQuery := fmt.Sprintf(CountCrud, whereClause)
		if err := tx.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
			return fmt.Errorf("ConsultarCrud: error contando registros: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *CrudRepository) ConsultarCrudPorID(ctx context.Context, id uuid.UUID) (*crud.Crud, error) {
	var row RowCrud
	err := r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, SelectCrudById, id)
		if err != nil {
			return fmt.Errorf("ConsultarCrudPorID: error ejecutando consulta: %w", err)
		}

		row, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[RowCrud])
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("registro no encontrado: %w", err)
			}
			return fmt.Errorf("ConsultarCrudPorID: error procesando fila: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return row.ToCrud(), nil
}

func (r *CrudRepository) CrearCrud(ctx context.Context, c crud.Crud) error {
	fechaHoraVal := c.FechaHora
	if fechaHoraVal == "" {
		fechaHoraVal = c.Fecha + "T" + c.Hora + ":00"
	}

	return r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		var returnedID uuid.UUID
		err := tx.QueryRow(ctx, InsertCrud,
			c.Id,
			c.TextoCorto,
			c.TextoLargo,
			c.Entero,
			c.Decimal,
			c.Booleano,
			c.Fecha,
			c.Hora,
			fechaHoraVal,
			c.Seleccion,
			c.Correo,
			c.Telefono,
			c.URL,
			c.Color,
			c.Razon_i,
		).Scan(&returnedID)

		if err != nil {
			return fmt.Errorf("CrearCrud: error al insertar: %w", err)
		}
		return nil
	})
}

func (r *CrudRepository) CrearCrudMasivo(ctx context.Context, items []crud.Crud) error {
	if len(items) == 0 {
		return nil
	}

	return r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		for _, c := range items {
			id := c.Id
			if id == uuid.Nil {
				newID, err := uuid.NewV7()
				if err != nil {
					return fmt.Errorf("CrearCrudMasivo: error generando UUID: %w", err)
				}
				id = newID
			}

			fechaHoraVal := c.FechaHora
			if fechaHoraVal == "" {
				fechaHoraVal = c.Fecha + "T" + c.Hora + ":00"
			}

			razonI := c.Razon_i
			if razonI == "" {
				razonI = "Carga Masiva CSV"
			}

			var returnedID uuid.UUID
			err := tx.QueryRow(ctx, InsertCrud,
				id,
				c.TextoCorto,
				c.TextoLargo,
				c.Entero,
				c.Decimal,
				c.Booleano,
				c.Fecha,
				c.Hora,
				fechaHoraVal,
				c.Seleccion,
				c.Correo,
				c.Telefono,
				c.URL,
				c.Color,
				razonI,
			).Scan(&returnedID)

			if err != nil {
				return fmt.Errorf("CrearCrudMasivo: error al insertar fila (id: %s): %w", id, err)
			}
		}
		return nil
	})
}

func (r *CrudRepository) ActualizarCrud(ctx context.Context, c crud.Crud) error {
	fechaHoraVal := c.FechaHora
	if fechaHoraVal == "" {
		fechaHoraVal = c.Fecha + "T" + c.Hora + ":00"
	}

	return r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		cmdTag, err := tx.Exec(ctx, UpdateCrud,
			c.TextoCorto,
			c.TextoLargo,
			c.Entero,
			c.Decimal,
			c.Booleano,
			c.Fecha,
			c.Hora,
			fechaHoraVal,
			c.Seleccion,
			c.Correo,
			c.Telefono,
			c.URL,
			c.Color,
			c.Razon_m,
			c.Id,
		)
		if err != nil {
			return fmt.Errorf("ActualizarCrud: error al actualizar: %w", err)
		}
		if cmdTag.RowsAffected() == 0 {
			return errors.New("registro no encontrado para actualizar")
		}
		return nil
	})
}

func (r *CrudRepository) EliminarCrud(ctx context.Context, id uuid.UUID) error {
	return r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		cmdTag, err := tx.Exec(ctx, DeleteCrud, id)
		if err != nil {
			return fmt.Errorf("EliminarCrud: error al eliminar: %w", err)
		}
		if cmdTag.RowsAffected() == 0 {
			return errors.New("registro no encontrado para eliminar")
		}
		return nil
	})
}

func (r *CrudRepository) ObtenerTodos(ctx context.Context) ([]crud.Crud, error) {
	var items []crud.Crud
	err := r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, SelectAllCrud)
		if err != nil {
			return fmt.Errorf("ObtenerTodos: error ejecutando consulta: %w", err)
		}

		rowsList, err := pgx.CollectRows(rows, pgx.RowToStructByName[RowCrud])
		if err != nil {
			return fmt.Errorf("ObtenerTodos: error mapeando resultados: %w", err)
		}

		items = make([]crud.Crud, len(rowsList))
		for i, row := range rowsList {
			items[i] = *row.ToCrud()
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return items, nil
}
