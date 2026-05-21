package repository

import (
	"PlantillaInstitucional/institucion"
	"PlantillaInstitucional/pagination"
	shared "PlantillaInstitucional/shared/db"
	"context"
	"fmt"

	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

var (
	colsTexto = map[string]bool{"nombre": true, "apellido1": true, "apellido2": true,
		"rfc": true, "curp": true, "claveissemym": true,
	}
	colsNumericas = map[string]bool{}
	colsUUID      = map[string]bool{"uuid_persona": true}
	// Columnas permitidas estrictamente para ordenamiento
	colsOrden = map[string]bool{
		"uuid_persona": true, "nombre": true, "apellido1": true, "apellido2": true,
		"rfc": true, "curp": true, "claveissemym": true,
	}
)

func (r *PersonaRepository) ConsultarPersonas(ctx context.Context, criterios pagination.CriteriosPagina[institucion.Persona]) ([]*institucion.Persona, int, error) {
	// pgxpool maneja el Acquire/Release internamente. Es más seguro y eficiente.
	db := r.db.ReadPool

	// 1. Construcción de filtros dinámicos
	filtros, args := shared.BuildDynamicWhere(criterios.FilteredBy, shared.FilterConfig{Texto: colsTexto, Numericas: colsNumericas, UUID: colsUUID}, "p")
	whereClause := ""
	if len(filtros) > 0 {
		whereClause = " WHERE " + strings.Join(filtros, " AND ")
	}

	// 2. Construcción de ordenamiento seguro
	orden, err := shared.BuildOrderBy(criterios.SortedBy, criterios.Direction, colsOrden, "", "nombre ASC")
	if err != nil {
		return nil, 0, err
	}

	// 3. Query Principal (Datos)
	// Los argumentos finales son: [Filtros..., Limit, Offset]
	queryArgs := append(args, criterios.Size, criterios.First)
	query := fmt.Sprintf(SelectPersonasFiltered, whereClause, orden, len(args)+1, len(args)+2)

	rows, err := db.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("ConsultarPersonas: error ejecutando query: %w", err)
	}

	// pgx v5: CollectRows itera, escanea y cierra rows automáticamente
	personasList, err := pgx.CollectRows(rows, pgx.RowToStructByName[RowPersona])
	if err != nil {
		return nil, 0, fmt.Errorf("ConsultarPersonas: error mapeando resultados: %w", err)
	}

	// Optimización: Si es la página 0 y no hay datos, el total es 0. Evitamos un count query.
	if len(personasList) == 0 && criterios.First == 0 {
		return nil, 0, nil
	}

	personas := make([]*institucion.Persona, len(personasList))
	for i, p := range personasList {
		personas[i] = p.ToPersona()
	}

	// 4. Query Secundario (Conteo Total para Paginación)
	var total int
	countQuery := fmt.Sprintf(CountPersonas, whereClause)
	if err := db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("ConsultarPersonas: error contando registros: %w", err)
	}

	return personas, total, nil
}
func (r *PersonaRepository) ConsultarPersonaPorID(ctx context.Context, uuid_persona uuid.UUID) (*institucion.Persona, error) {
	rows, err := r.db.ReadPool.Query(ctx, SelectPersonaXId, uuid_persona)
	if err != nil {
		return nil, fmt.Errorf("ConsultarPersonaPorID: error de conexión: %w", err)
	}

	pdb, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[RowPersona])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("persona no encontrada (id: %s): %w", uuid_persona, err)
		}
		return nil, fmt.Errorf("ConsultarPersonaPorID: %w", err)
	}

	p := pdb.ToPersona()

	return p, nil
}
func (r *PersonaRepository) CrearPersona(ctx context.Context, p institucion.Persona) error {
	conn, err := r.db.WritePool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error obteniendo conexión: %w", err)
	}
	defer conn.Release()

	if err := shared.SetAuditSession(ctx, conn); err != nil {
		return fmt.Errorf("error configurando auditoría: %w", err)
	}

	// Usamos QueryRow con RETURNING id para verificar que se insertó correctamente
	var newID uuid.UUID
	err = conn.QueryRow(ctx, InsertPersona,
		p.Nombre,
		p.Apellido1,
		p.Apellido2,
		p.Rfc,
		p.Curp,
		p.ClaveIssemym,
		p.Razon_i,
	).Scan(&newID)

	if err != nil {
		return fmt.Errorf("CrearPersona: error al insertar: %w", err)
	}
	return nil
}
func (r *PersonaRepository) ActualizarPersona(ctx context.Context, p institucion.Persona) error {
	conn, err := r.db.WritePool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error obteniendo conexión: %w", err)
	}
	defer conn.Release()

	if err := shared.SetAuditSession(ctx, conn); err != nil {
		return fmt.Errorf("error configurando auditoría: %w", err)
	}

	ct, err := conn.Exec(ctx, UpdatePersona,
		p.Nombre,
		p.Apellido1,
		p.Apellido2,
		p.Rfc,
		p.Curp,
		p.ClaveIssemym,
		p.UuidPersona,
	)
	if err != nil {
		return fmt.Errorf("ActualizarPersona: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("no se encontró la persona para actualizar (id: %s)", p.UuidPersona)
	}
	return nil
}

func (r *PersonaRepository) ActualizarImagenPerfil(ctx context.Context, uuid_persona uuid.UUID, imagen string) error {
	conn, err := r.db.WritePool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error obteniendo conexión: %w", err)
	}
	defer conn.Release()

	if err := shared.SetAuditSession(ctx, conn); err != nil {
		return fmt.Errorf("error configurando auditoría: %w", err)
	}

	ct, err := conn.Exec(ctx, UpdateImagenPerfil, imagen, uuid_persona)
	if err != nil {
		return fmt.Errorf("ActualizarImagenPerfil: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("no se encontró la persona para actualizar su imagen (id: %s)", uuid_persona)
	}
	return nil
}

func (r *PersonaRepository) EliminarPersona(ctx context.Context, uuid_persona uuid.UUID) error {
	conn, err := r.db.WritePool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error obteniendo conexión: %w", err)
	}
	defer conn.Release()

	if err := shared.SetAuditSession(ctx, conn); err != nil {
		return fmt.Errorf("error configurando auditoría: %w", err)
	}

	ct, err := conn.Exec(ctx, DeletePersona, uuid_persona)
	if err != nil {
		return fmt.Errorf("EliminarPersona: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("no se encontró la persona para eliminar (id: %s)", uuid_persona)
	}
	return nil
}

func (r *PersonaRepository) ObtenerTodasCorto(ctx context.Context) ([]*institucion.Persona, error) {
	db := r.db.ReadPool
	rows, err := db.Query(ctx, SelectPersonasCorto)
	if err != nil {
		return nil, fmt.Errorf("ObtenerTodasCorto: error ejecutando query: %w", err)
	}

	personasList, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[RowPersona])
	if err != nil {
		return nil, fmt.Errorf("ObtenerTodasCorto: error mapeando resultados: %w", err)
	}

	personas := make([]*institucion.Persona, len(personasList))
	for i, p := range personasList {
		personas[i] = p.ToPersona()
	}

	return personas, nil
}

// --- Helpers Privados ---

// Conversión segura de tipos
