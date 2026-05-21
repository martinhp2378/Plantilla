package sesiones

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
)

type PgxQueryable interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// RegistroTipos es un mapa de llave -> reflect.Type que se utiliza para que
// un ConstructorSesion sepa reconstruir los datos de una Sesion a partir
// de JSON obtenido de la BD.
type RegistroTipos struct {
	registro map[string]reflect.Type
}

// NuevoRegistroTipos se utiliza para crear un RegistroTipos.
func NuevoRegistroTipos() *RegistroTipos {
	return &RegistroTipos{
		registro: make(map[string]reflect.Type),
	}
}

// RegistrarTipo genera una llave de tipo "packageName.TipoStruct" y
// registra el tipo de la instancia con esa llave. El tipo se obtiene
// con el módulo reflect.
// La instancia puede ser "vacía". Es decir, creada de la forma:
// var instancia TipoStruct
// Regresa [ErrorSeEsperabaStruct] cuando la instancia no es de un struct.
// No se permite registrar "sesiones.Sesion". Los campos de la Sesion se
// pueden recuperar en las políticas con este prefijo sin tener que
// registrar o agregar una instancia de Sesion a los datos.
func (r *RegistroTipos) RegistrarTipo(instancia any) (string, error) {
	v := reflect.ValueOf(instancia)
	tipo := v.Type()
	if tipo.Kind() != reflect.Struct {
		return "", ErrorSeEsperabaStruct{
			instanciaStr: fmt.Sprintf("%+v", instancia),
		}
	}
	// por default un struct se guarda con la llave
	// packageName.structType
	llave := tipo.String()
	if llave == "sesiones.Sesion" {
		return "", ErrorNoSePermiteRegistrarSesion
	}
	r.registro[llave] = tipo
	return llave, nil
}

// RegistrarTipoLlave registra el tipo de dato de una instancia de un struct
// con la llave dada. Esta llave debe coincidir con la llave usada en
// [AgregarInstanciaLlave] para que el constructor sepa con qué tipo de dato
// hacer el unmarshall de la instancia.
// Regresa [ErrorSeEsperabaStruct] cuando la instancia no es de un struct.
// No se permite registrar "sesiones.Sesion". Los campos de la Sesion se
// pueden recuperar en las políticas con este prefijo sin tener que
// registrar o agregar una instancia de Sesion a los datos.
func (r *RegistroTipos) RegistrarTipoLlave(llave string, instancia any) error {
	if llave == "sesiones.Sesion" {
		return ErrorNoSePermiteRegistrarSesion
	}
	tipo := reflect.TypeOf(instancia)
	if tipo.Kind() != reflect.Struct {
		return ErrorSeEsperabaStruct{
			instanciaStr: fmt.Sprintf("%+v", instancia),
		}
	}
	r.registro[llave] = tipo
	return nil
}

// ConstructorSesion se utiliza para construir sesiones y administrarlas.
// La mayoría de las funciones que modifican, crean o destruyen una sesión
// son métodos sobre ConstructorSesion y no sobre Sesion.
// Se recomienda usar [NuevoConstructor] para crear una instancia.
// Tiene los siguientes parámetros:
//
//   - registroTipos: los tipos de datos para hacer unmarshall del JSON.
//     Este registro debe ser creado con [NuevoRegistroTipos] y luego se pueden
//     registrar los tipos de datos con [RegistrarTipo] o [RegistrarTipoLlave].
//
//   - conexion: La conexión a la BD para hacer CRUD de la sesión. Puede ser una
//     instancia de *pgx.Conn o *pgxpool.Pool.
//
//   - Config: La configuración de las sesiones. Ver [ConfigSesion].
//
//   - FnObtenerDatosSesion - Una funcion que debe poder recibir el uuidPersona
//     de la Sesion y conseguir todos los datos que se van a almacenar en la
//     sesion en forma de instancias de structs. Estas instancias deben ser
//     concatenadas en un []any. Si hay múltiples instancias del mismo tipo de
//     dato se pueden combinar en otro slice []any, antes de ser concatenadas.
//     Los tipos de datos de estas instancias deben coincidir con los tipos de
//     datos registrados en el registroTipos.
//     Un ejemplo del formato del slice que debe regresar la funcion es:
//     datosUsuario := []any{}
//     datosUsuario = append(datosPersona1, persona1)
//     datosUsuario = append(datosPersona1, empleado1)
//     datosUsuario = append(datosPersona1, clinica1)
//     datosUsuario = append(datosPersona1, clinica2)
type ConstructorSesion struct {
	registroTipos        RegistroTipos
	conexion             PgxQueryable
	config               *ConfigSesion
	fnObtenerDatosSesion func(uuidPersona uuid.UUID, perfil Perfil, datosExtra ...map[string]any) []any
	sesionStore          StoreSesion
}

// NuevoConstructor se utiliza para crear una instancia de ConstructorSesion.
func NuevoConstructor(conexion PgxQueryable,
	config *ConfigSesion,
	registroTipos *RegistroTipos,
	fnObtenerDatos func(uuid.UUID, Perfil, ...map[string]any) []any) *ConstructorSesion {

	cs := &ConstructorSesion{
		registroTipos:        *registroTipos,
		conexion:             conexion,
		config:               config,
		fnObtenerDatosSesion: fnObtenerDatos,
	}

	// Iniciar un store de redis
	if config.HabilitarRedis {
		redisClient := redis.NewClient(&redis.Options{
			Addr:     config.RedisAddr,
			Password: config.RedisPassword,
			DB:       config.RedisDB,
		})
		cs.sesionStore = NuevoStoreRedis(cs, redisClient)
	} else {
		cs.sesionStore = NuevaStorePostgres(cs)
	}

	return cs
}

// NuevoConstructorConRedis crea un ConstructorSesion con el cliente de Redis
// proveído.
func NuevoConstructorConRedis(conexion PgxQueryable,
	config *ConfigSesion,
	registroTipos *RegistroTipos,
	fnObtenerDatos func(uuid.UUID, Perfil, ...map[string]any) []any,
	redisClient *redis.Client) *ConstructorSesion {

	cs := &ConstructorSesion{
		registroTipos:        *registroTipos,
		conexion:             conexion,
		config:               config,
		fnObtenerDatosSesion: fnObtenerDatos,
	}

	if config.HabilitarRedis && redisClient != nil {
		cs.sesionStore = NuevoStoreRedis(cs, redisClient)
	} else {
		cs.sesionStore = NuevaStorePostgres(cs)
	}

	return cs
}

// GetStruct genera una llave con formato "packageName.T" y llama a
// [GetStructLlave].
// La llave coincide con aquella generada por [RegistrarTipo]
func GetStruct[T any](sesion *Sesion) (instancia T, _ error) {
	llave := reflect.TypeOf(instancia).String()
	return GetStructLlave[T](sesion, llave)
}

// GetStructLlave recupera una instancia de un struct registrado con una llave.
// Regresa la primer instancia registrada.
// Puede regresar [ErrorNoExiste] cuando no se encuentra la instancia
// o [ErrorTipoInvalido] cuando falló la conversión al tipo de dato T.
func GetStructLlave[T any](sesion *Sesion, llave string) (T, error) {
	structs, ok := sesion.mapaStructs[llave]
	var t T
	if !ok || len(structs) == 0 {
		return t, ErrorNoExiste{
			llaveStr: llave,
		}
	}
	t, ok = structs[0].(T)
	if !ok {
		return t, ErrorTipoInvalido
	}
	return t, nil
}

// AgregarInstancia agrega una instancia de un struct a los datos de la sesion.
// Esto incluye registrar todos los atributos y sus valores con
// [registrarCamposStructRecursivo].
// Crea una llave del mismo formato que [RegistrarTipo]:
// "packageName.structType". Verifica que el tipo de la instancia sea el
// mismo que el registrado previamente con [RegistrarTipo] o
// [RegistrarTipoLlave].
// La instancia se agrega a un slice que puede tener otras instancias del mismo
// tipo.
// Para recuperar la instancia hay que usar [GetStruct] o [GetStructLlave].
func (cs *ConstructorSesion) AgregarInstancia(sesion *Sesion, instancia any) error {
	v := reflect.ValueOf(instancia)
	tipo := v.Type()
	llave := tipo.String()
	tipoRegistrado, ok := cs.registroTipos.registro[llave]
	if !ok {
		return ErrorNoExiste{
			llaveStr: llave,
		}
	}
	if tipo != tipoRegistrado {
		return ErrorTipoInvalido
	}
	sesion.mapaStructs[llave] = append(sesion.mapaStructs[llave], instancia)
	registrarCamposStructRecursivo(sesion.atributosAnidados, sesion.atributosSinAnidar, instancia, "")
	return nil
}

// AgregarInstanciaLlave agrega una instancia de un struct a los datos de la sesion.
// Esto incluye registrar todos los atributos y sus valores con
// [registrarCamposStructRecursivo].
// Verifica que el tipo de la instancia sea el
// mismo que el registrado previamente con [RegistrarTipoLlave].
// La instancia se agrega a un slice que puede tener otras instancias del mismo
// tipo.
// Para recuperar la instancia hay que usar [GetStructLlave].
func (cs *ConstructorSesion) AgregarInstanciaLlave(sesion *Sesion, llave string, instancia any) error {
	tipo, ok := cs.registroTipos.registro[llave]
	if !ok {
		return ErrorNoExiste{
			llaveStr: llave,
		}
	}
	if reflect.TypeOf(instancia) != tipo {
		return ErrorTipoInvalido
	}
	sesion.mapaStructs[llave] = append(sesion.mapaStructs[llave], instancia)
	registrarCamposStructRecursivo(sesion.atributosAnidados, sesion.atributosSinAnidar, instancia, "")
	return nil
}

// AgregarColeccion genera una llave con formato "packageName.T"
// en donde T es el tipo de dato del primer elemento de la colección
// y llama a [AgregarColeccionLlave].
func (cs *ConstructorSesion) AgregarColeccion(sesion *Sesion, coleccion []any) error {
	if len(coleccion) == 0 {
		return nil
	}
	instancia := coleccion[0]
	v := reflect.ValueOf(instancia)
	tipo := v.Type()
	llave := tipo.String()
	return cs.AgregarColeccionLlave(sesion, llave, coleccion)
}

// AgregarColeccionLlave agrega múltiples instancias a los datos de Sesion con
// la llave dada. Las instancias se almacenan en un slice, por lo que si ya
// existen instancias para esa llave se concatenan.
//
// Aunque el código lo permite, se debe evitar agregar instancias de
// tipos diferentes a otras registradas previamente con la misma llave.
// Solo se verifica que el tipo de dato de la primera instancia sea el mismo
// que el registrado previamente con [RegistrarTipo] o [RegistrarTipoLlave].
func (cs *ConstructorSesion) AgregarColeccionLlave(sesion *Sesion, llave string, coleccion []any) error {
	if len(coleccion) == 0 {
		return nil
	}

	tipo, ok := cs.registroTipos.registro[llave]
	if !ok {
		return ErrorNoExiste{
			llaveStr: llave,
		}
	}

	if reflect.TypeOf(coleccion[0]) != tipo {
		return ErrorTipoInvalido
	}

	for _, instancia := range coleccion {
		sesion.mapaStructs[llave] = append(sesion.mapaStructs[llave], instancia)
		registrarCamposStructRecursivo(sesion.atributosAnidados, sesion.atributosSinAnidar, instancia, "")
	}

	return nil
}

// ReemplazarInstancia genera una llave de tipo "packageName.tipoInstancia" y
// si existe, reemplaza la colección almacenada ahí con una nueva colección de
// un solo elemento: [instancia].
//
// Marca error si la llave no existía previamente o si el tipo de dato
// de la instancia difiere del registro.
//
// Nota: Esto obliga a recalcular todos los atributos de la Sesion por lo que
// puede ser un proceso más pesado de lo esperado.
func (cs *ConstructorSesion) ReemplazarInstancia(sesion *Sesion, instancia any) error {
	v := reflect.ValueOf(instancia)
	tipo := v.Type()
	llave := tipo.String()
	tipoRegistrado, ok := cs.registroTipos.registro[llave]
	if !ok {
		return ErrorNoExiste{
			llaveStr: llave,
		}
	}
	if tipo != tipoRegistrado {
		return ErrorTipoInvalido
	}

	// reemplazar la colección anterior
	coll := make([]any, 1)
	coll[0] = instancia
	sesion.mapaStructs[llave] = coll

	// recalcular todos los atributos
	sesion.atributosAnidados = make(map[string][]any)
	sesion.atributosSinAnidar = make(map[string][]any)
	for _, coll := range sesion.mapaStructs {
		for _, instancia := range coll {
			registrarCamposStructRecursivo(sesion.atributosAnidados, sesion.atributosSinAnidar, instancia, "")
		}
	}

	return nil
}

// ReemplazarInstanciaLlave reemplaza la colección registrada con la llave dada
// con el nuevo slice de un solo elemento: [instancia].
//
// Marca error si la llave no existía previamente o si el tipo de dato
// de la instancia difiere del registro.
//
// Nota: Este proceso requiere de recalcular todos los atributos de la Sesion,
// por lo que puede ser más pesado de lo esperado.
func (cs *ConstructorSesion) ReemplazarInstanciaLlave(sesion *Sesion, llave string, instancia any) error {
	tipo, ok := cs.registroTipos.registro[llave]
	if !ok {
		return ErrorNoExiste{
			llaveStr: llave,
		}
	}
	if reflect.TypeOf(instancia) != tipo {
		return ErrorTipoInvalido
	}

	// reemplazar la colección anterior
	coll := make([]any, 1)
	coll[0] = instancia
	sesion.mapaStructs[llave] = coll

	// recalcular todos los atributos
	sesion.atributosAnidados = make(map[string][]any)
	sesion.atributosSinAnidar = make(map[string][]any)
	for _, coll := range sesion.mapaStructs {
		for _, instancia := range coll {
			registrarCamposStructRecursivo(sesion.atributosAnidados, sesion.atributosSinAnidar, instancia, "")
		}
	}

	return nil
}

// ReemplazarColeccion genera una llave con formato "packageName.T"
// en donde T es el tipo de dato del primer elemento de la colección
// y llama a [ReemplazarColeccionLlave].
func (cs *ConstructorSesion) ReemplazarColeccion(sesion *Sesion, coleccion []any) error {
	if len(coleccion) == 0 {
		return nil
	}
	instancia := coleccion[0]
	v := reflect.ValueOf(instancia)
	tipo := v.Type()
	llave := tipo.String()
	return cs.ReemplazarColeccionLlave(sesion, llave, coleccion)
}

// ReemplazarColeccionLlave reemplaza toda la colección de instancias
// registradas con cierta llave por la colección dada.
// Solo valida el tipo de dato del primer elemento de la colección.
//
// Nota: todos los atributos de la Sesion se tienen que recalcular por
// lo que puede que sea un procedimiento más pesado de lo que se espera.
func (cs *ConstructorSesion) ReemplazarColeccionLlave(sesion *Sesion, llave string, coleccion []any) error {
	if len(coleccion) == 0 {
		return nil
	}

	tipo, ok := cs.registroTipos.registro[llave]
	if !ok {
		return ErrorNoExiste{
			llaveStr: llave,
		}
	}

	if reflect.TypeOf(coleccion[0]) != tipo {
		return ErrorTipoInvalido
	}

	sesion.mapaStructs[llave] = coleccion

	// recalcular todos los atributos
	sesion.atributosAnidados = make(map[string][]any)
	sesion.atributosSinAnidar = make(map[string][]any)
	for _, coll := range sesion.mapaStructs {
		for _, instancia := range coll {
			registrarCamposStructRecursivo(sesion.atributosAnidados, sesion.atributosSinAnidar, instancia, "")
		}
	}

	return nil
}

// GetStructTodos genera una llave tipo string con formato "packageName.T"
// y llama a [GetStructTodosLLave].
// La llave coincide con aquella generada por [RegistrarColeccionStructs].
func GetStructTodos[T any](sesion *Sesion) (coleccion []T, err error) {
	llave := reflect.TypeOf(coleccion).Elem().String()
	return GetStructTodosLlave[T](sesion, llave)
}

// GetStructTodosLlave regresa la colección de todas las instancias registradas
// con una llave. Intenta convertir todas las instancias al tipo T. Por lo que
// va a regresar un error si las instancias de la colección tienen tipos
// diferentes.
func GetStructTodosLlave[T any](sesion *Sesion, llave string) ([]T, error) {
	tipos, ok := sesion.mapaStructs[llave]
	if !ok || len(tipos) == 0 {
		return nil, ErrorNoExiste{
			llaveStr: llave,
		}
	}
	resultado := make([]T, len(tipos))
	for i, tipo := range tipos {
		resultado[i], ok = tipo.(T)
		if !ok {
			return nil, ErrorTipoInvalido
		}
	}
	return resultado, nil
}

// Crea una nueva Sesion sin ligarla a un UuidPersona y sin obtener los datos
// con cs.FnObtenerDatosSesion para cuando no se obtenga el uuidSesion del
// cookie o la Sesion ya haya expirado.
func (cs *ConstructorSesion) NuevaSesion(r *http.Request) *Sesion {
	ahora := time.Now().Local()
	uuidSesion, err := uuid.NewV4()
	if err != nil {
		uuidSesion = uuid.Nil
	}

	sesion := &Sesion{
		constructorSesion:  cs,
		mapaStructs:        make(map[string][]any),
		atributosAnidados:  make(map[string][]any),
		atributosSinAnidar: make(map[string][]any),
		conexion:           cs.conexion,
		UuidPersona:        nil,
		UuidSesion:         uuidSesion,
		Perfil:             PerfilAnonimo,
		TipoSesion:         TipoSesionCookie, // Default to cookie session type
		FechaCreacion:      ahora,
		FechaActualizacion: ahora,
		FechaUltimoAcceso:  ahora,
		UltimaPeticion:     &ahora,
		DireccionIP:        obtenerIPCliente(r),
		AgenteUsuario:      r.UserAgent(),
		Huella:             generarHuella(r),
		InfoDispositivo:    parsearInfoDispositivo(r.UserAgent()),
		TotalPeticiones:    0, // La sesion se debe [Tocar] inmediatamente
		FechaExpiracion:    nil,
		MensajesFlash:      []MensajeFlash{},
		Recuerdame:         false,
	}
	return sesion
}

func (cs *ConstructorSesion) calcularFechaExpiracion(fechaBase time.Time) time.Time {
	// Calcular expiración basada en DuracionInactiva (no DuracionMaxima)
	// La DuracionMaxima es un límite absoluto que se verifica en Tocar()
	limite := fechaBase.Add(cs.config.DuracionInactiva)

	// Verificar que no exceda el límite máximo de duración total de sesión
	limiteMaximo := fechaBase.Add(cs.config.DuracionMaxima)
	if limite.After(limiteMaximo) {
		limite = limiteMaximo
	}

	if cs.config.SesionDuraMasDeUnDia {
		return limite
	}

	// Limitar a 23:59:59 del mismo día si SesionDuraMasDeUnDia es false
	fechaBase = fechaBase.Local()
	if limite.Local().Day() != fechaBase.Day() ||
		limite.Local().Month() != fechaBase.Month() ||
		limite.Local().Year() != fechaBase.Year() {
		diaBase := fechaBase.Format("01/02 '06 -0700")
		horaMaxStr := "23:59:59 " + diaBase
		t, err := time.Parse("15:04:05 01/02 '06 -0700", horaMaxStr)
		if err != nil {
			panic(err)
		}
		return t
	}
	return limite
}

// FirmarPersona agrega el uuidPersona, la fechaExpiracion y los datos de la
// Sesion a una Sesion. Se debería usar esta función para modificar la Sesion
// del contexto al firmar un usuario.
// Llama a la funcion
// cs.FnObtenerDatosSesion(uuidPersona, perfil, datosExtra...) para obtener los
// datos de la Sesion.
func (cs *ConstructorSesion) FirmarPersona(sesion *Sesion, uuidPersona uuid.UUID, perfil Perfil, datosExtra ...map[string]any) error {
	err := ValidarPerfil(perfil)
	if err != nil {
		return err
	}
	ahora := time.Now().Local()
	fechaExpiracion := cs.calcularFechaExpiracion(ahora)

	sesion.CerrarSesion()

	sesion.FechaCreacion = ahora
	sesion.FechaExpiracion = &fechaExpiracion
	sesion.FechaActualizacion = ahora
	sesion.UuidPersona = &uuidPersona
	sesion.Perfil = perfil

	datos := cs.fnObtenerDatosSesion(uuidPersona, perfil, datosExtra...)
	err = cs.registrarDatosVariados(sesion, datos)
	if err != nil {
		sesion.CerrarSesion()
		return err
	}

	return nil
}

func (cs *ConstructorSesion) registrarDatosVariados(sesion *Sesion, datos []any) error {
	var err error
	for _, dato := range datos {
		tipo := reflect.TypeOf(dato)
		if tipo.Kind() != reflect.Struct {
			return ErrorSeEsperabaStruct{
				instanciaStr: fmt.Sprintf("%+v", dato),
			}
		}
		err = cs.AgregarInstancia(sesion, dato)
		if err != nil {
			return err
		}
	}

	return err
}

func (cs *ConstructorSesion) debeRegenerar(sesion *Sesion) bool {
	return time.Since(sesion.FechaActualizacion) > cs.config.IntervaloRegeneracion
}

func (cs *ConstructorSesion) ConfigCodificarCookie() bool {
	return cs.config.CookieCodificar
}

// Metodos que debe usar el usuario, delegan al metodo correcto del Store
// configurado.

func (cs *ConstructorSesion) AlmacenarSesion(ctx context.Context, sesion *Sesion) error {
	return cs.sesionStore.Almacenar(ctx, sesion)
}

func (cs *ConstructorSesion) SesionDeUuid(ctx context.Context, uuidSesion uuid.UUID) (*Sesion, error) {
	return cs.sesionStore.Recuperar(ctx, uuidSesion)
}

func (cs *ConstructorSesion) SesionDeHuella(ctx context.Context, huella string) (*Sesion, error) {
	return cs.sesionStore.RecuperarPorHuella(ctx, huella)
}

func (cs *ConstructorSesion) ActualizarSesion(ctx context.Context, sesion *Sesion) error {
	return cs.sesionStore.Actualizar(ctx, sesion)
}

func (cs *ConstructorSesion) EliminarSesion(ctx context.Context, uuidSesion uuid.UUID) error {
	return cs.sesionStore.Eliminar(ctx, uuidSesion)
}

func (cs *ConstructorSesion) EliminarSesionesPersona(ctx context.Context, uuidPersona uuid.UUID) error {
	return cs.sesionStore.EliminarSesionesPersona(ctx, uuidPersona)
}

func (cs *ConstructorSesion) ActualizarOInsertarSesion(ctx context.Context, sesion *Sesion) error {
	return cs.sesionStore.AlmacenarOActualizar(ctx, sesion)
}
