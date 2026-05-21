package sesiones

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
)

type Operador string

const (
	OperadorIgual      Operador = "="
	OperadorDiferente  Operador = "!="
	OperadorMayor      Operador = ">"
	OperadorMenor      Operador = "<"
	OperadorMayorIgual Operador = ">="
	OperadorMenorIgual Operador = "<="
	OperadorAfter      Operador = "after"
	OperadorBefore     Operador = "before"
	OperadorExiste     Operador = "existe"
)

type Origen string

const (
	OrigenSujeto  Origen = "sujeto"
	OrigenRecurso Origen = "recurso"
	OrigenEscalar Origen = "escalar"
)

type ValorEscalar string

const (
	ValorEscalarAhora     ValorEscalar = "ahora"
	ValorEscalarEntero    ValorEscalar = "int"
	ValorEscalarFlotante  ValorEscalar = "float"
	ValorEscalarCadena    ValorEscalar = "string"
	ValorEscalarBooleano  ValorEscalar = "bool"
	ValorEscalarFecha     ValorEscalar = "date"
	ValorEscalarFechaHora ValorEscalar = "datetime"
	ValorEscalarUuid      ValorEscalar = "uuid"
)

type DescrAtributo struct {
	Llave  string `json:"llave"`
	Origen Origen `json:"origen"`
	// Para cuando se quiere poner un escalar
	Valor string `json:"valor"`
}

type Regla struct {
	Operador   Operador       `json:"operador"`
	Argumento1 DescrAtributo  `json:"argumento1"`
	Argumento2 *DescrAtributo `json:"argumento2"`
}

// Politica es una serie de reglas que se deben cumplir para que un usuario
// pueda realizar una acción sobre un recurso.
// Una Politica tiene 5 partes importantes:
// - Nombre: Concatenación del Perfil, NombreRecurso, Accion y posiblemente un
// número. El nombre es llave primaria y por ende es único
// - Perfil - El Perfil de la Sesion, debe ser instancia de sesiones.Perfil.
// - NombreRecurso - Una cadena que concatena el módulo y el recurso, similar a
// incluir el schema y el nombre de la tabla. Ej: "portal.cita_medica". Debe
// estar incluido en la tabla sesion.recurso.
// - Accion - Una cadena que representa una acción. Ej: "buscar" o "crear".
// Debe estar incluido en la tabla sesion.cat_accion.
type Politica struct {
	Nombre        string  `json:"nombre"`
	Descripcion   string  `json:"descripcion"`
	Perfil        Perfil  `json:"perfil"`
	NombreRecurso string  `json:"recurso"`
	Accion        string  `json:"accion"`
	Reglas        []Regla `json:"reglas"`
}

type attrsRecurso struct {
	mapaAtributosAnidado   map[string][]any
	mapaAtributosSinAnidar map[string][]any
}

func SatisfaceRegla[T any](sesion *Sesion, regla Regla, recurso *T) bool {
	attrs := attrsRecurso{
		mapaAtributosAnidado:   make(map[string][]any),
		mapaAtributosSinAnidar: make(map[string][]any),
	}
	if recurso == nil {
		return sesion.satisfaceReglaAttrs(regla, attrs)
	}

	// checar si es necesario obtener los atributos del recurso
	// los atributos se obtienen por medio de reflexión
	if regla.Argumento1.Origen == OrigenRecurso ||
		(regla.Argumento2 != nil && regla.Argumento2.Origen == OrigenRecurso) {

		registrarCamposStructRecursivo(attrs.mapaAtributosAnidado,
			attrs.mapaAtributosSinAnidar, *recurso, "")
	}

	return sesion.satisfaceReglaAttrs(regla, attrs)
}

func SatisfaceReglaSinRecurso(sesion *Sesion, regla Regla) bool {
	attrs := attrsRecurso{
		mapaAtributosAnidado:   make(map[string][]any),
		mapaAtributosSinAnidar: make(map[string][]any),
	}
	return sesion.satisfaceReglaAttrs(regla, attrs)
}

func SatisfacePolitica[T any](sesion *Sesion, politica Politica, recurso *T) bool {
	attrs := attrsRecurso{
		mapaAtributosAnidado:   make(map[string][]any),
		mapaAtributosSinAnidar: make(map[string][]any),
	}
	// checar si es necesario obtener los atributos del recurso
	for _, regla := range politica.Reglas {
		if regla.Argumento1.Origen == OrigenRecurso ||
			(regla.Argumento2 != nil && regla.Argumento2.Origen == OrigenRecurso) {

			registrarCamposStructRecursivo(attrs.mapaAtributosAnidado,
				attrs.mapaAtributosSinAnidar, *recurso, "")
			break
		}
	}

	return sesion.satisfacePoliticaAttrs(politica, attrs)
}

func SatisfacePoliticaSinRecurso(sesion *Sesion, politica Politica) bool {
	attrs := attrsRecurso{
		mapaAtributosAnidado:   make(map[string][]any),
		mapaAtributosSinAnidar: make(map[string][]any),
	}
	return sesion.satisfacePoliticaAttrs(politica, attrs)
}

func SatisfaceAlgunaPolitica[T any](sesion *Sesion, politicas []Politica, recurso *T) bool {
	attrs := attrsRecurso{
		mapaAtributosAnidado:   make(map[string][]any),
		mapaAtributosSinAnidar: make(map[string][]any),
	}
	// checar si es necesario obtener los atributos del recurso
out:
	for _, politica := range politicas {
		for _, regla := range politica.Reglas {
			if regla.Argumento1.Origen == OrigenRecurso ||
				(regla.Argumento2 != nil && regla.Argumento2.Origen == OrigenRecurso) {

				registrarCamposStructRecursivo(attrs.mapaAtributosAnidado,
					attrs.mapaAtributosSinAnidar, *recurso, "")
				break out
			}
		}
	}

	for _, politica := range politicas {
		if sesion.satisfacePoliticaAttrs(politica, attrs) {
			return true
		}
	}
	return false
}

func SatisfaceAlgunaPoliticaSinRecurso(sesion *Sesion, politicas []Politica) bool {
	attrs := attrsRecurso{
		mapaAtributosAnidado:   make(map[string][]any),
		mapaAtributosSinAnidar: make(map[string][]any),
	}
	for _, politica := range politicas {
		if sesion.satisfacePoliticaAttrs(politica, attrs) {
			return true
		}
	}
	return false
}

// UsuarioPuedeRealizarAccionSinRecurso obtiene las Politicas que corresponden
// a un recurso, accion y sesion.Perfil y regresa si la Sesion satisface alguna
// de acuerdo a [SatisfaceAlgunaPoliticaSinRecurso].
func UsuarioPuedeRealizarAccionSinRecurso(sesion *Sesion, ctx context.Context,
	nombreRecurso, accion string) bool {

	politicas, err := ObtenerPoliticas(ctx, sesion, nombreRecurso, accion)
	if err != nil {
		return false
	}
	return SatisfaceAlgunaPoliticaSinRecurso(sesion, politicas)
}

// UsuarioPuedeRealizarAccion obtiene las Politicas que corresponden
// a un recurso, accion y sesion.Perfil y regresa si la Sesion satisface alguna
// de acuerdo a [SatisfaceAlgunaPolitica].
func UsuarioPuedeRealizarAccion[T any](sesion *Sesion, ctx context.Context,
	nombreRecurso, accion string, recurso *T) bool {

	politicas, err := ObtenerPoliticas(ctx, sesion, nombreRecurso, accion)
	if err != nil {
		return false
	}
	return SatisfaceAlgunaPolitica(sesion, politicas, recurso)
}

// Una politica solo se satisface si todas sus reglas se satisfacen.
func (sesion *Sesion) satisfacePoliticaAttrs(politica Politica, attrs attrsRecurso) bool {
	for _, regla := range politica.Reglas {
		if !sesion.satisfaceReglaAttrs(regla, attrs) {
			return false
		}
	}
	return true
}

func (sesion *Sesion) satisfaceReglaAttrs(regla Regla, attrs attrsRecurso) bool {
	switch regla.Operador {
	case OperadorIgual:
		return sesion.CumpleReglaIgual(regla, attrs)
	case OperadorDiferente:
		return sesion.CumpleReglaDiferente(regla, attrs)
	case OperadorMayor:
		return sesion.CumpleReglaMayor(regla.Argumento1, regla.Argumento2, attrs)
	case OperadorMenor:
		return sesion.CumpleReglaMenor(regla.Argumento1, regla.Argumento2, attrs)
	case OperadorMayorIgual:
		return sesion.CumpleReglaMayorIgual(regla.Argumento1, regla.Argumento2, attrs)
	case OperadorMenorIgual:
		return sesion.CumpleReglaMenorIgual(regla.Argumento1, regla.Argumento2, attrs)
	case OperadorAfter:
		return sesion.CumpleReglaAfter(regla.Argumento1, regla.Argumento2, attrs)
	case OperadorBefore:
		return sesion.CumpleReglaBefore(regla.Argumento1, regla.Argumento2, attrs)
	case OperadorExiste:
		return sesion.CumpleReglaExiste(regla.Argumento1, attrs)
	}
	return false
}

func atributoEscalar(llave string, valorParseable string) ([]any, error) {
	resultado := make([]any, 0)
	switch ValorEscalar(llave) {
	case ValorEscalarAhora:
		resultado = append(resultado, time.Now().Local())
		return resultado, nil
	case ValorEscalarEntero:
		val, err := strconv.Atoi(valorParseable)
		if err != nil {
			return nil, err
		}
		resultado = append(resultado, val)
		return resultado, nil

	case ValorEscalarFlotante:
		val, err := strconv.ParseFloat(valorParseable, 64)
		if err != nil {
			return nil, err
		}
		resultado = append(resultado, val)
		return resultado, nil

	case ValorEscalarCadena:
		resultado = append(resultado, valorParseable)
		return resultado, nil

	case ValorEscalarBooleano:
		val, err := strconv.ParseBool(valorParseable)
		if err != nil {
			return nil, err
		}
		resultado = append(resultado, val)
		return resultado, nil

	case ValorEscalarFecha:
		val, err := time.Parse("2006-01-02", valorParseable)
		if err != nil {
			return nil, err
		}
		resultado = append(resultado, val)
		return resultado, nil

	case ValorEscalarFechaHora:
		val, err := time.Parse("2006-01-02 15:04:05-07:00", valorParseable)
		if err != nil {
			return nil, err
		}
		resultado = append(resultado, val)
		return resultado, nil
	case ValorEscalarUuid:
		val, err := uuid.FromString(valorParseable)
		if err != nil {
			return nil, err
		}
		resultado = append(resultado, val)
		return resultado, nil
	}

	return nil, ErrorNoExiste{
		llaveStr: llave,
	}
}

func (sesion *Sesion) obtenerAtributoSesion(llave string) ([]any, error) {
	partes := strings.Split(llave, ".")
	campo := partes[len(partes)-1]
	datos := make([]any, 0)
	switch campo {
	case "UuidSesion":
		datos = append(datos, sesion.UuidSesion)
		return datos, nil
	case "UuidPersona":
		datos = append(datos, sesion.UuidPersona)
		return datos, nil
	case "Perfil":
		datos = append(datos, string(sesion.Perfil))
		return datos, nil
	case "FechaCreacion":
		datos = append(datos, sesion.FechaCreacion)
		return datos, nil
	case "FechaActualizacion":
		datos = append(datos, sesion.FechaActualizacion)
		return datos, nil
	case "FechaUltimoAcceso":
		datos = append(datos, sesion.FechaUltimoAcceso)
		return datos, nil
	case "FechaExpiracion":
		datos = append(datos, sesion.FechaExpiracion)
		return datos, nil
	case "DireccionIP":
		datos = append(datos, sesion.DireccionIP)
		return datos, nil
	case "AgenteUsuario":
		datos = append(datos, sesion.AgenteUsuario)
		return datos, nil
	case "MensajesFlash":
		datos = append(datos, sesion.MensajesFlash)
		return datos, nil
	case "Huella":
		datos = append(datos, sesion.Huella)
		return datos, nil
	case "InfoDispositivo":
		datos = append(datos, sesion.InfoDispositivo)
		return datos, nil
	case "Recuerdame":
		datos = append(datos, sesion.Recuerdame)
		return datos, nil
	case "TotalPeticiones":
		datos = append(datos, sesion.TotalPeticiones)
		return datos, nil
	case "UltimaPeticion":
		datos = append(datos, sesion.UltimaPeticion)
		return datos, nil
	default:
		return nil, ErrorCampoSesionNoExiste{
			err: fmt.Errorf("la llave %s, no es apunta a un campo válido de sesiones.Sesion", llave),
		}
	}
}

func (sesion *Sesion) obtenerAtributos(arg DescrAtributo, recurso attrsRecurso) ([]any, error) {
	// caso especial para parámetros de Sesion:
	if arg.Origen == OrigenSujeto && (strings.HasPrefix(arg.Llave, "sesiones.Sesion.") ||
		strings.HasPrefix(arg.Llave, "Sesion.")) {

		return sesion.obtenerAtributoSesion(arg.Llave)
	}
	// Los puntos separan structs anidados y el campo.
	// Si tiene un solo punto se debe encontrar en el mapaSinAnidar.
	buscarEnAnidados := false
	totalPuntos := 0
	for _, r := range arg.Llave {
		if r == '.' {
			totalPuntos += 1
			if totalPuntos >= 2 {
				buscarEnAnidados = true
				break
			}
		}
	}

	switch arg.Origen {
	case OrigenSujeto:
		if buscarEnAnidados {
			return sesion.AtributoTodos(arg.Llave)
		} else {
			return sesion.AtributoSATodos(arg.Llave)
		}
	case OrigenRecurso:
		var mapa map[string][]any
		if buscarEnAnidados {
			mapa = recurso.mapaAtributosAnidado
		} else {
			mapa = recurso.mapaAtributosSinAnidar
		}
		attrs, ok := mapa[arg.Llave]
		if !ok || len(attrs) == 0 {
			return nil, ErrorNoExiste{
				llaveStr: arg.Llave,
			}
		}
		return attrs, nil
	case OrigenEscalar:
		return atributoEscalar(arg.Llave, arg.Valor)
	}
	return nil, ErrorNoExiste{
		llaveStr: arg.Llave,
	}
}

func (sesion *Sesion) CumpleReglaIgual(regla Regla, recurso attrsRecurso) bool {
	if regla.Argumento2 == nil {
		return false
	}
	arg1, err := sesion.obtenerAtributos(regla.Argumento1, recurso)
	if err != nil || len(arg1) == 0 {
		return false
	}
	arg2, err := sesion.obtenerAtributos(*regla.Argumento2, recurso)
	if err != nil || len(arg2) == 0 {
		return false
	}
	for _, a1 := range arg1 {
		if slices.Contains(arg2, a1) {
			return true
		}
	}

	// Intentar nuevamente pero ahora comparando su representación como cadena
	arg2Cadenas := make([]string, len(arg2))
	for i, v := range arg2 {
		arg2Cadenas[i] = fmt.Sprintf("%v", v)
	}
	for _, a1 := range arg1 {
		s1 := fmt.Sprintf("%v", a1)
		if slices.Contains(arg2Cadenas, s1) {
			return true
		}
	}
	return false
}

func (sesion *Sesion) CumpleReglaDiferente(regla Regla, recurso attrsRecurso) bool {
	if regla.Argumento2 == nil {
		return false
	}
	arg1, err := sesion.obtenerAtributos(regla.Argumento1, recurso)
	if err != nil || len(arg1) == 0 {
		return false
	}
	arg2, err := sesion.obtenerAtributos(*regla.Argumento2, recurso)
	if err != nil || len(arg2) == 0 {
		return false
	}
	for _, a1 := range arg1 {
		for _, a2 := range arg2 {
			if a1 != a2 {
				return true
			}
		}
	}
	return false
}

// se usa para >, <, >=, <=, en donde Go obliga a convertir a algún tipo como
// int, float64, string.
func comparacionCartesiana(arg1 []any, arg2 []any, op Operador) bool {
	if !slices.Contains([]Operador{
		OperadorMayor,
		OperadorMenor,
		OperadorMayorIgual,
		OperadorMenorIgual},
		op) {
		return false
	}
	ints1 := make([]int, len(arg1))
	ints2 := make([]int, len(arg2))
	float64s1 := make([]float64, len(arg1))
	float64s2 := make([]float64, len(arg2))
	strings1 := make([]string, len(arg1))
	strings2 := make([]string, len(arg2))

	// TODO - debe haber alguna manera mejor de hacer esto?
	switch arg1[0].(type) {
	case int:
		for i, a1 := range arg1 {
			ints1[i] = a1.(int)
		}
		for i, a2 := range arg2 {
			int2, ok := a2.(int)
			if !ok {
				return false
			}
			ints2[i] = int2
		}

		for _, a1 := range ints1 {
			for _, a2 := range ints2 {
				switch op {
				case OperadorMayor:
					if a1 > a2 {
						return true
					}
					break
				case OperadorMenor:
					if a1 < a2 {
						return true
					}
					break
				case OperadorMayorIgual:
					if a1 >= a2 {
						return true
					}
					break
				case OperadorMenorIgual:
					if a1 <= a2 {
						return true
					}
					break
				default: // no debe llegar aquí
					return false
				}
			}
		}
	case float64:
		for i, a1 := range arg1 {
			float64s1[i] = a1.(float64)
		}
		for i, a2 := range arg2 {
			float642, ok := a2.(float64)
			if !ok {
				return false
			}
			float64s2[i] = float642
		}

		for _, a1 := range float64s1 {
			for _, a2 := range float64s2 {
				switch op {
				case OperadorMayor:
					if a1 > a2 {
						return true
					}
					break
				case OperadorMenor:
					if a1 < a2 {
						return true
					}
					break
				case OperadorMayorIgual:
					if a1 >= a2 {
						return true
					}
					break
				case OperadorMenorIgual:
					if a1 <= a2 {
						return true
					}
					break
				default: // no debe llegar aquí
					return false
				}
			}
		}
	case string:
		for i, a1 := range arg1 {
			strings1[i] = a1.(string)
		}
		for i, a2 := range arg2 {
			string2, ok := a2.(string)
			if !ok {
				return false
			}
			strings2[i] = string2
		}

		for _, a1 := range strings1 {
			for _, a2 := range strings2 {
				switch op {
				case OperadorMayor:
					if a1 > a2 {
						return true
					}
					break
				case OperadorMenor:
					if a1 < a2 {
						return true
					}
					break
				case OperadorMayorIgual:
					if a1 >= a2 {
						return true
					}
					break
				case OperadorMenorIgual:
					if a1 <= a2 {
						return true
					}
					break
				default: // no debe llegar aquí
					return false
				}
			}
		}
	}

	return false
}

func (sesion *Sesion) CumpleReglaMayor(desc1 DescrAtributo, desc2 *DescrAtributo, recurso attrsRecurso) bool {
	if desc2 == nil {
		// se necesitan 2 argumentos para este operador
		return false
	}
	arg1, err := sesion.obtenerAtributos(desc1, recurso)
	if err != nil || len(arg1) == 0 {
		return false
	}
	arg2, err := sesion.obtenerAtributos(*desc2, recurso)
	if err != nil || len(arg2) == 0 {
		return false
	}
	return comparacionCartesiana(arg1, arg2, OperadorMayor)
}

func (sesion *Sesion) CumpleReglaMenor(desc1 DescrAtributo, desc2 *DescrAtributo, recurso attrsRecurso) bool {
	if desc2 == nil {
		// se necesitan 2 argumentos para este operador
		return false
	}
	arg1, err := sesion.obtenerAtributos(desc1, recurso)
	if err != nil || len(arg1) == 0 {
		return false
	}
	arg2, err := sesion.obtenerAtributos(*desc2, recurso)
	if err != nil || len(arg2) == 0 {
		return false
	}
	return comparacionCartesiana(arg1, arg2, OperadorMenor)
}

func (sesion *Sesion) CumpleReglaMayorIgual(desc1 DescrAtributo, desc2 *DescrAtributo, recurso attrsRecurso) bool {
	if desc2 == nil {
		// se necesitan 2 argumentos para este operador
		return false
	}
	arg1, err := sesion.obtenerAtributos(desc1, recurso)
	if err != nil || len(arg1) == 0 {
		return false
	}
	arg2, err := sesion.obtenerAtributos(*desc2, recurso)
	if err != nil || len(arg2) == 0 {
		return false
	}
	return comparacionCartesiana(arg1, arg2, OperadorMayorIgual)
}

func (sesion *Sesion) CumpleReglaMenorIgual(desc1 DescrAtributo, desc2 *DescrAtributo, recurso attrsRecurso) bool {
	if desc2 == nil {
		// se necesitan 2 argumentos para este operador
		return false
	}
	arg1, err := sesion.obtenerAtributos(desc1, recurso)
	if err != nil || len(arg1) == 0 {
		return false
	}
	arg2, err := sesion.obtenerAtributos(*desc2, recurso)
	if err != nil || len(arg2) == 0 {
		return false
	}
	return comparacionCartesiana(arg1, arg2, OperadorMenorIgual)
}

func (sesion *Sesion) CumpleReglaAfter(desc1 DescrAtributo, desc2 *DescrAtributo, recurso attrsRecurso) bool {
	if desc2 == nil {
		// se necesitan 2 argumentos para este operador
		return false
	}
	arg1, err := sesion.obtenerAtributos(desc1, recurso)
	if err != nil || len(arg1) == 0 {
		return false
	}
	arg2, err := sesion.obtenerAtributos(*desc2, recurso)
	if err != nil || len(arg2) == 0 {
		return false
	}
	times1 := make([]time.Time, len(arg1))
	times2 := make([]time.Time, len(arg2))
	for i, a1 := range arg1 {
		t, ok := a1.(time.Time)
		if !ok {
			return false
		}
		times1[i] = t
	}
	for i, a2 := range arg2 {
		t, ok := a2.(time.Time)
		if !ok {
			return false
		}
		times2[i] = t
	}
	for _, a1 := range times1 {
		if slices.ContainsFunc(times2, a1.After) {
			return true
		}
	}
	return false
}

func (sesion *Sesion) CumpleReglaBefore(desc1 DescrAtributo, desc2 *DescrAtributo, recurso attrsRecurso) bool {
	if desc2 == nil {
		// se necesitan 2 argumentos para este operador
		return false
	}
	arg1, err := sesion.obtenerAtributos(desc1, recurso)
	if err != nil || len(arg1) == 0 {
		return false
	}
	arg2, err := sesion.obtenerAtributos(*desc2, recurso)
	if err != nil || len(arg2) == 0 {
		return false
	}
	times1 := make([]time.Time, len(arg1))
	times2 := make([]time.Time, len(arg2))
	for i, a1 := range arg1 {
		t, ok := a1.(time.Time)
		if !ok {
			return false
		}
		times1[i] = t
	}
	for i, a2 := range arg2 {
		t, ok := a2.(time.Time)
		if !ok {
			return false
		}
		times2[i] = t
	}
	for _, a1 := range times1 {
		if slices.ContainsFunc(times2, a1.Before) {
			return true
		}
	}
	return false
}

// CumpleReglaExiste checa si la regla existe. Nota que regresa true si la
// llave existe y el valor al que apunta es un puntero nil.
func (sesion *Sesion) CumpleReglaExiste(desc1 DescrAtributo, recurso attrsRecurso) bool {
	arg1, err := sesion.obtenerAtributos(desc1, recurso)
	return err == nil && len(arg1) > 0
}
