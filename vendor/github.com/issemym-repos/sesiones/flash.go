package sesiones

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

func generarId() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

type TipoFlash string

const (
	FlashExito       TipoFlash = "exito"
	FlashInfo        TipoFlash = "info"
	FlashAdvertencia TipoFlash = "advertencia"
	FlashError       TipoFlash = "error"
	FlashDebug       TipoFlash = "debug"
)

type MensajeFlash struct {
	Id              string         `json:"id"`
	Tipo            TipoFlash      `json:"tipo"`
	Titulo          string         `json:"titulo"`
	Contenido       string         `json:"contenido"`
	Datos           map[string]any `json:"datos"`
	FechaCreacion   time.Time      `json:"fecha_creacion"`
	FechaExpiracion time.Time      `json:"fecha_expiracion"`
	Escondido       bool           `json:"escondido"`
}

func (cs *ConstructorSesion) AgregarFlash(sesion *Sesion, tipo TipoFlash, titulo string, contenido string, datos ...map[string]any) {
	flash := MensajeFlash{
		Id:              generarId(),
		Tipo:            tipo,
		Titulo:          titulo,
		Contenido:       contenido,
		FechaCreacion:   time.Now().Local(),
		FechaExpiracion: time.Now().Local().Add(cs.config.FlashTTL),
	}
	if len(datos) > 0 {
		flash.Datos = datos[0]
	}

	sesion.MensajesFlash = append(sesion.MensajesFlash, flash)
	if cs.config.MaximosMensajesFlash > 0 && len(sesion.MensajesFlash) > cs.config.MaximosMensajesFlash {
		sesion.MensajesFlash = sesion.MensajesFlash[len(sesion.MensajesFlash)-cs.config.MaximosMensajesFlash:]
	}
	sesion.FechaActualizacion = time.Now().Local()
}

func (sesion *Sesion) ObtenerFlashes() []MensajeFlash {
	var activos []MensajeFlash
	ahora := time.Now().Local()

	for _, flash := range sesion.MensajesFlash {
		if !flash.Escondido && ahora.Before(flash.FechaExpiracion) {
			activos = append(activos, flash)
		}
	}

	return activos
}

func (sesion *Sesion) LimpiarFlashes() {
	sesion.MensajesFlash = []MensajeFlash{}
	sesion.FechaActualizacion = time.Now().Local()
}

func (sesion *Sesion) PopFlashes() []MensajeFlash {
	flashes := sesion.ObtenerFlashes()
	sesion.LimpiarFlashes()
	return flashes
}

func (sesion *Sesion) EsconderFlash(id string) {
	for _, flash := range sesion.MensajesFlash {
		if flash.Id == id {
			flash.Escondido = true
			sesion.FechaActualizacion = time.Now().Local()
			break
		}
	}
}
