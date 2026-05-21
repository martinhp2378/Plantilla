# Changelog

Lista de cambios por versión

## v0.4.1

Se corrigieron bugs que podrían resultar en sesiones cerrándose inesperadamente.

## v0.4.0

Se implementaron sesiones con JWTs para poder utilizar esta librerías en
backends para apps móviles o frontends javascript.

- Se incluye middleware y se utiliza el mismo store para mantener un
flujo de trabajo similar con sesiones de cookies.

## v0.3.6

Métodos que originalmente requerían del ConstructorSesion se agregaron
directamente a Sesion los cuales son:

- AgregarFlash
- Tocar
- Regenerar
- YaExpiro
- AgregarInstancia
- AgregarInstanciaLlave
- AgregarColeccion
- AgregarColeccionLlave
- ReemplazarInstancia
- ReemplazarInstanciaLlave
- ReemplazarColeccion
- ReemplazarColeccionLlave
- FirmarPersona

Para usar estos métodos ya no es necesario tener la instancia de
ConstructorSesion.

## v0.3.2

- Se incluyó la librería de logs.

## v0.3.1

- Se eliminan métodos para guardar sesiones anónimas que cerraban la sesión.
- Se permite ahora almacenar datos en sesiones anónimas.

## v0.3.0

- Se sube la versión de Go a 1.25
- Se elimina la implementación del token CSRF debido a que go ya la incluye.
- Ver en el README el apartado de `Protección CSRF` para saber cómo implementarla.

## v0.2.4

- Se corrigió un bug en AgregarColeccion
- Se agregaron los métodos para reemplazar instancias de structs:
  - `ReemplazarInstancia`
  - `ReemplazarInstanciaLlave`
  - `ReemplazarColeccion`
  - `ReemplazarColeccionLlave`

## v0.2.3

Se cambió el tipo de dato de `ConstructorSesion.fnObtenerDatosSesion` a
`fnObtenerDatosSesion func(uuidPersona uuid.UUID, perfil Perfil, datosExtra ...map[string]any) []any`
para permitir agregar datos arbitrarios a la función de `ConstructorSesion.FirmarPersona`
