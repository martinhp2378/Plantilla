package repository

import (
	"PlantillaInstitucional/configuracion/usuario"
	"PlantillaInstitucional/pagination"
	shared "PlantillaInstitucional/shared/db"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

// Definición de listas blancas (Allow-lists) para seguridad
var (
	colsTexto     = map[string]bool{"email": true, "password": true, "perfil": true}
	colsNumericas = map[string]bool{}
	colsUUID      = map[string]bool{"uuid_persona": true}
	// Columnas permitidas estrictamente para ordenamiento
	colsOrden = map[string]bool{"uuid_persona": true, "email": true, "password": true, "perfil": true, "nombre_completo": true}
)

func (r *UsuarioRepository) ConsultarUsuarios(ctx context.Context, criterios pagination.CriteriosPagina[usuario.Usuario]) ([]usuario.Usuario, int, error) {
	// 1. Construcción de filtros dinámicos
	filtros, args := buildDynamicWhere(criterios.FilteredBy)
	whereClause := ""
	if len(filtros) > 0 {
		whereClause = " WHERE " + strings.Join(filtros, " AND ")
	}

	// 2. Construcción de ordenamiento seguro
	tableAlias := "u"
	if criterios.SortedBy == "nombre_completo" {
		tableAlias = ""
	}
	orden, err := shared.BuildOrderBy(criterios.SortedBy, criterios.Direction, colsOrden, tableAlias, "u.email ASC")
	if err != nil {
		return nil, 0, err
	}

	// 3. Query Principal (Datos)
	// Los argumentos finales son: [Filtros..., Limit, Offset]
	queryArgs := append(args, criterios.Size, criterios.First)
	query := fmt.Sprintf(SelectUsuariosFiltered, whereClause, orden, len(args)+1, len(args)+2)

	var usuariosList []usuario.Usuario
	var total int

	err = r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, queryArgs...)
		if err != nil {
			return fmt.Errorf("ConsultarUsuarios: error ejecutando query: %w", err)
		}

		rowsList, err := pgx.CollectRows(rows, pgx.RowToStructByName[RowUsuario])
		if err != nil {
			return fmt.Errorf("ConsultarUsuarios: error mapeando resultados: %w", err)
		}

		usuariosList = make([]usuario.Usuario, len(rowsList))
		for i, r := range rowsList {
			usuariosList[i] = r.ToUsuario()
		}

		// Optimización: Si es la página 0 y no hay datos, el total es 0. Evitamos un count query.
		if len(usuariosList) == 0 && criterios.First == 0 {
			total = 0
			return nil
		}

		// 4. Query Secundario (Conteo Total para Paginación)
		countQuery := fmt.Sprintf(CountUsuarios, whereClause)
		if err := tx.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
			return fmt.Errorf("ConsultarUsuarios: error contando registros: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return usuariosList, total, nil
}

func (r *UsuarioRepository) ConsultarUsuarioPorID(ctx context.Context, id uuid.UUID) (usuario.Usuario, error) {
	var row RowUsuario
	err := r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, SelectUsuarioXId, id)
		if err != nil {
			return fmt.Errorf("ConsultarUsuarioPorID: error de conexión: %w", err)
		}

		row, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[RowUsuario])
		if err != nil {
			if err == pgx.ErrNoRows {
				return fmt.Errorf("usuario no encontrado (id: %s): %w", id, err)
			}
			return fmt.Errorf("ConsultarUsuarioPorID: %w", err)
		}
		return nil
	})

	if err != nil {
		return usuario.Usuario{}, err
	}

	return row.ToUsuario(), nil
}

func (r *UsuarioRepository) CrearUsuario(ctx context.Context, b usuario.Usuario) error {
	err := r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		var newID uuid.UUID
		err := tx.QueryRow(ctx, InsertUsuario,
			b.UuidPersona,
			b.Email,
			b.Password,
			b.Perfil,
			b.Razon_i,
		).Scan(&newID)

		if err != nil {
			return fmt.Errorf("CrearUsuario: error al insertar: %w", err)
		}
		return nil
	})
	return err
}

func (r *UsuarioRepository) ActualizarUsuario(ctx context.Context, b usuario.Usuario) error {
	err := r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, UpdateUsuario,
			b.Email,
			b.Password,
			b.Perfil,
			b.Razon_m,
			b.UuidPersona,
		)
		if err != nil {
			return fmt.Errorf("ActualizarUsuario: %w", err)
		}
		if ct.RowsAffected() == 0 {
			return fmt.Errorf("no se encontró el usuario para actualizar (id: %s)", b.UuidPersona)
		}
		return nil
	})
	return err
}

func (r *UsuarioRepository) ActualizarPassword(ctx context.Context, uuid_persona uuid.UUID, password string) error {
	err := r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, UpdatePasswordUsuario, password, uuid_persona)
		if err != nil {
			return fmt.Errorf("ActualizarPassword: %w", err)
		}
		if ct.RowsAffected() == 0 {
			return fmt.Errorf("no se encontró el usuario para actualizar la contraseña (id: %s)", uuid_persona)
		}
		return nil
	})
	return err
}


func (r *UsuarioRepository) EliminarUsuario(ctx context.Context, uuid_persona uuid.UUID) error {
	err := r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, DeleteUsuario, uuid_persona)
		if err != nil {
			return fmt.Errorf("EliminarUsuario: %w", err)
		}
		if ct.RowsAffected() == 0 {
			return fmt.Errorf("no se encontró el usuario para eliminar (id: %s)", uuid_persona)
		}
		return nil
	})
	return err
}

// --- Helpers Privados ---

func buildDynamicWhere(filtros map[string]any) ([]string, []any) {
	var condiciones []string
	var args []any
	argID := 1

	for campo, valor := range filtros {
		if valor == nil {
			continue
		}

		// CASO ESPECIAL: nombre_completo (Campo calculado con JOIN)
		if campo == "nombre_completo" {
			strVal := fmt.Sprint(valor)
			if strVal != "" {
				condiciones = append(condiciones, fmt.Sprintf("CONCAT_WS(' ', p.nombre, p.apellido1, p.apellido2) ILIKE $%d", argID))
				args = append(args, "%"+strVal+"%")
				argID++
			}
			continue
		}

		// CASO ESTÁNDAR: Columnas en tabla 'usuario' (u)
		if colsTexto[campo] {
			strVal := fmt.Sprint(valor)
			if strVal != "" {
				condiciones = append(condiciones, fmt.Sprintf("u.%s ILIKE $%d", campo, argID))
				args = append(args, "%"+strVal+"%")
				argID++
			}
			continue
		}

		// Igualdad para UUIDs (u.uuid_persona)
		if colsUUID[campo] {
			if uuidVal, ok := toUUID(valor); ok && uuidVal != uuid.Nil {
				condiciones = append(condiciones, fmt.Sprintf("u.%s = $%d", campo, argID))
				args = append(args, uuidVal)
				argID++
			}
			continue
		}
	}
	return condiciones, args
}

// Conversión segura de tipos
func toInt(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case float64:
		return int(val), true
	case string:
		if val == "" {
			return 0, false
		}
		i, err := strconv.Atoi(val)
		return i, err == nil
	default:
		return 0, false
	}
}

func toUUID(v any) (uuid.UUID, bool) {
	switch val := v.(type) {
	case uuid.UUID:
		return val, true
	case string:
		if val == "" {
			return uuid.Nil, false
		}
		uid, err := shared.ParseUUID(val)
		return uid, err == nil
	default:
		return uuid.Nil, false
	}
}
