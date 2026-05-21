package sesiones

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/mssola/useragent"
)

type InfoDispositivo struct {
	SO            string `json:"sO"`
	Browser       string `json:"browser"`
	Dispositivo   string `json:"dispositivo"`
	EsMovil       bool   `json:"es_movil"`
	EsTableta     bool   `json:"es_tableta"`
	EsDesktop     bool   `json:"es_desktop"`
	AnchoPantalla int    `json:"ancho_pantalla"`
	AltoPantalla  int    `json:"alto_pantalla"`
}

func obtenerIPCliente(r *http.Request) string {
	// Ver diferentes headers para el IP real
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	real := r.Header.Get("X-Real-IP")
	if real != "" {
		return real
	}

	return r.RemoteAddr
}

func parsearInfoDispositivo(agenteUsuario string) InfoDispositivo {
	au := useragent.New(agenteUsuario)
	browser, version := au.Browser()
	esMovil := au.Mobile()
	dispositivo := au.Platform()
	if esMovil {
		dispositivo = au.Model()
	}
	return InfoDispositivo{
		SO:          au.OS(),
		Browser:     fmt.Sprintf("%s %s", browser, version),
		Dispositivo: dispositivo,
		EsMovil:     esMovil,
		EsTableta:   false,
		EsDesktop:   !esMovil,
		// TODO: estos deberían ser recuperados por javascript
		AnchoPantalla: 0,
		AltoPantalla:  0,
	}
}

func generarHuella(r *http.Request) string {
	// Huella sencilla. TODO: hacerla más sofisticada.
	datos := r.UserAgent() + r.Header.Get("Accept-Language") + r.Header.Get("Accept-Encoding")
	res := base64.URLEncoding.EncodeToString([]byte(datos))
	if len(res) > 16 {
		return res[:16]
	}
	if res == "" {
		// no tiene nada de datos, hacer un uuid random
		res = uuid.Must(uuid.NewV4()).String()
	}
	return res
}
