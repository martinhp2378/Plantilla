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
	// pgxpool maneja el Acquire/Release internamente. Es más seguro y eficiente.
	db := r.db.ReadPool

	// 1. Construcción de filtros dinámicos
	filtros, args := shared.BuildDynamicWhere(criterios.FilteredBy, shared.FilterConfig{Texto: colsTexto, Numericas: colsNumericas, UUID: colsUUID}, "p")
	whereClause := ""
	if len(filtros) > 0 {
		whereClause = " WHERE " + strings.Join(filtros, " AND ")
	}

	// 2. Construcción de ordenamiento seguro
	orden, err := shared.BuildOrderBy(criterios.SortedBy, criterios.Direction, colsOrden, "p", "nombre ASC")
	if err != nil {
		return nil, 0, err
	}

	// 3. Query Principal (Datos)
	// Los argumentos finales son: [Filtros..., Limit, Offset]
	queryArgs := append(args, criterios.Size, criterios.First)
	query := fmt.Sprintf(SelectPerfilesFiltered, whereClause, orden, len(args)+1, len(args)+2)

	rows, err := db.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("ConsultarPerfiles: error ejecutando query: %w", err)
	}

	// pgx v5: CollectRows itera, escanea y cierra rows automáticamente
	perfilsList, err := pgx.CollectRows(rows, pgx.RowToStructByName[perfil.Perfil])
	if err != nil {
		return nil, 0, fmt.Errorf("ConsultarPerfiles: error mapeando resultados: %w", err)
	}

	// Optimización: Si es la página 0 y no hay datos, el total es 0. Evitamos un count query.
	if len(perfilsList) == 0 && criterios.First == 0 {
		return perfilsList, 0, nil
	}

	// 4. Query Secundario (Conteo Total para Paginación)
	var total int
	countQuery := fmt.Sprintf(CountPerfiles, whereClause)
	if err := db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("ConsultarPerfiles: error contando registros: %w", err)
	}

	return perfilsList, total, nil
}

func (r *PerfilRepository) ConsultarPerfilPorNombre(ctx context.Context, nombre string) (perfil.Perfil, error) {
	rows, err := r.db.ReadPool.Query(ctx, SelectPerfilXNombre, nombre)
	if err != nil {
		return perfil.Perfil{}, fmt.Errorf("ConsultarPerfilPorNombre: error de conexión: %w", err)
	}

	b, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[perfil.Perfil])
	if err != nil {
		if err == pgx.ErrNoRows {
			return perfil.Perfil{}, fmt.Errorf("perfil no encontrado (nombre: %s): %w", nombre, err)
		}
		return perfil.Perfil{}, fmt.Errorf("ConsultarPerfilPorNombre: %w", err)
	}

	return b, nil
}

func (r *PerfilRepository) CrearPerfil(ctx context.Context, p perfil.Perfil) error {
	conn, err := r.db.WritePool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error obteniendo conexión: %w", err)
	}
	defer conn.Release()

	if err := shared.SetAuditSession(ctx, conn); err != nil {
		return fmt.Errorf("error configurando auditoría: %w", err)
	}

	// Usamos QueryRow con RETURNING id para verificar que se insertó correctamente
	var newID string
	err = conn.QueryRow(ctx, InsertPerfil,
		p.Nombre,
	).Scan(&newID)

	if err != nil {
		return fmt.Errorf("CrearPerfil: error al insertar: %w", err)
	}
	return nil
}

func (r *PerfilRepository) ActualizarPerfil(ctx context.Context, p perfil.Perfil, nombreOld string) error {
	conn, err := r.db.WritePool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error obteniendo conexión: %w", err)
	}
	defer conn.Release()

	if err := shared.SetAuditSession(ctx, conn); err != nil {
		return fmt.Errorf("error configurando auditoría: %w", err)
	}

	ct, err := conn.Exec(ctx, UpdatePerfil,
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
}

func (r *PerfilRepository) EliminarPerfil(ctx context.Context, nombre string) error {
	conn, err := r.db.WritePool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error obteniendo conexión: %w", err)
	}
	defer conn.Release()

	if err := shared.SetAuditSession(ctx, conn); err != nil {
		return fmt.Errorf("error configurando auditoría: %w", err)
	}

	ct, err := conn.Exec(ctx, DeletePerfil, nombre)
	if err != nil {
		return fmt.Errorf("EliminarPerfil: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("no se encontró el perfil para eliminar (nombre: %s)", nombre)
	}
	return nil
}

// --- Helpers Privados ---

// Conversión segura de tipos
