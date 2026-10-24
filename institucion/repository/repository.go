package repository

import (
	"PlantillaInstitucional/institucion"
	"PlantillaInstitucional/pagination"
	shared "PlantillaInstitucional/shared/db"
	"context"
	"fmt"
	"errors"

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

	var personas []*institucion.Persona
	var total int

	err = r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, queryArgs...)
		if err != nil {
			return fmt.Errorf("ConsultarPersonas: error ejecutando query: %w", err)
		}

		// pgx v5: CollectRows itera, escanea y cierra rows automáticamente
		personasList, err := pgx.CollectRows(rows, pgx.RowToStructByName[RowPersona])
		if err != nil {
			return fmt.Errorf("ConsultarPersonas: error mapeando resultados: %w", err)
		}

		// Optimización: Si es la página 0 y no hay datos, el total es 0. Evitamos un count query.
		if len(personasList) == 0 && criterios.First == 0 {
			total = 0
			return nil
		}

		personas = make([]*institucion.Persona, len(personasList))
		for i, p := range personasList {
			personas[i] = p.ToPersona()
		}

		// 4. Query Secundario (Conteo Total para Paginación)
		countQuery := fmt.Sprintf(CountPersonas, whereClause)
		if err := tx.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
			return fmt.Errorf("ConsultarPersonas: error contando registros: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return personas, total, nil
}
func (r *PersonaRepository) ConsultarPersonaPorID(ctx context.Context, uuid_persona uuid.UUID) (*institucion.Persona, error) {
	var pdb RowPersona
	err := r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, SelectPersonaXId, uuid_persona)
		if err != nil {
			return fmt.Errorf("ConsultarPersonaPorID: error de conexión: %w", err)
		}

		pdb, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[RowPersona])
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("persona no encontrada (id: %s): %w", uuid_persona, err)
			}
			return fmt.Errorf("ConsultarPersonaPorID: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	p := pdb.ToPersona()

	return p, nil
}
func (r *PersonaRepository) CrearPersona(ctx context.Context, p institucion.Persona) error {
	err := r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		var newID uuid.UUID
		err := tx.QueryRow(ctx, InsertPersona,
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
	})
	return err
}
func (r *PersonaRepository) ActualizarPersona(ctx context.Context, p institucion.Persona) error {
	err := r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, UpdatePersona,
			p.Nombre,
			p.Apellido1,
			p.Apellido2,
			p.Rfc,
			p.Curp,
			p.ClaveIssemym,
			p.Razon_m,
			p.UuidPersona,
		)
		if err != nil {
			return fmt.Errorf("ActualizarPersona: %w", err)
		}
		if ct.RowsAffected() == 0 {
			return fmt.Errorf("no se encontró la persona para actualizar (id: %s)", p.UuidPersona)
		}
		return nil
	})
	return err
}

func (r *PersonaRepository) ActualizarImagenPerfil(ctx context.Context, uuid_persona uuid.UUID, imagen string) error {
	err := r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, UpdateImagenPerfil, imagen, uuid_persona)
		if err != nil {
			return fmt.Errorf("ActualizarImagenPerfil: %w", err)
		}
		if ct.RowsAffected() == 0 {
			return fmt.Errorf("no se encontró la persona para actualizar su imagen (id: %s)", uuid_persona)
		}
		return nil
	})
	return err
}

func (r *PersonaRepository) EliminarPersona(ctx context.Context, uuid_persona uuid.UUID) error {
	err := r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, DeletePersona, uuid_persona)
		if err != nil {
			return fmt.Errorf("EliminarPersona: %w", err)
		}
		if ct.RowsAffected() == 0 {
			return fmt.Errorf("no se encontró la persona para eliminar (id: %s)", uuid_persona)
		}
		return nil
	})
	return err
}

func (r *PersonaRepository) ObtenerTodasCorto(ctx context.Context) ([]*institucion.Persona, error) {
	var personas []*institucion.Persona
	err := r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, SelectPersonasCorto)
		if err != nil {
			return fmt.Errorf("ObtenerTodasCorto: error ejecutando query: %w", err)
		}

		personasList, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[RowPersona])
		if err != nil {
			return fmt.Errorf("ObtenerTodasCorto: error mapeando resultados: %w", err)
		}

		personas = make([]*institucion.Persona, len(personasList))
		for i, p := range personasList {
			personas[i] = p.ToPersona()
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return personas, nil
}

// ConsultarPersonaSesionPorID obtiene los campos básicos de una persona para hidratar la sesión de usuario.
func (r *PersonaRepository) ConsultarPersonaSesionPorID(ctx context.Context, uuidPersona uuid.UUID) (institucion.Persona, error) {
	if r == nil || r.db == nil || r.db.ReadPool() == nil {
		return institucion.Persona{
			UuidPersona:  uuidPersona,
			ImagenPerfil: "placeholder.png",
		}, nil
	}

	var p institucion.Persona
	p.UuidPersona = uuidPersona
	err := r.db.ReadPool().QueryRow(ctx, SelectPersonaResumenSesionPorID, uuidPersona).Scan(
		&p.Nombre,
		&p.Apellido1,
		&p.Apellido2,
		&p.Curp,
		&p.ClaveIssemym,
		&p.ImagenPerfil,
	)
	if err != nil {
		return institucion.Persona{
			UuidPersona:  uuidPersona,
			ImagenPerfil: "placeholder.png",
		}, fmt.Errorf("ConsultarPersonaSesionPorID: %w", err)
	}
	return p, nil
}

// --- Helpers Privados ---

// Conversión segura de tipos
