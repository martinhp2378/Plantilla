package repository

import (
	"PlantillaInstitucional/configuracion/perfil"
	"PlantillaInstitucional/pagination"
	shared "PlantillaInstitucional/shared/db"
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// Definición de listas blancas (Allow-lists) para seguridad
var (
	colsTexto     = map[string]bool{"nombre": true}
	colsNumericas = map[string]bool{}
	colsUUID      = map[string]bool{}
	// Columnas permitidas estrictamente para ordenamiento
	colsOrden = map[string]bool{"nombre": true}
)

func (r *PerfilRepository) ConsultarPerfiles(ctx context.Context, criterios pagination.CriteriosPagina[perfil.Perfil]) ([]perfil.Perfil, int, error) {
	var perfilsList []perfil.Perfil
	var total int

	err := r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		// 1. Construcción de filtros dinámicos
		filtros, args := shared.BuildDynamicWhere(criterios.FilteredBy, shared.FilterConfig{Texto: colsTexto, Numericas: colsNumericas, UUID: colsUUID}, "p")
		whereClause := ""
		if len(filtros) > 0 {
			whereClause = " WHERE " + strings.Join(filtros, " AND ")
		}

		// 2. Construcción de ordenamiento seguro
		orden, err := shared.BuildOrderBy(criterios.SortedBy, criterios.Direction, colsOrden, "p", "nombre ASC")
		if err != nil {
			return err
		}

		// 3. Query Principal (Datos)
		queryArgs := append(args, criterios.Size, criterios.First)
		query := fmt.Sprintf(SelectPerfilesFiltered, whereClause, orden, len(args)+1, len(args)+2)

		rows, err := tx.Query(ctx, query, queryArgs...)
		if err != nil {
			return fmt.Errorf("ConsultarPerfiles: error ejecutando query: %w", err)
		}

		perfilsList, err = pgx.CollectRows(rows, pgx.RowToStructByName[perfil.Perfil])
		if err != nil {
			return fmt.Errorf("ConsultarPerfiles: error mapeando resultados: %w", err)
		}

		// Optimización: Si es la página 0 y no hay datos, el total es 0. Evitamos un count query.
		if len(perfilsList) == 0 && criterios.First == 0 {
			return nil
		}

		// 4. Query Secundario (Conteo Total para Paginación)
		countQuery := fmt.Sprintf(CountPerfiles, whereClause)
		if err := tx.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
			return fmt.Errorf("ConsultarPerfiles: error contando registros: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return perfilsList, total, nil
}

func (r *PerfilRepository) ConsultarPerfilPorNombre(ctx context.Context, nombre string) (perfil.Perfil, error) {
	var b perfil.Perfil
	err := r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, SelectPerfilXNombre, nombre)
		if err != nil {
			return fmt.Errorf("ConsultarPerfilPorNombre: error de conexión: %w", err)
		}

		b, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[perfil.Perfil])
		if err != nil {
			if err == pgx.ErrNoRows {
				return fmt.Errorf("perfil no encontrado (nombre: %s): %w", nombre, err)
			}
			return fmt.Errorf("ConsultarPerfilPorNombre: %w", err)
		}
		return nil
	})
	if err != nil {
		return perfil.Perfil{}, err
	}

	return b, nil
}

func (r *PerfilRepository) CrearPerfil(ctx context.Context, p perfil.Perfil) error {
	return r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		var newID string
		err := tx.QueryRow(ctx, InsertPerfil,
			p.Nombre,
		).Scan(&newID)

		if err != nil {
			return fmt.Errorf("CrearPerfil: error al insertar: %w", err)
		}
		return nil
	})
}

func (r *PerfilRepository) ActualizarPerfil(ctx context.Context, p perfil.Perfil, nombreOld string) error {
	return r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, UpdatePerfil,
			p.Nombre,
			nombreOld,
		)
		if err != nil {
			return fmt.Errorf("ActualizarPerfil: %w", err)
		}
		if ct.RowsAffected() == 0 {
			return fmt.Errorf("no se encontró el perfil para actualizar (nombre: %s)", nombreOld)
		}
		return nil
	})
}

func (r *PerfilRepository) EliminarPerfil(ctx context.Context, nombre string) error {
	return r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, DeletePerfil, nombre)
		if err != nil {
			return fmt.Errorf("EliminarPerfil: %w", err)
		}
		if ct.RowsAffected() == 0 {
			return fmt.Errorf("no se encontró el perfil para eliminar (nombre: %s)", nombre)
		}
		return nil
	})
}

// --- Helpers Privados ---

// Conversión segura de tipos
