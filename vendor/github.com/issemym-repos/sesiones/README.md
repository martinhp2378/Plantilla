# Sesiones - Manejo de sesiones para projectos web en plataforma Go

Libreria que se encarga del manejo de sesiones para aplicaciones del ISSEMYM. Requiere de una
una conexion con Postgresql y opcionalmente una conexion con Redis.

## Objetivos

- Definir un diseño de Sesiones unificado para todas las aplicaciones monolíticas Go del organismo.
- Que el usuario de la librería no tenga que definir métodos para interactuar con el 'store' de la Sesion.
- Sesiones basadas en cookies que almacenan los datos en BD o caché.
- Implementación de un sistema de permisos basado en atributos ([ABAC](https://en.wikipedia.org/wiki/Attribute-based_access_control)).
- El sistema de permisos es flexible y también permite un sistema de acceso basado en roles (perfiles) para aplicaiones
más sencillas.
- Poder almacenar datos tipados (instancias de structs) en la Sesion y recuperarlos.
- Permitir la opción de utilizar Redis para el almacenamiento de Sesiones.
- Implementar algunas medidas de seguridad:
  - Expiración de sesiones
  - Regeneración de sesiones
  - Captura de huella de un usuario
  - Límite de Peticiones y otras características (TODO)
	- Protección CSRF a través de `http.NewCrossOriginProtection` (go 1.25)
- Implementar mensajes flash.
- Incluir middleware compatible con `chi` (y `net/http`), este middleware automatiza la administración de la Sesión.

## Instalación

Para instalar se requiere acceso a la organización issemym-repos.

```bash
go get -u github.com/issemym-repos/sesiones
```

Este repositorio recibe cambios constantemente por lo que es importante correr el comando anterior seguido.

## Ejemplo de uso

### 1. Iniciar la base de datos

La base de datos debe contener por lo menos lo siguiente:

- 2 schemas: `institucion` y `sesion`
- Las siguientes tablas para `institucion`:
```sql
create table institucion.persona
(
    uuid_persona  uuid primary key         default gen_random_uuid(),
    nombre        varchar(100) not null,
    ...
);
```
- Las siguientes tablas para `sesion`:
```sql
create table sesion.cat_accion
(
    nombre  varchar(50) primary key
);

create table sesion.recurso
(
    nombre      varchar(50) primary key,
    descripcion text
);

create table sesion.perfil
(
    nombre varchar(50) primary key
);

create table sesion.politica
(
    nombre      varchar(150) primary key,
    descripcion text         not null default '',
    perfil      varchar(50)  not null,
    recurso     varchar(50)  not null,
    accion      varchar(50)  not null,
    reglas      jsonb        not null,

    constraint fk_perfil foreign key (perfil)
        references sesion.perfil (nombre),
    constraint fk_recurso foreign key (recurso)
        references sesion.recurso (nombre),
    constraint fk_accion foreign key (accion)
        references sesion.cat_accion (nombre)
);
create index idx_politica_perfil on sesion.politica (perfil);
create index idx_politica_recurso on sesion.politica (recurso);
create index idx_politica_accion on sesion.politica (accion);

create table if not exists sesion.sesion
(
    uuid_sesion           uuid primary key                  default gen_random_uuid(),
    uuid_persona          uuid,
    mapa_structs          jsonb                    not null,
    perfil                varchar(50)              not null,

    fecha_creacion        timestamp with time zone not null default now(),
    fecha_actualizacion   timestamp with time zone not null default now(),
    fecha_ultimo_acceso   timestamp with time zone not null default now(),
    fecha_expiracion      timestamp with time zone,

    direccion_ip          varchar(45),
    agente_usuario        text,
    mensajes_flash        jsonb[],

    huella                varchar(36) not null default gen_random_uuid(),
    info_dispositivo      jsonb                             default '{}'::jsonb,
    recuerdame            bool                              default false,

    -- para limitar accesos
    total_peticiones      int                               default 0,
    fecha_ultima_peticion timestamp with time zone default now(),

    razon_i               varchar(40),
    razon_m               varchar(40),
    fecha_i               timestamp with time zone          default now(),
    fecha_m               timestamp with time zone,

    constraint fk_perfil foreign key (perfil)
        references sesion.perfil (nombre),
    constraint fk_persona foreign key (uuid_persona)
        references institucion.persona (uuid_persona)
);

create index idx_sesion_persona on sesion.sesion (uuid_persona);
create index idx_sesion_huella on sesion.sesion (huella);
```

> Nota: La librería no valida passwords. El cliente debe de implementar esa funcionalidad en su sistema.

Los perfiles deben de ser los siguientes (coincidentes con el archivo `./sesion.go`)
```sql
insert into sesion.perfil
(nombre)
values
('anonimo'),
('superadmin'),
('medico'),
('enfermeria'),
('recepcion'),
('adminclinica');
```

Ejemplo de datos para `cat_accion`:

```sql
insert into sesion.cat_accion
  (nombre)
values
  ('buscar'),
  ('consultar'),
  ('modificar'),
  ('crear'),
  ('eliminar');
```

Ejemplo para sesion.recurso, los recursos deben de tener un prefijo como el schema de la tabla.

```sql
insert into sesion.recurso
(nombre)
values
('institucion.persona'),
('portal.cita_medica');
```

Ejemplo para sesion.politica (Ver el apartado de políticas)

```sql
insert into sesion.politica
(nombre, perfil, recurso, accion, reglas)
values
('superadmin_cita_medica_modificar', 'superadmin', 'portal.cita_medica', 'modificar',
'[{"operador": "=",
   "argumento1": {"llave": "Sesion.Perfil", "origen": "sujeto", "valor": ""},
   "argumento2": {"llave": "string", "origen": "escalar", "valor": "superadmin"}}]'),
('medico_cita_medica_modificar', 'medico', 'portal.cita_medica', 'modificar',
 '[{"operador": "=",
   "argumento1": {"llave": "Persona.UuidPersona", "origen": "sujeto", "valor": ""},
   "argumento2": {"llave": "Cita.Persona.UuidPersona", "origen": "recurso", "valor": ""}},
   {"operador": "after",
     "argumento1": {"llave": "Cita.Fecha", "origen": "recurso", "valor": ""},
     "argumento2": {"llave": "ahora", "origen": "escalar", "valor": ""}}]');
```

### 2. Definir los tipos de datos que se registran en la sesion

Los tipos de datos que se definan deben ser structs y deben de tener el metadato `json` para poder
ser serializados correctamente y ser almacenados/recuperados del `Store`.

`./institucion/types.go`

```go
package institucion

import (
	"github.com/issemym-repos/sesiones"
	"github.com/gofrs/uuid/v5"
)

// un tipo de dato que se va a registrar en la Sesion
type Persona struct {
	UuidPersona uuid.UUID `json:"uuid_persona"`
	Nombre      string    `json:"nombre"`
}
```

### 3. Definir una función o método que va a conseguir los datos de sesion a la hora de llamar `FirmarPersona`

Se tiene que definir una función con el siguiente tipo de dato:

`func(uuidPersona uuid.UUID, perfil Perfil, datosExtra ...map[string]any) []any`

Esta función será llamada al momento de llamar el método `ConstructorSesion.FirmarPersona`.

Los mismos argumentos que se pasen a `FirmarPersona` se pasarán a esta nueva función definida.
El propósito de esta función es de obtener todos los datos de sesión del usuario (persona) que se está firmando.

```go
package store

import (
  "context"
  "github.com/gofrs/uuid/v5"
  "github.com/jackc/pgx/v5"
  "github.com/jackc/pgx/v5/pgxpool"

  "miproyecto/institucion"
)

type ConexionPsql struct {
  *pgxpool.Pool
}

func CrearConexionPsql() ConexionPsql {
	...
}

func (conn ConexionPsql) ObtenerDatos(uuidPersona uuid.UUID, perfil sesiones.Perfil, datosExtra ...map[string]any) []any {
  datos := make([]any, 0)
  rowsPersona, _ := conn.Query(context.Background(),
    `select * from institucion.persona where uuid_persona = $1`, uuidPersona)
  persona, err := pgx.CollectOneRow(rowsPersona, pgx.RowToStructByPos[institucion.Persona])
  if err != nil {
	  panic(err)
  }
  datos = append(datos, persona)
  return datos
}
```

### 4. Crear una configuracion e instanciar el `RegistroTipos` y el `ConstructorSesion`

```go
package main

import (
  "github.com/issemym-repos/sesiones"
  "github.com/redis/go-redis/v9"
  "miproyecto/institucion"
  "miproyecto/store"
)

// un tipo de dato que permitira a los handlers accesar el ConstructorSesion
type Server struct {
	...
	ConstructorSesion *sesiones.ConstructorSesion
}

func main() {
  ...
  // crear un nuevo registro
  registroTipos := sesiones.NuevoRegistroTipos()
  // registrar todos los tipos que se pueden almacenar 
  registroTipos.RegistrarTipo(institucion.Persona{})
  
  // crear una conexion (o pool) a Psql
  conn := store.CrearConexionPsql()
  
  // crear una configuracion
  configSesion := sesiones.ConfigDefault()
  // hacer algunos ajustes a la configuracion
  configSesion.HabilitarRedis = true
  
  // usar para un constructor con o sin redis, crea un cliente de redis de forma automatica
  // (si se habilita)
  constructorSesion := sesiones.NuevoConstructor(conn, configSesion, registroTipos, conn.ObtenerDatos)
  
  // la otra opcion es tener una conexión a redis ya hecha y usarla 
  redisClient := redis.NewClient(&redis.Options{
    Addr:     config.RedisAddr,
    Password: config.RedisPassword,
    DB:       config.RedisDB,
  })
  constructorSesion := sesiones.NuevoConstructorConRedis(conn, configSesion, registroTipos, conn.ObtenerDatos, redisClient)
  
  
  // almacenamos el constructor
  server := Server{
	  ConstructorSesion: constructorSesion,
  }
  ...
}
```

### 5. Crear un mux que utilice el middleware

```go
package main

import (
	"github.com/issemym-repos/sesiones"
	"github.com/issemym-repos/logs"
    "github.com/go-chi/chi/v5"
)

func main() {
  ...
  constructorSesion := sesiones.NuevoConstructor(conn, registroTipos, conn.ObtenerDatos)
  
  logger := logs.NuevoLogger(configLogger)  // crear un logger estructurado

  router := chi.NewRouter()
  router.Use(logger.MiddlewareLogger)  // colocar el logger en el context antes de la sesion
  router.Use(constructorSesion.Middleware)
  ...
}
```

### 6. Firmar a un usuario

```go
package main

import (
  "github.com/go-chi/chi/v5"
  "github.com/gofrs/uuid/v5"
  "github.com/issemym-repos/sesiones"
  "net/http"
)

// funcion que obtiene el perfil de usuario de una cuenta
func obtenerPerfilUsuario(usuario string) sesiones.Perfil {
  return sesiones.PerfilExterno
}

// funcion que obtiene el uuidPersona de un usuario que se esta firmando
func obtenerUuidPersona(nombreUsuario string) uuid.UUID {
	...
}

// handler que firma a un usuario con usuario/pass correcto
func (server Server) postEntrarSitio(w http.ResponseWriter, r *http.Request) {
  // si el middleware actua antes que esta ruta, esta sesion siempre existe. 
  sesion := sesiones.RecuperarSesion(r.Context())

  usuario := r.PostFormValue("usuario")
  password := r.PostFormValue("password")

  perfilUsuario := obtenerPerfilUsuario(usuario)
  uuidPersona := obtenerUuidPersona(usuario)

  if esUsuarioValido(usuario, password) {
	  // este metodo modifica la sesion para firmar al usuario, obtener todos sus datos, al terminar
	  // la peticion actualiza el resultado en el Store
    server.ConstructorSesion.FirmarPersona(sesion, uuidPersona, perfilUsuario)
  }
}

func main() {
  ...
  router.Post("/login", postEntrarSitio)
  ...
}
```

### 7. Usar la sesion dentro de un handler

```go
// Obtener la sesion de un contexto
func (server Server) postEntrarSitio(w http.ResponseWriter, r *http.Request) {
  sesion := sesiones.RecuperarSesion(r.Context())
  
  // Checar perfil
  if sesion.Perfil == sesiones.PerfilMedico {
  // el usuario se firmo con tal perfil
  }
  
  // recuperar datos de una sesion
  persona, err := sesiones.GetStruct[institucion.Persona](sesion)
  if err != nil {
    // el registro no existe	  
  }
  uuidPersona := persona.UuidPersona
  nombrePersona := persona.Nombre
  
  // agregar un mensaje flash
  server.ConstructorSesion.AgregarFlash(sesion, sesiones.FlashExito, "Acción realizada correctamente")
}

// crear funciones de utileria
func ObtenerNombreSesion(sesion *sesiones.Sesion) string {
	persona, err := sesiones.GetStruct[institucion.Persona](sesion)
	if err != nil {
        return ""
    }
	return persona.Nombre
}
```

### 8. Usar la sesion dentro de un componente Templ

```go
package componentes

import "github.com/issemym-repos/sesiones"

templ MostrarMensajesFlash() {
	{{ sesion := sesiones.RecuperarSesion(ctx)
	   flashes := sesion.PopFlashes() }}
	
	for _, flash := range flashes {
	    <div>
			<h1>{ flash.Titulo }</h1>
			<p>{ flash.Contenido }</p>
        </div>	
    }
}
```

### 9. Utilizar validación basado en perfiles

```go
func handlePaginaSuperadmin(w http.ResponseWriter, r *http.Request) {
    sesion := sesiones.RecuperarSesion(r.Context())
	if sesion.Perfil != sesiones.PerfilSuperadmin {
        http.Error(w, "no tiene permisos para accesar a este recurso", http.StatusUnauthorized)
		return
    }
    ...
}
```

### 10. Utilizar el sistema ABAC

- Definir un recurso

```go
package miaplicacion

import (
  "institucion"
  "github.com/gofrs/uuid/v5"
  "time"
)

// un tipo de dato que se va a registrar en la Sesion
type Documento struct {
	// Este es el dueño del documento 
	institucion.Persona `json:"persona"`
	Contenido           string `json:"contenido"`
}
```

- Insertar el recurso, y la política en la BD.
```sql
insert into sesion.recurso
(nombre)
values
('miaplicacion.documento');

-- un superadmin puede modificar cualquier documento
insert into sesion.politica
(nombre, perfil, recurso, accion, reglas)
values
('superadmin_miaplicacion_documento_modificar', 'superadmin', 'miaplicacion.documento', 'modificar',
'[{"operador": "=",
   "argumento1": {"llave": "Sesion.Perfil", "origen": "sujeto", "valor": ""},
   "argumento2": {"llave": "string", "origen": "escalar", "valor": "superadmin"}}]');

-- un usuario de tipo 'paciente' debe ser el dueño del documento para modificar
insert into sesion.politica
  (nombre, perfil, recurso, accion, reglas)
values
  ('paciente_miaplicacion_documento_modificar', 'paciente', 'miaplicacion.documento', 'modificar',
   '[{"operador": "=",
      "argumento1": {"llave": "Sesion.UuidPersona", "origen": "sujeto", "valor": ""},
      "argumento2": {"llave": "Documento.Persona.UuidPersona", "origen": "recurso", "valor": ""}}]');
```

- Verificar el permiso en un handler

```go

func handlePaginaSuperadmin(w http.ResponseWriter, r *http.Request) {
    sesion := sesiones.RecuperarSesion(r.Context())
	permiso := sesiones.UsuarioPuedeRealizarAccionSinRecurso(sesion, r.Context(),
		"miaplicacion.documento", "modificar")
	if !permiso {
      http.Error(w, "no tiene permisos para accesar a este recurso", http.StatusUnauthorized)
      return
    }
    ...
}
```

- Definir un middleware y usarlo en rutas

```go
func MiddlewareRequierePolitica(recurso auth.NombreRecurso, accion auth.Accion) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sesion := sesiones.RecuperarSesion(r.Context())
			if sesion.UuidPersona == nil {
				http.Error(w, "No se encontró la sesión", http.StatusForbidden)
				return
			}
			if !sesiones.UsuarioPuedeRealizarAccionSinRecurso(sesion, r.Context(),
				recurso.String(), accion.String()) {

				http.Error(w, "No tiene permisos para acceder a este recurso", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

...
r.Route("/documento/{uuidDocumento}/modificar", func(r chi.Router) {
    r.Use(a.MiddlewareRequierePolitica("miaplicacion.documento", "modificar"))
	...
}
...
```

## Protección CSRF

A partir de go 1.25 Go incluye protección CSRF en su librería `net/http`. Para
utilizarla hay que envolver nuestro router de chi en un handler de la siguiente
manera:

```go
routerChi := chi.NewRouter()
// registramos rutas, middleware, etc...

antiCSRF := http.NewCrossOriginProtection()
// agregamos un dominio en el que confiamos si es necesario
antiCSRF.AddTrustedOrigin("https://example.com")

// envolvemos a nuestro router con el handler csrf
routerCSRF := antiCSRF.Handler(routerChi)

// utilizamos este handler como el mux de nuestro server
srv := http.Server{
	Addr:    ":8080",
	Handler: routerCSRF,
}
srv.ListenAndServe()
```

## Configuración

La libreria da un `ConfigDefault` que se puede usar como base y despues configurar:

```go
config := sesiones.ConfigDefault()
config.CookieNombre = "nombre_de_cookie"
config.DuracionInactiva = 60 * time.Minute
config.DuracionMaxima = 24 * time.Hour
```

## Politicas y sistema ABAC

La libreria proporciona un sistema para almacenar politicas y checar su cumplimiento. El uso de este módulo es
opcional. Se puede fácilmente implementar también un sistema basado en roles utilizando la propiedad
`Sesion.Perfil` y almacenando un objeto que represente la lista de acciones permitidas.

Una política es un conjunto de reglas que se relacionan con:

- Un perfil
- Un nombre de recurso
- Una accion

Una política determina si un usuario (Sesion) con un cierto `Perfil` puede realizar una cierta `Accion` sobre
un `recurso`.

Una política se define como el siguiente struct:

```go
type Politica struct {
  Nombre        string
  Descripcion   string
  Perfil        Perfil
  NombreRecurso string
  Accion        string
  Reglas        []Regla
}
```

El nombre de una política es la concatenación del Perfil, el nombre del recurso y la acción separados por guión bajo.
La descripción es para documentación. Perfil es una instancia de `sesiones.Perfil` que también debe existir en la BD,
en la tabla `sesion.perfil`. El nombre del recurso y la acción también deben de existir en la BD, en las tablas
`sesion.recurso` y`sesion.cat_accion` respectivamente.

### Obtener y validar políticas

Para que una política se cumpla **todas las reglas deben de cumplirse**. Para validar que un usuario tenga un permiso
se puede obtener todas las políticas respectivas con la función:

```go
func ObtenerPoliticas(ctx context.Context, sesion *Sesion, recurso string, accion string) ([]Politica, error)
```

Si algun campo es vacio como recurso, accion, entonces no se toma en cuenta para la busqueda (regresa todas las
politicas independiente de dicho campo). El perfil se obtiene de `sesion.Perfil`.

Y luego ver si se cumple alguna política con alguna de las funciones:

```go
func SatisfaceAlgunaPoliticaSinRecurso(sesion *Sesion, politicas []Politica) bool
func SatisfaceAlgunaPolitica[T any](sesion *Sesion, politicas []Politica, recurso *T) bool
```

También se proveen las siguientes funciones que combinan la búsqueda con la validación:

```go
func UsuarioPuedeRealizarAccionSinRecurso(sesion *Sesion, ctx context.Context, nombreRecurso, accion string) bool
func UsuarioPuedeRealizarAccion[T any](sesion *Sesion, ctx context.Context, nombreRecurso, accion string, recurso *T) bool
```

En algunos casos se tiene que verificar algunas políticas sin tener acceso a una `Sesion`, por ejemplo 
cuando se quieren obtener dentro de la funcion `ConstructorSesion.fnObtenerDatosSesion`, en estos casos se
puede llamar la siguiente funcion:

```go
func ObtenerPoliticasSinSesion(ctx context.Context, conn PgxQueryable, perfil Perfil, recurso string, accion string) ([]Politica, error)
```

`PgxQueryable` es una instancia de `*pgx.Conn` o `*pgxpool.Pool`.

### Reglas

Una regla se define con el siguiente struct:

```go
type Regla struct {
	Operador   Operador       `json:"operador"`
	Argumento1 DescrAtributo  `json:"argumento1"`
	Argumento2 *DescrAtributo `json:"argumento2"`
}
```

Las 3 partes de una regla son:

- Un operador (ver el tipo `Operador`)
- Un primer argumento obligatorio (instancia de un `DescrAtributo`)
- Un segundo argumento opcional (apuntador a un `DescrAtributo` que puede ser `nil`)

El objetivo de una regla es hacer una operación sobre 1 ó 2 atributos obtenidos de la `Sesion`, de un recurso. O de
algún valor escalar. Por ejemplo: en la aplicación pueden recursos llamados `Documento`; para saber si un usuario puede
modificar una instancia de este recurso, se decide que solo el 'dueño' del `Documento` lo puede modificar. Por ende,
la regla sería la siguiente:

- Operador: `=` (igual a).
- Primer argumento: el `UuidPersona` del usuario.
- Segundo argumento: el `UuidPersona` del dueño del documento.

Los diferentes operadores posibles y valores 'escalares' se pueden ver en el archivo `abac.go`.

> Nota: Cuando el operador es `=`, se intenta primero comparar como si ambos argumentos son del mismo tipo. En caso de
ser falso, se intenta nuevamente pero ahora comparando su representación en `string`.

### Atributos

Los argumentos de una regla son `atributos`, estos pueden ser recuperados de la `Sesion`, del `recurso` o ser un valor
`escalar`.

Un atributo se define por el struct:

```go
type DescrAtributo struct {
  Llave  string `json:"llave"`
  Origen Origen `json:"origen"`
  // Para cuando se quiere poner un escalar
  Valor string `json:"valor"`
}
```

- La llave describe el 'camino' que hay que tomar para llegar al atributo, separado por puntos.
- El origen puede ser `sesiones.OrigenRecurso`, `sesiones.OrigenSujeto` o `sesiones.OrigenEscalar`
  - `OrigenSujeto` significa que el atributo se obtiene de la sesion. Ya sea de alguno de los atributos públicos del
  objeto `Sesion` como `Perfil` o `UuidPersona` o de uno de los datos ingresados por la función
  `ConstructorSesion.fnObtenerDatos(uuid.UUID) []any`.
  - `OrigenRecurso` significa que el usuario debe proveer una instancia del recurso y el atributo debe ser público.
  - `OrigenEscalar` significa que se pueden utilizar algunos valores predefinidos, ver el tipo `ValorEscalar` para darse
  una idea.
- El valor se utiliza solo cuando el origen es escalar, siempre es cadena y es útil para definir cadenas o valores
numéricos fijos.

### Cómo se registran y recuperan los atributos.

Digamos que se tienen los siguientes structs:

```go
type Persona {
	UuidPersona uuid.UUID `json:"uuid_persona"`
	Nombre string `json:"nombre"`
}

type Clinica {
	UuidClinica uuid.UUID `json:"uuid_clinica"`
	Nombre string `json:"nombre"`
}

// Este tipo de dato se registra en la sesion
type Empleado {
	Persona `json:"persona"`
	Clinica `json:"clinica"`
}
```

El usuario de esta librería registra al tipo `Empleado` como dato de la sesion:

```go
...
registroTipos := sesiones.NuevoRegistroTipos()
// para registrar un tipo basta con una instancia vacía
registroTipos.RegistrarTipo(Empleado)

...
// esta funcion regresa un slice de any, pero es importante que cada elemento sea una instancia
// de un struct registrado en el paso anterior (no se permiten slices, ni mapas, ni escalares).
func obtenerDatosSesion(uuidPersona uuid.UUID) []any {
	datos = make([]any, 0)
	...
	empleado, err := store.ObtenerEmpleadoDePersona(uuidPersona)
	...
	datos = append(datos, empleado)
	return datos
}

constructorSesion := sesiones.NuevoConstructor(conn, registroTipos, obtenerDatosSesion)
...

// en algún handler se firma al usuario después de validar sus credenciales por ejemplo
sesion := sesiones.RecuperarSesion(r.Context())
constructorSesion.FirmarPersona(sesion, uuidPersona, sesiones.PerfilPaciente)
```

Al momento de firmar una persona es cuando se recuperan los datos definidos por la función registrada en el constructor
y se obtienen todos los atributos de una Sesion. Se extraen todos los campos públicos de los structs del slice `[]any`
que regresa la función de forma recursiva. En este caso como se registró un struct de tipo `Empleado` y se
registraron los atributos con las siguientes llaves:

- `Empleado.Persona.UuidPersona`
- `Empleado.Persona.Nombre`
- `Empleado.Clinica.UuidClinica`
- `Empleado.Clinica.Nombre`

También se registraron en el sujeto las siguientes llaves sin 'prefijo' que equivalen a las anteriores:

- `Persona.UuidPersona`
- `Persona.Nombre`
- `Clinica.UuidClinica`
- `Clinica.Nombre`

También se almacenan los campos de la sesión de forma automática y se pueden accesar con las siguientes llaves:

- `Sesion.UuidSesion`
- `Sesion.UuidPersona`
- `Sesion.Perfil`
- `Sesion.FechaCreacion`
- `Sesion.FechaActualizacion`
- `Sesion.FechaUltimoAcceso`
- `Sesion.FechaExpiracion`
- `Sesion.DireccionIP`
- `Sesion.AgenteUsuario`
- `Sesion.MensajesFlash`
- `Sesion.Huella`
- `Sesion.InfoDispositivo`
- `Sesion.Recuerdame`
- `Sesion.TotalPeticiones`
- `Sesion.UltimaPeticion`

En el ejemplo anterior el nombre de los campos de `Empleado` y el tipo de los structs hijo coinciden. En caso de que no
se sigue la siguiente regla:

```go
type Persona struct {
  UuidPersona uuid.UUID `json:"uuid_persona"`
  Nombre string `json:"nombre"`
}

type Email struct {
	UuidEmail uuid.UUID `json:"uuid_email"`
	Remitente Persona `json:"remitente"`
	Destinatario Persona `json:"destinatario"`
}
```

En este caso al obtener los atributos de `Email` se pueden accesar por las siguientes llaves:

- `Email.UuidEmail`
- `Email.Remitente.UuidPersona` - Se utiliza el nombre del campo
- `Email.Remitente.Nombre`
- `Email.Destinatario.UuidPersona`
- `Email.Destinatario.Nombre`

También por las siguientes llaves sin "prefijo":

- `Email.UuidEmail`
- `Persona.UuidPersona` - Se utiliza el tipo de dato (nombre del struct)
- `Persona.Nombre`

> Las últimas 2 llaves contienen múltiples valores.

### Evaluación de una regla

Para evaluar una regla se sigue el siguiente proceso:

- Se obtiene el operador. El operador puede ser 'unario' o 'binario', es decir, usa 1 ó 2 argumentos. Los operadores
representan funciones que regresan un valor booleano.
- Se obtiene el primer argumento y luego el segundo argumento si es necesario. Si el origen de un argumento es `sujeto`,
se busca el atributo por llave en la sesión. Si el origen es `recurso`, el recurso se descompone de la misma manera que
mostró en el apartado previo y se busca el atributo igual por llave. Si el origen es `escalar` se intenta recuperar
de algún valor predeterminado.
- Si hay algún error a la hora de conseguir los argumentos la regla automáticamente se evalúa como falsa.
- Los argumentos resultantes son de tipo `[]any`. Por lo que al evaluar un operador se hace un producto cartesiano. Si
algún valor del primer argumento combinado con algún valor del segundo argumento hace que el 'operador' regrese `true`,
la evaluación regresa `true`. De otra manera la regla se evalúa como falsa.

Debido a este último punto hay que evaluar si es conveniente especificar una llave más específica o usar una sin
"prefijo".

### Registro de una política.

Las políticas y reglas están diseñadas para almacenarse en formato JSON en la BD. Una política es un arreglo de reglas:

```sql
insert into sesion.politica
(nombre, perfil, recurso, accion, reglas)
values
('superadmin_cita_modificar', 'superadmin', 'portal.cita_medica', 'modificar',
'[{"operador": "=",
   "argumento1": {"llave": "Sesion.Perfil", "origen": "sujeto", "valor": ""},
   "argumento2": {"llave": "string", "origen": "escalar", "valor": "superadmin"}}]');
```

> Nota: Es importante respetar la primer letra mayúsucula en las llaves que accesan atributos de 'sujeto' o 'recurso'

## Integracion con Redis

De manera opcional se puede usar Redis para almacenar y recuperar los datos que originalmente se almacenan
en la tabla `sesion.sesion`:

```go
config := sesiones.ConfigDefault()
config.HabilitarRedis = true
config.RedisAddr = "localhost:6379"
config.RedisPassword = "password"
config.RedisDB = 0
```

## Algunas consideraciones de seguridad

- Habilitar Cookie seguro en sitios HTTPS (`config.CookieEsSecure: true`)
- Habilitar la protección CSRF (ver apartado con título 'Protección CSRF')

## JWTs

Para generar sesiones con JWTs en lugar de cookies, se sigue un flujo de trabajo
similar:

- Se crea una configuración incluyendo si se va a utilizar el backend de
postgresql o redis.
- Se crea un ConstructorSesionJWT y se siguen los mismos pasos para registrar
structs.
- Se coloca el middleware específico para JWTs en las rutas que van a
autenticarse con JWTs.

> Nota: En caso de JWTs no se crean sesiones anónimas o "vacías". Por lo que
el middleware no debe de colocarse en rutas de login o refresh (aquellas que no
requieran de autenticación).

### Setup de la base de datos

Si se utiliza el backend postgres, la tabla requiere de dos columnas adicionales
a las que se ocupan en las sesiones con cookies:

```sql
tipo_sesion varchar(20) not null default 'cookie', CHECK (tipo_sesion IN ('cookie', 'jwt')),
refresh_token_hash bytea,
```

### Configuración de JWTs

Se tienen los siguientes parámetros que afectan la implementación de JWTs:

- `SecretKey`: Una llave secreta única para el app. Se recomienda pasarla a la
aplicación por una variable de ambiente.
- `Algorithm`: Algoritmo HMAC: `HS256`, `HS384`, `HS512`. Default: `HS256`
- `AccessTokenExpiry`: Duración del token de acceso (1 min - 24 hrs). Default:
15 minutos
- `RefreshTokenExpiry`: Duración del refresh token. Default: 7 días
- `EnableRefreshTokens`: Usar/Habilitar refresh tokens. Default: `true`
- `Issuer`: Se recomienda colocar el nombre del app (de backend) aquí.
- `Audience`: Se recomienda colocar en un arreglo el nombre o nombres de los
frontends que pueden consumir esta app: `["app-movil"]`
- `AllowedClockSkew`: Tolerancia para validar tokens de acceso. Default: 30
segundos.

### Ejemplo básico de uso

```go
import (
    "time"
    "github.com/issemym-repos/sesiones"
)

func main() {
	
	...
	// Usar la configuración por defecto
	configJWT := sesiones.ConfigJWTDefault()
	configJWT.Issuer = "mi-aplicacion"
	configJWT.Audience = []string{"api.miapp.com"}

	registro := sesiones.NuevoRegistroTipos()
	registro.RegistrarTipo(Usuario{})
	registro.RegistrarTipo(Empresa{})

	fnObtenerDatos := func(uuidPersona uuid.UUID, perfil sesiones.Perfil, extra ...map[string]any) []any {
		// Consultar datos del usuario de la BD
		usuario := obtenerUsuario(uuidPersona)
		empresa := obtenerEmpresa(usuario.UuidEmpresa)
		return []any{usuario, empresa}
	}

	constructorJWT, err := sesiones.NuevoConstructorJWT(
		conn,
		config,
		configJWT,
		registro,
		fnObtenerDatos,
	)

	if err != nil {
		log.Fatal(err)
	}
	
	r := setupRouter(constructorJWT)
	...
}

func setupRouter(constructorJWT *sesiones.ConstructorSesionJWT) *chi.Mux {
	r := chi.NewRouter()

	// rutas con autenticación - con middleware
	r.Group(func(r chi.Router) {
		r.Use(constructorJWT.MiddlewareJWT)

		r.Get("/api/v1/perfil", handlePerfil)
		r.Get("/api/v1/datos", handleDatos)
		r.Post("/api/v1/recurso", handleCrearRecurso)
	})

	// rutas sin autenticación - sin middleware
	r.Post("/api/v1/login", handleLogin)
	r.Post("/api/v1/refresh", handleRefresh)

	return r
}
```

### Usar sesiones JWT dentro de handlers.

Siempre que el middleware haya sido registrado para la ruta donde se usa el
handler, se puede utilizar el mismo flujo de trabajo que con sesiones de
cookies:

```go
func (h *MiHandler) handlePerfil(w http.ResponseWriter, r *http.Request) {
	// Recuperar sesión del contexto
	sesion := sesiones.RecuperarSesion(r.Context())

	// Acceder a información de la sesión
	if !sesion.EstaFirmado() {
		http.Error(w, "No autenticado", http.StatusUnauthorized)
		return
	}

	// Obtener algún struct registrado
	usuario, err := sesiones.GetStruct[Usuario](sesion)
	if err != nil {
		http.Error(w, "Usuario no encontrado", http.StatusInternalServerError)
		return
	}

	// Usar datos del usuario
	json.NewEncoder(w).Encode(map[string]any{
		"uuid":   usuario.UuidPersona,
		"nombre": usuario.Nombre,
		"email":  usuario.Email,
		"perfil": sesion.Perfil,
	})
}
```

### Firmar un usuario y regresar tokens al cliente

Se da un ejemplo de cómo se puede firmar un usuario y un ejemplo de formato del
JSON a regresar (esto debe acordarse con el desarrollador de frontend):

```go
func handleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Validar credenciales (username/password, OAuth, etc.)
	usuario := validarCredenciales(r)
	if usuario == nil {
		http.Error(w, "Credenciales inválidas", http.StatusUnauthorized)
		return
	}

	// Crear nueva sesión (a diferencia de una sesión con cookies no existe una
	// sesión anónima en este momento)
	sesion := constructorJWT.NuevaSesion(r)

	// Firmar sesión y generar tokens
	accessToken, refreshToken, err := constructorJWT.FirmarPersonaJWT(
		ctx,
		sesion,
		usuario.UuidPersona,
		sesiones.PerfilUsuario, // el perfil correspondiente
	)
	if err != nil {
		http.Error(w, "Error al crear sesión", http.StatusInternalServerError)
		return
	}

	// Enviar tokens al cliente
	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer", // Esto es constante
		"expires_in":    "900", // Asegurarse que coincida con la configuración
	})
}
```

## Características de la implementación JWT

En estos siguientes apartados se muestran algunas especificaciones que pueden
ser importantes para los desarrolladores de aplicaciones frontend o móviles que
quieran autenticarse con backends que utilizan esta librería:

### Formato del request

Para rutas protegidas el cliente debe de generar un request como el siguiente:

```http
GET /api/perfil HTTP/1.1
Host: api.miapp.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

En donde la cadena después de Bearer debe ser el token de acceso.

### Errores posibles

El middleware regresa los siguientes errores si no logra recuperar una sesión
válida y vigente:

- "Token no encontrado"
- "Token inválido"
- "Tipo de token inválido"
- "UUID de sesión inválido"
- "Sesión no encontrada"
- "Tipo de sesión inválido"
- "Sesión expirada"

Estos se regresan con un header `401 - Unauthorized` y dentro de un objeto
JSON:

```json
{
	"error": "<mensaje de error>",
	"detalle": "<detalles del error>",
}
```

### Refresh Tokens

Si se habilitan en la configuración. Estos permiten generar nuevos access tokens
sin requerir que el usuario vuelva a hacer login.

Un ejemplo de implementación del handler que refresca es el siguiente:

```go
func handleRefresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Leer refresh token del cuerpo del request
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Request inválido", http.StatusBadRequest)
		return
	}

	// Generar nuevo access token
	nuevoAccessToken, err := constructorJWT.RefrescarTokenJWT(ctx, req.RefreshToken)
	if err != nil { // Puede ser ErrorRefreshTokenInvalido o ErrorTokenExpirado
		http.Error(w, "Refresh token inválido", http.StatusUnauthorized)
		return
	}

	// Enviar nuevo access token
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": nuevoAccessToken,
		"token_type":   "Bearer",
		"expires_in":   "900",
	})
}
```

Un ejemplo del request del lado del cliente es:

```http
POST /api/refresh HTTP/1.1
Host: api.miapp.com
Content-Type: application/json

{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```


