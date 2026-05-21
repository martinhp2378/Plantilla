package sesiones

import (
	"errors"
	"fmt"
)

var ErrorPunteroNulo error = errors.New("el puntero es nulo")
var ErrorTipoInvalido error = errors.New("el tipo no coincide")
var ErrorSeEsperabaHuella error = errors.New("la huella no puede estar vacía")
var ErrorPerfilInvalido error = errors.New("el perfil es inválido o está vacío")
var ErrorNoSePermiteRegistrarSesion error = errors.New("no se permite registar la llave sesiones.Sesion")

type ErrorUnmarshallJson struct {
	err error
}

func (e ErrorUnmarshallJson) Error() string {
	return e.err.Error()
}

func (e ErrorUnmarshallJson) Is(target error) bool {
	_, ok := target.(ErrorUnmarshallJson)
	return ok
}

type ErrorSeEsperabaStruct struct {
	instanciaStr string
}

func (e ErrorSeEsperabaStruct) Error() string {
	return fmt.Sprintf(
		"se esperaba la instancia de un struct, se obtuvo: %s",
		e.instanciaStr)
}

func (e ErrorSeEsperabaStruct) Is(target error) bool {
	_, ok := target.(ErrorSeEsperabaStruct)
	return ok
}

type ErrorNoExiste struct {
	llaveStr string
}

func (e ErrorNoExiste) Error() string {
	return fmt.Sprintf(
		"no existe el registro o atributo con llave: %s",
		e.llaveStr)
}

func (e ErrorNoExiste) Is(target error) bool {
	_, ok := target.(ErrorNoExiste)
	return ok
}

type ErrorCampoSesionNoExiste struct {
	err error
}

func (e ErrorCampoSesionNoExiste) Error() string {
	return e.err.Error()
}

func (e ErrorCampoSesionNoExiste) Is(target error) bool {
	_, ok := target.(ErrorCampoSesionNoExiste)
	return ok
}

type ErrorNoExisteSesionEnStore struct {
	uuidSesion string
}

func (e ErrorNoExisteSesionEnStore) Error() string {
	return fmt.Sprintf(
		"no existe Sesion en el store con uuidSesion: %s",
		e.uuidSesion)
}

func (e ErrorNoExisteSesionEnStore) Is(target error) bool {
	_, ok := target.(ErrorNoExisteSesionEnStore)
	return ok
}

type ErrorPersonaNoTieneSesion struct {
	uuidPersona string
}

func (e ErrorPersonaNoTieneSesion) Error() string {
	return fmt.Sprintf(
		"no existe Sesion para la persona con uuid: %s",
		e.uuidPersona)
}

func (e ErrorPersonaNoTieneSesion) Is(target error) bool {
	_, ok := target.(ErrorPersonaNoTieneSesion)
	return ok
}

type ErrorHuellaNoTieneSesion struct {
	huella string
}

func (e ErrorHuellaNoTieneSesion) Error() string {
	return fmt.Sprintf(
		"no existe Sesion para la huella: %s",
		e.huella)
}

func (e ErrorHuellaNoTieneSesion) Is(target error) bool {
	_, ok := target.(ErrorHuellaNoTieneSesion)
	return ok
}

type ErrorTokenInvalido struct {
	err error
}

func (e ErrorTokenInvalido) Error() string {
	if e.err != nil {
		return fmt.Sprintf("token JWT inválido: %v", e.err)
	}
	return "token JWT inválido"
}

func (e ErrorTokenInvalido) Is(target error) bool {
	_, ok := target.(ErrorTokenInvalido)
	return ok
}

func (e ErrorTokenInvalido) Unwrap() error {
	return e.err
}

type ErrorTokenExpirado struct {
	err error
}

func (e ErrorTokenExpirado) Error() string {
	if e.err != nil {
		return fmt.Sprintf("token JWT expirado: %v", e.err)
	}
	return "token JWT expirado"
}

func (e ErrorTokenExpirado) Is(target error) bool {
	_, ok := target.(ErrorTokenExpirado)
	return ok
}

func (e ErrorTokenExpirado) Unwrap() error {
	return e.err
}

type ErrorTokenNoEncontrado struct {
	detalle string
}

func (e ErrorTokenNoEncontrado) Error() string {
	if e.detalle != "" {
		return fmt.Sprintf("token JWT no encontrado: %s", e.detalle)
	}
	return "token JWT no encontrado en el header Authorization"
}

func (e ErrorTokenNoEncontrado) Is(target error) bool {
	_, ok := target.(ErrorTokenNoEncontrado)
	return ok
}

type ErrorRefreshTokenInvalido struct {
	err error
}

func (e ErrorRefreshTokenInvalido) Error() string {
	if e.err != nil {
		return fmt.Sprintf("refresh token inválido: %v", e.err)
	}
	return "refresh token inválido"
}

func (e ErrorRefreshTokenInvalido) Is(target error) bool {
	_, ok := target.(ErrorRefreshTokenInvalido)
	return ok
}

func (e ErrorRefreshTokenInvalido) Unwrap() error {
	return e.err
}
