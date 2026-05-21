package sesiones

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/redis/go-redis/v9"
)

// StoreSesion es una interfaz que permite el uso de Postgres o Redis para
// almacenar y recuperar instancias de Sesion.
type StoreSesion interface {
	// Almacenar en el Store
	Almacenar(ctx context.Context, sesion *Sesion) error
	// Recuperar sesion por uuidSesion
	Recuperar(ctx context.Context, uuidSesion uuid.UUID) (*Sesion, error)
	// RecuperarPorHuella es para sesiones anonimas
	RecuperarPorHuella(ctx context.Context, huella string) (*Sesion, error)
	// Actualizar una sesion existente
	Actualizar(ctx context.Context, sesion *Sesion) error
	// Eliminar una sesion por UUID
	Eliminar(ctx context.Context, uuidSesion uuid.UUID) error
	// Eliminar Sesiones de una Persona
	EliminarSesionesPersona(ctx context.Context, uuidPersona uuid.UUID) error
	// Almacenar o Actualizar una Sesion
	AlmacenarOActualizar(ctx context.Context, sesion *Sesion) error
}

type StorePostgresSesion struct {
	cs *ConstructorSesion
}

func NuevaStorePostgres(cs *ConstructorSesion) *StorePostgresSesion {
	return &StorePostgresSesion{cs: cs}
}

func (p *StorePostgresSesion) Almacenar(ctx context.Context, sesion *Sesion) error {
	return p.cs.InsertarSesionEnBd(ctx, sesion)
}

func (p *StorePostgresSesion) Recuperar(ctx context.Context, uuidSesion uuid.UUID) (*Sesion, error) {
	return p.cs.SesionBdDeUuid(ctx, uuidSesion)
}

func (p *StorePostgresSesion) RecuperarPorHuella(ctx context.Context, huella string) (*Sesion, error) {
	return p.cs.SesionBdDeHuella(ctx, huella)
}

func (p *StorePostgresSesion) Actualizar(ctx context.Context, sesion *Sesion) error {
	return p.cs.ActualizarSesionBd(ctx, sesion)
}

func (p *StorePostgresSesion) Eliminar(ctx context.Context, uuidSesion uuid.UUID) error {
	return p.cs.EliminarSesionBd(ctx, uuidSesion)
}

func (p *StorePostgresSesion) EliminarSesionesPersona(ctx context.Context, uuidPersona uuid.UUID) error {
	tempSesion := &Sesion{constructorSesion: p.cs, UuidPersona: &uuidPersona}
	return p.cs.EliminarSesionesPersonaBd(ctx, tempSesion)
}

func (p *StorePostgresSesion) AlmacenarOActualizar(ctx context.Context, sesion *Sesion) error {
	return p.cs.ActualizarOInsertarSesionEnBd(ctx, sesion)
}

type StoreRedisSesion struct {
	cs     *ConstructorSesion
	client *redis.Client
}

func NuevoStoreRedis(cs *ConstructorSesion, client *redis.Client) *StoreRedisSesion {
	return &StoreRedisSesion{
		cs:     cs,
		client: client,
	}
}

// datosSesionRedis son los datos de una sesion que se almacenan en un Store
// de Redis.
type datosSesionRedis struct {
	UuidSesion         uuid.UUID       `json:"uuid_sesion"`
	UuidPersona        *uuid.UUID      `json:"uuid_persona"`
	TipoSesion         string          `json:"tipo_sesion"`
	MapaStructs        []byte          `json:"mapa_structs"`
	Perfil             Perfil          `json:"perfil"`
	FechaCreacion      time.Time       `json:"fecha_creacion"`
	FechaActualizacion time.Time       `json:"fecha_actualizacion"`
	FechaUltimoAcceso  time.Time       `json:"fecha_ultimo_acceso"`
	FechaExpiracion    *time.Time      `json:"fecha_expiracion"`
	DireccionIP        string          `json:"direccion_ip"`
	AgenteUsuario      string          `json:"agente_usuario"`
	MensajesFlash      []MensajeFlash  `json:"mensajes_flash"`
	Huella             string          `json:"huella"`
	InfoDispositivo    InfoDispositivo `json:"info_dispositivo"`
	Recuerdame         bool            `json:"recuerdame"`
	TotalPeticiones    int             `json:"total_peticiones"`
	UltimaPeticion     *time.Time      `json:"ultima_peticion"`
	RefreshTokenHash   []byte          `json:"refresh_token_hash"`
}

func (r *StoreRedisSesion) sesionADatos(sesion *Sesion) *datosSesionRedis {
	mapaJson, err := json.Marshal(sesion.mapaStructs)
	if err != nil {
		mapaJson = []byte("{}")
	}

	return &datosSesionRedis{
		UuidSesion:         sesion.UuidSesion,
		UuidPersona:        sesion.UuidPersona,
		TipoSesion:         string(sesion.TipoSesion),
		MapaStructs:        mapaJson,
		Perfil:             sesion.Perfil,
		FechaCreacion:      sesion.FechaCreacion,
		FechaActualizacion: sesion.FechaActualizacion,
		FechaUltimoAcceso:  sesion.FechaUltimoAcceso,
		FechaExpiracion:    sesion.FechaExpiracion,
		DireccionIP:        sesion.DireccionIP,
		AgenteUsuario:      sesion.AgenteUsuario,
		MensajesFlash:      sesion.MensajesFlash,
		Huella:             sesion.Huella,
		InfoDispositivo:    sesion.InfoDispositivo,
		Recuerdame:         sesion.Recuerdame,
		TotalPeticiones:    sesion.TotalPeticiones,
		UltimaPeticion:     sesion.UltimaPeticion,
		RefreshTokenHash:   sesion.RefreshTokenHash,
	}
}

// datosASesion reconstruye una Sesion a partir de los datos obtenidos de un
// Store.
func (r *StoreRedisSesion) datosRedisASesion(datos *datosSesionRedis) (*Sesion, error) {
	resParseoDatos, err := unmarshallMapaDatos(&r.cs.registroTipos, datos.MapaStructs)
	if err != nil {
		return nil, err
	}
	tipoSesion := TipoSesion(datos.TipoSesion)
	if tipoSesion == "" {
		tipoSesion = TipoSesionCookie
	}

	return &Sesion{
		constructorSesion:  r.cs,
		UuidSesion:         datos.UuidSesion,
		UuidPersona:        datos.UuidPersona,
		TipoSesion:         tipoSesion,
		mapaStructs:        resParseoDatos.mapaStructs,
		Perfil:             datos.Perfil,
		atributosAnidados:  resParseoDatos.atributosAnidados,
		atributosSinAnidar: resParseoDatos.atributosSinAnidar,
		conexion:           r.cs.conexion,
		FechaCreacion:      datos.FechaCreacion,
		FechaActualizacion: datos.FechaActualizacion,
		FechaUltimoAcceso:  datos.FechaUltimoAcceso,
		FechaExpiracion:    datos.FechaExpiracion,
		DireccionIP:        datos.DireccionIP,
		AgenteUsuario:      datos.AgenteUsuario,
		MensajesFlash:      datos.MensajesFlash,
		Huella:             datos.Huella,
		InfoDispositivo:    datos.InfoDispositivo,
		Recuerdame:         datos.Recuerdame,
		TotalPeticiones:    datos.TotalPeticiones,
		UltimaPeticion:     datos.UltimaPeticion,
		RefreshTokenHash:   datos.RefreshTokenHash,
	}, nil
}

func (r *StoreRedisSesion) getLlaveSesion(uuidSesion uuid.UUID) string {
	return fmt.Sprintf("sesion:%s", uuidSesion.String())
}

func (r *StoreRedisSesion) getLlaveHuella(huella string) string {
	return fmt.Sprintf("sesion:huella:%s", huella)
}

func (r *StoreRedisSesion) getLlavePersona(uuidPersona uuid.UUID) string {
	return fmt.Sprintf("sesion:persona:%s", uuidPersona.String())
}

func (r *StoreRedisSesion) calcularTTL(sesion *Sesion) time.Duration {
	if sesion.FechaExpiracion == nil {
		// para sesiones anónimas usar la duración inactiva
		return r.cs.config.DuracionInactiva
	}

	ttl := time.Until(*sesion.FechaExpiracion)
	if ttl <= 0 {
		// Si el TTL es negativo o cero (por latencia, procesamiento lento, o desfase de reloj),
		// usar DuracionInactiva completa en lugar de solo 1 minuto.
		// Esto previene que sesiones válidas expiren prematuramente en Redis.
		slog.Warn("TTL negativo o cero detectado al guardar en Redis, usando DuracionInactiva",
			"ttl_calculado", ttl,
			"FechaExpiracion", sesion.FechaExpiracion,
			"ahora", time.Now(),
			"uuid_sesion", sesion.UuidSesion)
		return r.cs.config.DuracionInactiva
	}
	return ttl
}

func (r *StoreRedisSesion) Almacenar(ctx context.Context, sesion *Sesion) error {
	if sesion.UuidSesion == uuid.Nil {
		uuidSesion, err := uuid.NewV4()
		if err != nil {
			return err
		}
		sesion.UuidSesion = uuidSesion
	}

	datos := r.sesionADatos(sesion)
	datosJson, err := json.Marshal(datos)
	if err != nil {
		return fmt.Errorf("error al hacer marshalling de la Sesion: %w", err)
	}

	key := r.getLlaveSesion(sesion.UuidSesion)
	ttl := r.calcularTTL(sesion)

	err = r.client.Set(ctx, key, datosJson, ttl).Err()
	if err != nil {
		return fmt.Errorf("error al almacenar Sesion en Redis: %w", err)
	}

	// Relacion huella -> Sesion para sesiones anonimas.
	if sesion.UuidPersona == nil && sesion.Huella != "" {
		llaveHuella := r.getLlaveHuella(sesion.Huella)
		err = r.client.Set(ctx, llaveHuella, sesion.UuidSesion.String(), ttl).Err()
		if err != nil {
			return fmt.Errorf("error al almacenar Sesion por huella: %w", err)
		}
	}

	// Relacion uuidPersona -> Sesion para sesiones firmadas.
	if sesion.UuidPersona != nil {
		llavePersona := r.getLlavePersona(*sesion.UuidPersona)
		err = r.client.SAdd(ctx, llavePersona, sesion.UuidSesion.String()).Err()
		if err != nil {
			return fmt.Errorf("error al almacenar Sesion por uuidPersona: %w", err)
		}
		err = r.client.Expire(ctx, llavePersona, ttl).Err()
		if err != nil {
			return fmt.Errorf("error al colocar TTL de Sesion firmada: %w", err)
		}
	}

	return nil
}

func (r *StoreRedisSesion) Recuperar(ctx context.Context, uuidSesion uuid.UUID) (*Sesion, error) {
	llave := r.getLlaveSesion(uuidSesion)
	datosJson, err := r.client.Get(ctx, llave).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrorNoExisteSesionEnStore{
				uuidSesion: uuidSesion.String(),
			}
		}
		return nil, fmt.Errorf("error al recuperar Sesion de Redis: %w", err)
	}

	var datos datosSesionRedis
	err = json.Unmarshal([]byte(datosJson), &datos)
	if err != nil {
		return nil, fmt.Errorf("error al hacer unmarshall de objeto de Redis: %w", err)
	}

	return r.datosRedisASesion(&datos)
}

func (r *StoreRedisSesion) RecuperarPorHuella(ctx context.Context, huella string) (*Sesion, error) {
	llaveHuella := r.getLlaveHuella(huella)
	uuidStr, err := r.client.Get(ctx, llaveHuella).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrorHuellaNoTieneSesion{
				huella: huella,
			}
		}
		return nil, fmt.Errorf("error al obtener mapeo de huella: %w", err)
	}

	uuidSesion, err := uuid.FromString(uuidStr)
	if err != nil {
		return nil, fmt.Errorf("error al parsear UUID de la Sesion de Redis: %w", err)
	}

	return r.Recuperar(ctx, uuidSesion)
}

func (r *StoreRedisSesion) Actualizar(ctx context.Context, sesion *Sesion) error {
	llave := r.getLlaveSesion(sesion.UuidSesion)
	exists, err := r.client.Exists(ctx, llave).Result()
	if err != nil {
		return fmt.Errorf("error al verificar la existencia de la Sesion en Redis: %w", err)
	}
	if exists == 0 {
		return ErrorNoExisteSesionEnStore{
			uuidSesion: sesion.UuidSesion.String(),
		}
	}

	return r.Almacenar(ctx, sesion)
}

func (r *StoreRedisSesion) Eliminar(ctx context.Context, uuidSesion uuid.UUID) error {
	llaveSesion := r.getLlaveSesion(uuidSesion)
	err := r.client.Del(ctx, llaveSesion).Err()
	if err != nil {
		return fmt.Errorf("error al eliminar la Sesion %s de Redis: %w", uuidSesion.String(), err)
	}
	return nil
}

func (r *StoreRedisSesion) EliminarSesionesPersona(ctx context.Context, uuidPersona uuid.UUID) error {
	llave := r.getLlavePersona(uuidPersona)
	sessionUUIDs, err := r.client.SMembers(ctx, llave).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrorPersonaNoTieneSesion{
				uuidPersona: uuidPersona.String(),
			}
		}
		return fmt.Errorf("error al recuperar Sesiones de la llave '%s': %w", llave, err)
	}

	if len(sessionUUIDs) == 0 {
		return ErrorPersonaNoTieneSesion{
			uuidPersona: uuidPersona.String(),
		}
	}

	// Eliminar todas las sesiones de la persona
	for _, uuidStr := range sessionUUIDs {
		llaveSesion := fmt.Sprintf("sesion:%s", uuidStr)
		err = r.client.Del(ctx, llaveSesion).Err()
		if err != nil {
			return fmt.Errorf("error al eliminar la Sesion %s: %w", uuidStr, err)
		}
	}

	// Eliminar el mapeo de la persona
	err = r.client.Del(ctx, llave).Err()
	if err != nil {
		return fmt.Errorf("error al eliminar el mapeo de la llave '%s': %w", llave, err)
	}

	return nil
}

func (r *StoreRedisSesion) AlmacenarOActualizar(ctx context.Context, sesion *Sesion) error {
	return r.Almacenar(ctx, sesion)
}
