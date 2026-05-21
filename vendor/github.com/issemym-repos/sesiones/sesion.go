package sesiones

import (
	"reflect"
	"slices"
	"time"

	"github.com/gofrs/uuid/v5"
)

// Perfil es necesario a la hora de firmar un usuario. es una de las 4
// propiedades que determinan una Politica.
type Perfil string

const (
	PerfilAnonimo      Perfil = "anonimo"
	PerfilSuperAdmin   Perfil = "superadmin"
	PerfilPaciente     Perfil = "paciente"
	PerfilMedico       Perfil = "medico"
	PerfilEnfermeria   Perfil = "enfermeria"
	PerfilRecepcion    Perfil = "recepcion"
	PerfilAdminClinica Perfil = "adminclinica"
	PerfilExterno      Perfil = "externo"

	PerfilAgendaSuperAdmin Perfil = "agenda.superadmin"
	PerfilAgendaEmpleado   Perfil = "agenda.empleado"
	PerfilAgendaSustituto  Perfil = "agenda.sustituto"
)

type TipoSesion string

const (
	TipoSesionCookie TipoSesion = "cookie"
	TipoSesionJWT    TipoSesion = "jwt"
)

func ValidarPerfil(p Perfil) error {
	if p == "" || (p != PerfilAnonimo &&
		p != PerfilSuperAdmin &&
		p != PerfilMedico &&
		p != PerfilEnfermeria &&
		p != PerfilRecepcion &&
		p != PerfilAdminClinica) {
		return ErrorPerfilInvalido
	}
	return nil
}

// Sesion corresponde al objeto que almacena todos los datos de una Sesion.
// Cuando un usuario se firma debe ser construida con [NuevaSesion] y
// almacenada en BD con [ActualizarOInsertarSesionEnBd].
// Posteriormente el middleware se encargará de recuperar la Sesion de la BD y
// almacenarla en el contexto del request con la llave [LlaveContexto].
type Sesion struct {
	// para accesar métodos de ConstructorSesion dentro de la misma Sesion
	constructorSesion *ConstructorSesion

	// TipoSesion puede ser "cookie" o "jwt"
	TipoSesion TipoSesion

	// mapaStructs se usa para registrar todos los datos de la aplicación
	// que van a formar parte de una sesión. Estos datos deben de ser instancias
	// de structs. Los datos pueden ser recuperados a partir de una llave tipo
	// string. Esta llave puede ser proveida como argumento o generada
	// automaticamente.
	mapaStructs map[string][]any

	// atributosAnidados se calculan a la hora de crear la Sesion a partir del
	// mapaStructs. Las llaves son siempre automáticamente generadas y son
	// de tipo prefijo.nombreCampo, en donde prefijo es la concatenación de los
	// structs padre con punto. Ejemplo: Paciente.Persona.Domicilio.Calle1
	atributosAnidados map[string][]any
	// atributosSinAnidar es similar a atributosAnidados con la diferencia que el
	// prefijo no incluye los structs padre solo el que contiene el atributo:
	// Domicilio.Calle1
	atributosSinAnidar map[string][]any
	// el ConstructorSesion pasa la conexion a la Sesion para usarla en algunas
	// funciones de utilería. Como para conseguir las políticas para alguna
	// acción.
	conexion PgxQueryable

	UuidSesion  uuid.UUID
	UuidPersona *uuid.UUID
	// Perfil identifica el tipo de usuario. Ayuda a obtener las políticas que le
	// pertenecen de acuerdo al recurso.
	Perfil             Perfil
	FechaCreacion      time.Time
	FechaActualizacion time.Time
	FechaUltimoAcceso  time.Time
	FechaExpiracion    *time.Time

	DireccionIP   string
	AgenteUsuario string
	MensajesFlash []MensajeFlash

	Huella          string
	InfoDispositivo InfoDispositivo

	Recuerdame bool

	TotalPeticiones int
	UltimaPeticion  *time.Time

	// RefreshTokenHash es el hash del refresh token para validación
	// Solo se usa con sesiones JWT cuando EnableRefreshTokens es true
	RefreshTokenHash []byte
}

// registrarCamposStructRecursivo recorre recursivamente un struct y agrega todos
// sus campos públicos con sus valores en 2 mapas similares que relacionan
// llave -> slice de valores. si la función encuentra que los valores son
// instancias de structs, entonces se llama recursivamente con cada instancia.
//
// En el mapaAnidado, la llave se forma concatenando un prefijo y el nombre
// del campo con '.'. El prefijo es la concatenación del nombre del struct raíz
// + los nombres de los campos hasta llegar al struct hoja que contiene al
// atributo. Ejemplos: "Cita.Paciente.Persona.UuidPersona" o
// "Cita.Medico.Empleado.Persona.UuidPersona".
//
// En el mapaSinAnidar, se ignoran los structs padre de propiedades anidadas
// y solo se coloca el nombre (tipo) del struct inmediato y el nombre del
// campo. Ejemplos: "Persona.UuidPersona", "Domicilio.Calle1".
//
// Los atributos se almacenan en slices por lo que si 2 o más atributos
// coinciden en el nombre de la llave, ambos se guardan sin reemplazar.
func registrarCamposStructRecursivo(mapaAnidado map[string][]any,
	mapaSinAnidar map[string][]any,
	structARegistrar any,
	prefijoAcumulado string) {

	vStruct := reflect.ValueOf(structARegistrar)
	tStruct := vStruct.Type()
	if tStruct.Kind() != reflect.Struct {
		return
	}

	// para las llaves anidadas siempre se empieza con el nombre del struct raíz
	if prefijoAcumulado == "" {
		prefijoAcumulado = tStruct.Name()
	}

	for i := range vStruct.NumField() {
		vf := vStruct.Field(i)
		tf := tStruct.Field(i)
		if tf.IsExported() {
			val := vf.Interface()
			nomCampo := tf.Name
			// para llaves anidadas se coloca el nombre del campo en structs hijos no el tipo de su struct
			nomAnidado := prefijoAcumulado + "." + nomCampo

			if tf.Type.Kind() == reflect.Struct {
				// algunos structs los dejamos pasar tal cual
				// TODO - tal vez sería mejor checar el registro de tipos del ConstructorSesion?
				if !slices.Contains([]string{"time.Time", "time.Duration"}, tf.Type.String()) {
					registrarCamposStructRecursivo(mapaAnidado, mapaSinAnidar, val, nomAnidado)
					continue
				}
			}

			// para llaves no anidadas se pone el tipo del struct inmediato + el nombre del campo
			nomSinAnidar := tStruct.Name() + "." + nomCampo
			mapaAnidado[nomAnidado] = append(mapaAnidado[nomAnidado], val)
			mapaSinAnidar[nomSinAnidar] = append(mapaSinAnidar[nomSinAnidar], val)
		}
	}
}

// Atributo1 regresa el valor correspondiente al atributo con la llave dada.
// Regresa el primer valor registrado.
// La llave debe incluir los structs padre. Es decir busca en el mapa de llaves
// anidadas.
// Ejemplo de llave: "Paciente.Clinica.Nombre"
func (sesion *Sesion) Atributo1(llaveConPrefijo string) (any, error) {
	valores, ok := sesion.atributosAnidados[llaveConPrefijo]
	if !ok || len(valores) == 0 {
		return nil, ErrorNoExiste{
			llaveStr: llaveConPrefijo,
		}
	}
	return valores[0], nil
}

// AtributoTodos regresa el slice de valores correspondiente al atributo con la
// llave dada. Regresa [ErrorNoExiste] si el slice está vacío.
// La llave debe incluir los structs padre. Es decir busca en el mapa de llaves
// anidadas.
// Ejemplo de llave: "Paciente.Clinica.Nombre"
func (sesion *Sesion) AtributoTodos(llaveConPrefijo string) ([]any, error) {
	valores, ok := sesion.atributosAnidados[llaveConPrefijo]
	if !ok || len(valores) == 0 {
		return nil, ErrorNoExiste{
			llaveStr: llaveConPrefijo,
		}
	}
	return valores, nil
}

// AtributoSA1 regresa el valor correspondiente al atributo con la llave dada.
// La llave es sin anidar y regresa el primer valor registrado.
// Ejemplo de llave: "Clinica.Nombre"
func (sesion *Sesion) AtributoSA1(llave string) (any, error) {
	valores, ok := sesion.atributosSinAnidar[llave]
	if !ok || len(valores) == 0 {
		return nil, ErrorNoExiste{
			llaveStr: llave,
		}
	}
	return valores[0], nil
}

// AtributoSATodos regresa todos los valores correspondiente a una llave dada.
// La llave es sin anidar y los valores pueden variar en el tipo de dato.
// Ejemplo de llave: "Clinica.Nombre"
func (sesion *Sesion) AtributoSATodos(llave string) ([]any, error) {
	valores, ok := sesion.atributosSinAnidar[llave]
	if !ok || len(valores) == 0 {
		return nil, ErrorNoExiste{
			llaveStr: llave,
		}
	}
	return valores, nil
}

func (cs *ConstructorSesion) Tocar(sesion *Sesion) {
	ahora := time.Now().Local()
	sesion.FechaActualizacion = ahora
	sesion.TotalPeticiones++
	sesion.FechaUltimoAcceso = ahora
	sesion.UltimaPeticion = &ahora

	// Actualizar la expiración de acuerdo a la DuracionInactiva
	// para sesiones no anónimas.
	// Teniendo en cuenta la lógica SesionDuraMasDeUnDia
	if sesion.FechaExpiracion != nil && sesion.EstaFirmado() {
		nuevaExpiracion := ahora.Add(cs.config.DuracionInactiva)

		limiteMaximo := sesion.FechaCreacion.Add(cs.config.DuracionMaxima)
		if nuevaExpiracion.After(limiteMaximo) {
			nuevaExpiracion = limiteMaximo
		}

		if !cs.config.SesionDuraMasDeUnDia {
			nuevaExpiracion = cs.limitarAMismoDia(sesion.FechaCreacion, nuevaExpiracion)
		}

		sesion.FechaExpiracion = &nuevaExpiracion
	}
}

// limitarAMismoDia se asegura que la sesión no pase de 23:59:59
func (cs *ConstructorSesion) limitarAMismoDia(fechaCreacion time.Time,
	expiracion time.Time) time.Time {

	fechaCreacion = fechaCreacion.Local()
	expiracion = expiracion.Local()

	if expiracion.Year() != fechaCreacion.Year() ||
		expiracion.Month() != fechaCreacion.Month() ||
		expiracion.Day() != fechaCreacion.Day() {

		finDia := time.Date(
			fechaCreacion.Year(),
			fechaCreacion.Month(),
			fechaCreacion.Day(),
			23, 59, 59, 0,
			fechaCreacion.Location(),
		)
		return finDia
	}

	return expiracion
}

// Regenerar regenera el id de una Sesion para evitar ataques de fijación de
// sesión.
func (cs *ConstructorSesion) Regenerar(sesion *Sesion) {
	nuevoUuid, err := uuid.NewV4()
	if err != nil {
		nuevoUuid = uuid.Nil
	}
	sesion.UuidSesion = nuevoUuid
	sesion.FechaActualizacion = time.Now().Local()
}

// YaExpiro verifica si la Sesion pasó de la fecha de expiración.
// Depende de que la FechaExpiracion haya sido correctamente colocada.
// El método [ConstructorSesion.Tocar] se asegura de esto, sobre todo para
// cuando [SesionDuraMasDeUnDia] es `false`.
func (cs *ConstructorSesion) YaExpiro(sesion *Sesion) bool {
	if sesion.FechaExpiracion == nil {
		return false
	}
	return time.Now().After(*sesion.FechaExpiracion)
}

func (cs *ConstructorSesion) tieneLimite(sesion *Sesion) bool {
	// TODO - implementar límites de peticiones
	// if !cs.Config.HabilitarLimitePeticiones {
	// 	return false
	// }
	// tiempoDesdeUltimaPeticion := time.Since(sesion.UltimaPeticion)
	// if tiempoDesdeUltimaPeticion < time.Minute {
	// 	peticionesUltimoMinuto := sesion.TotalPeticiones
	// 	return peticionesUltimoMinuto > cs.Config.PeticionesMaximasPorMinuto
	// }

	return false
}

func (sesion *Sesion) EstaFirmado() bool {
	return sesion.UuidPersona != nil
}

// CerrarSesion elimina los datos de la Sesion y el UuidPersona que la liga a
// un usuario. También elimina la FechaExpiracion.
func (sesion *Sesion) CerrarSesion() {
	ahora := time.Now().Local()
	sesion.UuidPersona = nil
	sesion.mapaStructs = make(map[string][]any)
	sesion.atributosAnidados = make(map[string][]any)
	sesion.atributosSinAnidar = make(map[string][]any)
	sesion.FechaActualizacion = ahora
	sesion.FechaExpiracion = nil
	sesion.Perfil = PerfilAnonimo
}

// Métodos que funcionan sobre ConstructorSesion pero adecuados para usar la
// Sesion.
// Esto porque la sesión se recupera fácilmente del contexto y hay varios
// métodos que requieren del ConstructorSesion pero es incómodo incluir a este
// objeto siempre que se quiera llamar uno de estos métodos.

func (sesion *Sesion) AgregarFlash(tipo TipoFlash, titulo string, contenido string, datos ...map[string]any) {
	sesion.constructorSesion.AgregarFlash(sesion, tipo, titulo, contenido, datos...)
}
func (sesion *Sesion) Tocar() {
	sesion.constructorSesion.Tocar(sesion)
}
func (sesion *Sesion) Regenerar() {
	sesion.constructorSesion.Regenerar(sesion)
}
func (sesion *Sesion) YaExpiro() bool {
	return sesion.constructorSesion.YaExpiro(sesion)
}
func (sesion *Sesion) AgregarInstancia(instancia any) error {
	return sesion.constructorSesion.AgregarInstancia(sesion, instancia)
}
func (sesion *Sesion) AgregarInstanciaLlave(llave string, instancia any) error {
	return sesion.constructorSesion.AgregarInstanciaLlave(sesion, llave, instancia)
}
func (sesion *Sesion) AgregarColeccion(coleccion []any) error {
	return sesion.constructorSesion.AgregarColeccion(sesion, coleccion)
}
func (sesion *Sesion) AgregarColeccionLlave(llave string, coleccion []any) error {
	return sesion.constructorSesion.AgregarColeccionLlave(sesion, llave, coleccion)
}
func (sesion *Sesion) ReemplazarInstancia(instancia any) error {
	return sesion.constructorSesion.ReemplazarInstancia(sesion, instancia)
}
func (sesion *Sesion) ReemplazarInstanciaLlave(llave string, instancia any) error {
	return sesion.constructorSesion.ReemplazarInstanciaLlave(sesion, llave, instancia)
}
func (sesion *Sesion) ReemplazarColeccion(coleccion []any) error {
	return sesion.constructorSesion.ReemplazarColeccion(sesion, coleccion)
}
func (sesion *Sesion) ReemplazarColeccionLlave(llave string, coleccion []any) error {
	return sesion.constructorSesion.ReemplazarColeccionLlave(sesion, llave, coleccion)
}
func (sesion *Sesion) FirmarPersona(uuidPersona uuid.UUID, perfil Perfil, datosExtra ...map[string]any) error {
	return sesion.constructorSesion.FirmarPersona(sesion, uuidPersona, perfil, datosExtra...)
}
