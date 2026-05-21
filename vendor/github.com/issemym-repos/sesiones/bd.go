package sesiones

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

type RowSesion struct {
	UuidSesion         uuid.UUID
	UuidPersona        uuid.NullUUID
	TipoSesion         string
	MapaStructs        []byte
	Perfil             Perfil
	FechaCreacion      time.Time
	FechaActualizacion time.Time
	FechaUltimoAcceso  time.Time
	FechaExpiracion    sql.NullTime
	DireccionIP        string
	AgenteUsuario      string
	MensajesFlash      []MensajeFlash
	Huella             string
	InfoDispositivo    InfoDispositivo
	Recuerdame         bool
	TotalPeticiones    sql.NullInt32
	UltimaPeticion     sql.NullTime
	RefresTokenHash    []byte
}

// InsertarSesionEnBd inserta la sesión. Si el objeto ya tiene un UuidSesion
// diferente de uuid.Nil, entonces lo usa como uuid_sesion.
// De otra manera, genera un nuevo uuid V4 antes de insertar.
func (cs *ConstructorSesion) InsertarSesionEnBd(ctx context.Context, sesion *Sesion) error {
	if sesion.UuidSesion == uuid.Nil {
		uuidSesion, err := uuid.NewV4()
		if err != nil {
			return err
		}
		sesion.UuidSesion = uuidSesion
	}
	_, err := cs.conexion.Exec(ctx, `
			insert into sesion.sesion (
				uuid_sesion, uuid_persona, tipo_sesion, mapa_structs, perfil,
				fecha_creacion, fecha_actualizacion,
				fecha_ultimo_acceso, fecha_expiracion,
				direccion_ip, agente_usuario, mensajes_flash, huella,
				info_dispositivo, recuerdame,
				total_peticiones, fecha_ultima_peticion, refresh_token_hash)
			values 
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`,
		sesion.UuidSesion, sesion.UuidPersona, sesion.TipoSesion, sesion.mapaStructs, sesion.Perfil,
		sesion.FechaCreacion, sesion.FechaActualizacion,
		sesion.FechaUltimoAcceso, sesion.FechaExpiracion,
		sesion.DireccionIP, sesion.AgenteUsuario, sesion.MensajesFlash, sesion.Huella,
		sesion.InfoDispositivo, sesion.Recuerdame,
		sesion.TotalPeticiones, sesion.UltimaPeticion, sesion.RefreshTokenHash,
	)
	if err != nil {
		return err
	}
	return nil
}

func (cs *ConstructorSesion) EliminarSesionBd(ctx context.Context, uuidSesion uuid.UUID) error {
	ct, err := cs.conexion.Exec(ctx, `
			delete from sesion.sesion
			where uuid_sesion = $1`,
		uuidSesion)
	if err != nil {
		return fmt.Errorf("error al eliminar sesion: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrorNoExisteSesionEnStore{
			uuidSesion: uuidSesion.String(),
		}
	}
	return nil
}

func (cs *ConstructorSesion) EliminarSesionesPersonaBd(ctx context.Context, sesion *Sesion) error {
	if sesion.UuidPersona == nil {
		return fmt.Errorf("para eliminar sesiones, la sesion con uuid '%s' debía tener UuidPersona (estar firmada)", sesion.UuidSesion.String())
	}
	ct, err := cs.conexion.Exec(ctx, `
			delete from sesion.sesion
			where uuid_persona = $1`,
		sesion.UuidPersona)
	if err != nil || ct.RowsAffected() == 0 {
		return ErrorPersonaNoTieneSesion{
			uuidPersona: sesion.UuidPersona.String(),
		}
	}
	return nil
}

func (cs *ConstructorSesion) ActualizarSesionBd(ctx context.Context, sesion *Sesion) error {
	ct, err := cs.conexion.Exec(ctx, `
			update sesion.sesion set
				uuid_persona = $2, mapa_structs = $3, perfil = $4,
				fecha_creacion = $5, fecha_actualizacion = $6,
				fecha_ultimo_acceso = $7, fecha_expiracion = $8,
				direccion_ip = $9, agente_usuario = $10, mensajes_flash = $11, huella = $12,
				info_dispositivo = $13, recuerdame = $14,
				total_peticiones = $15, fecha_ultima_peticion = $16,
				tipo_sesion = $17, refresh_token_hash = $18
			where uuid_sesion = $1`,
		sesion.UuidSesion, sesion.UuidPersona, sesion.mapaStructs, sesion.Perfil,
		sesion.FechaCreacion, sesion.FechaActualizacion,
		sesion.FechaUltimoAcceso, sesion.FechaExpiracion,
		sesion.DireccionIP, sesion.AgenteUsuario, sesion.MensajesFlash, sesion.Huella,
		sesion.InfoDispositivo, sesion.Recuerdame,
		sesion.TotalPeticiones, sesion.UltimaPeticion,
		sesion.TipoSesion, sesion.RefreshTokenHash,
	)

	if err != nil {
		return fmt.Errorf("error al actualizar sesion: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return ErrorNoExisteSesionEnStore{
			uuidSesion: sesion.UuidSesion.String(),
		}
	}

	return nil
}

func (cs *ConstructorSesion) ActualizarOInsertarSesionEnBd(ctx context.Context, sesion *Sesion) error {
	err := cs.ActualizarSesionBd(ctx, sesion)
	if err != nil && errors.Is(err, ErrorNoExisteSesionEnStore{}) {
		// el uuidSesion no existe probablemente porque se cambió para pedir
		// regenerar la sesion.
		// TODO - política para eliminar otras sesiones?
		// if sesion.UuidPersona != nil {
		// 	cs.EliminarSesionesPersonaBd(ctx, sesion)
		// }
		return cs.InsertarSesionEnBd(ctx, sesion)
	}
	return err
}

// ActualizarOInsertarSesionAnonimaBd actualiza o inserta una Sesion anónima.
// Cierra la sesión para asegurarse que la Sesion sea anónima.
//
// Deprecated: este método ya no se usa debido a que destruye los datos al
// cerrar la sesión y se prefiere poder almacenar datos en sesiones anónimas.
func (cs *ConstructorSesion) ActualizarOInsertarSesionAnonimaBd(ctx context.Context, sesion *Sesion) error {
	// Una sesion anónima no debe tener uuidPersona o datos o FechaExpiracion
	sesion.CerrarSesion()

	ct, err := cs.conexion.Exec(ctx, `
			update sesion.sesion set
				uuid_persona = $2, mapa_structs = $3, perfil = $4,
				fecha_creacion = $5, fecha_actualizacion = $6,
				fecha_ultimo_acceso = $7, fecha_expiracion = $8,
				direccion_ip = $9, agente_usuario = $10, mensajes_flash = $11, huella = $12,
				info_dispositivo = $13, recuerdame = $14,
				total_peticiones = $15, fecha_ultima_peticion = $16,
				tipo_sesion = $17, refresh_token_hash = $18
			where uuid_sesion = $1`,
		sesion.UuidSesion, sesion.UuidPersona, sesion.mapaStructs, sesion.Perfil,
		sesion.FechaCreacion, sesion.FechaActualizacion,
		sesion.FechaUltimoAcceso, sesion.FechaExpiracion,
		sesion.DireccionIP, sesion.AgenteUsuario, sesion.MensajesFlash, sesion.Huella,
		sesion.InfoDispositivo, sesion.Recuerdame,
		sesion.TotalPeticiones, sesion.UltimaPeticion,
		sesion.TipoSesion, sesion.RefreshTokenHash,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return cs.InsertarSesionEnBd(ctx, sesion)
	}

	return nil
}

// Construye una instancia de Sesion a partir de un registro obtenido de la BD
func (cs *ConstructorSesion) SesionBdDeUuid(ctx context.Context,
	uuidSesion uuid.UUID) (*Sesion, error) {

	rows, _ := cs.conexion.Query(ctx, `
			select
				uuid_sesion, uuid_persona, tipo_sesion, mapa_structs, perfil,
				fecha_creacion, fecha_actualizacion,
				fecha_ultimo_acceso, fecha_expiracion,
				direccion_ip, agente_usuario, mensajes_flash, huella,
				info_dispositivo, recuerdame,
				total_peticiones, fecha_ultima_peticion,
				refresh_token_hash
			from sesion.sesion
			where uuid_sesion = $1
	`, uuidSesion)
	rowSesion, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[RowSesion])
	if err != nil {
		return nil, errors.Join(ErrorNoExisteSesionEnStore{
			uuidSesion: uuidSesion.String(),
		}, err)
	}

	return cs.RowBdASesion(&rowSesion)
}

func (cs *ConstructorSesion) SesionBdDeHuella(ctx context.Context,
	huella string) (*Sesion, error) {

	rows, _ := cs.conexion.Query(ctx, `
			select
				uuid_sesion, uuid_persona, tipo_sesion, mapa_structs, perfil,
				fecha_creacion, fecha_actualizacion,
				fecha_ultimo_acceso, fecha_expiracion,
				direccion_ip, agente_usuario, mensajes_flash, huella,
				info_dispositivo, recuerdame,
				total_peticiones, fecha_ultima_peticion,
				refresh_token_hash
			from sesion.sesion
			where huella = $1 and uuid_persona is null
	`, huella)
	rowSesion, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[RowSesion])
	if err != nil {
		return nil, errors.Join(ErrorHuellaNoTieneSesion{
			huella: huella,
		}, err)
	}

	return cs.RowBdASesion(&rowSesion)
}

type resultadoUnmarshallDatos struct {
	mapaStructs        map[string][]any
	atributosAnidados  map[string][]any
	atributosSinAnidar map[string][]any
}

func unmarshallMapaDatos(registroTipos *RegistroTipos, bytesDeMapa []byte) (*resultadoUnmarshallDatos, error) {
	resultado := &resultadoUnmarshallDatos{
		mapaStructs:        map[string][]any{},
		atributosAnidados:  map[string][]any{},
		atributosSinAnidar: map[string][]any{},
	}

	// los json.RawMessage se van a convertir en cada uno de los structs
	// registrados
	var mapaIntermedio map[string][]json.RawMessage
	if err := json.Unmarshal(bytesDeMapa, &mapaIntermedio); err != nil {
		return nil, ErrorUnmarshallJson{
			fmt.Errorf("error al hacer json.Unmarshal en mapaIntermedio: %w", err),
		}
	}

	for llave, sliceRaw := range mapaIntermedio {
		tipo, ok := registroTipos.registro[llave]
		if !ok {
			return nil, ErrorNoExiste{
				llaveStr: llave,
			}
		}
		// el slice de []any es para poder almacenarlo en el mapa de structs
		// resultado
		var sliceConvertido []any
		for i, raw := range sliceRaw {
			instancia := reflect.New(tipo).Interface()
			if err := json.Unmarshal(raw, instancia); err != nil {
				return nil, ErrorUnmarshallJson{
					fmt.Errorf("error al hacer unmarshal para indice %d para llave %s: %w",
						i, llave, err),
				}
			}

			// falta desreferenciar el puntero
			valor := reflect.ValueOf(instancia).Elem().Interface()
			sliceConvertido = append(sliceConvertido, valor)

			// construir mapas de atributos
			registrarCamposStructRecursivo(resultado.atributosAnidados, resultado.atributosSinAnidar, valor, "")
		}

		resultado.mapaStructs[llave] = sliceConvertido
	}

	return resultado, nil
}

func (cs *ConstructorSesion) RowBdASesion(row *RowSesion) (*Sesion, error) {
	resParseoDatos, err := unmarshallMapaDatos(&cs.registroTipos, row.MapaStructs)
	if err != nil {
		return nil, err
	}

	var uuidPersona *uuid.UUID
	if row.UuidPersona.Valid {
		uuidPersona = &row.UuidPersona.UUID
	}
	tipoSesion := TipoSesion(row.TipoSesion)
	if tipoSesion == "" {
		tipoSesion = TipoSesionCookie
	}
	var fechaExpiracion *time.Time
	if row.FechaExpiracion.Valid {
		fechaExpiracion = &row.FechaExpiracion.Time
	}
	var ultimaPeticion *time.Time
	if row.UltimaPeticion.Valid {
		ultimaPeticion = &row.UltimaPeticion.Time
	}

	return &Sesion{
		constructorSesion:  cs,
		UuidSesion:         row.UuidSesion,
		TipoSesion:         tipoSesion,
		mapaStructs:        resParseoDatos.mapaStructs,
		Perfil:             Perfil(row.Perfil),
		atributosAnidados:  resParseoDatos.atributosAnidados,
		atributosSinAnidar: resParseoDatos.atributosSinAnidar,
		conexion:           cs.conexion,
		UuidPersona:        uuidPersona,
		FechaCreacion:      row.FechaCreacion,
		FechaActualizacion: row.FechaActualizacion,
		FechaUltimoAcceso:  row.FechaUltimoAcceso,
		FechaExpiracion:    fechaExpiracion,
		DireccionIP:        row.DireccionIP,
		AgenteUsuario:      row.AgenteUsuario,
		MensajesFlash:      row.MensajesFlash,
		Huella:             row.Huella,
		InfoDispositivo:    row.InfoDispositivo,
		Recuerdame:         row.Recuerdame,
		TotalPeticiones:    int(row.TotalPeticiones.Int32),
		UltimaPeticion:     ultimaPeticion,
		RefreshTokenHash:   row.RefresTokenHash,
	}, nil
}

// ObtenerPoliticas llama a [ObtenerPoliticasSinSesion] con el Perfil de la
// Sesion y la conexion.
// Obtiene las Politicas para realizar una accion sobre un recurso con el
// Perfil de un usuario.
func ObtenerPoliticas(ctx context.Context, sesion *Sesion, recurso string, accion string) ([]Politica, error) {
	return ObtenerPoliticasSinSesion(ctx, sesion.conexion, sesion.Perfil, recurso, accion)
}

// ObtenerPoliticasSinSesion obtiene las Politicas para realizar una accion
// sobre un recurso con el Perfil dado.
// Los campos perfil, recurso, accion pueden ser cadenas vacías o Perfil(""),
// en cuyo caso no se toman en cuenta para el select.
// Puede usarse cuando no es posible obtener la Sesion, por ejemplo dentro de
// ConstructorSesion.fnObtenerDatosSesion.
func ObtenerPoliticasSinSesion(ctx context.Context, conn PgxQueryable,
	perfil Perfil, recurso string, accion string) ([]Politica, error) {

	args := make([]any, 0)
	wheres := make([]string, 0)
	count := 1
	if perfil != "" {
		args = append(args, perfil)
		wheres = append(wheres, fmt.Sprintf(" politica.perfil like $%d ", count))
		count++
	}
	if recurso != "" {
		args = append(args, recurso)
		wheres = append(wheres, fmt.Sprintf(" politica.recurso like $%d ", count))
		count++
	}
	if accion != "" {
		args = append(args, accion)
		wheres = append(wheres, fmt.Sprintf(" politica.accion like $%d ", count))
		count++
	}
	var whereStr string
	if len(args) == 0 {
		whereStr = "true"
	} else {
		whereStr = strings.Join(wheres, " and ")
	}
	rows, _ := conn.Query(ctx, fmt.Sprintf(`
			select
				nombre, descripcion, perfil, recurso, accion, reglas
			from sesion.politica
			where %s
		`, whereStr), args...)
	politicas, err := pgx.CollectRows(rows, pgx.RowToStructByPos[Politica])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.Join(err,
				fmt.Errorf("error al obtener políticas de perfil %s, recurso %s, accion %s", perfil, recurso, accion))
		}
		return nil, err
	}
	return politicas, nil
}
