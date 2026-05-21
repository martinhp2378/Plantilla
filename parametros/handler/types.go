package handler

import (
	"context"

	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/parametros"
	"PlantillaInstitucional/shared/db"
	errores "PlantillaInstitucional/shared/errores"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/logs"
)

// ParametrosServiceProvider define el contrato del servicio de parámetros.
type ParametrosServiceProvider interface {
	LoadParameters() error
	GetParameter(key string) string
	StartParameterMonitoring(updateChan <-chan string)
	ListarParametros(ctx context.Context) (pagination.CriteriosPagina[parametros.Parametro], error)
	ListarParametrosFiltrados(ctx context.Context, criterios pagination.CriteriosPagina[parametros.Parametro]) (pagination.CriteriosPagina[parametros.Parametro], error)
	ObtenerParametro(ctx context.Context, id string) (parametros.Parametro, error)
	CrearParametro(ctx context.Context, p parametros.Parametro) error
	ActualizarParametro(ctx context.Context, p parametros.Parametro) error
	EliminarParametro(ctx context.Context, id string) error
	ReportePdf(ctx context.Context) (string, error)
	ReporteExcel(ctx context.Context) (string, error)
}

func (h *ParametrosHandler) RegisterRoutes(r chi.Router) {
	r.Route("/configuracion/parametros", func(r chi.Router) {
		r.Get("/", h.CargarPagina_templ)
		r.Get("/tabla", h.CargarTabla_templ)
		r.Post("/", h.Create_templ)
		r.Get("/{id}", h.ModalUpdate_templ)
		r.Put("/{id}", h.Update_templ)
		r.Delete("/{id}", h.Delete_templ)
		r.Get("/nuevo", h.ModalCreate_templ)
		r.Get("/pdf", h.ReportePDF_templ)
		r.Get("/excel", h.ReporteExcel_templ)
	})
}

// ParametrosHandler es la implementación del handler de parámetros.
type ParametrosHandler struct {
	service         ParametrosServiceProvider
	sqlErrorHandler *errores.SQLErrorHandler
	logger          *logs.Logger
}

func NewHandler(svc ParametrosServiceProvider, logger *logs.Logger) *ParametrosHandler {
	return &ParametrosHandler{
		service:         svc,
		sqlErrorHandler: errores.NewSQLErrorHandler(logger),
		logger:          logger,
	}
}

type Handler = ParametrosHandler
func ParametroFromFormCreate(r *http.Request) (parametros.Parametro, error) {
	var parametrosRespuesta = parametros.Parametro{}
	var err error

	parametrosRespuesta.ParametroId, err = uuid.NewV4()
	if err != nil {
		return parametrosRespuesta, fmt.Errorf("error nuevoUUID %w", err)
	}

	parametrosRespuesta.Clave = r.FormValue("clave")
	parametrosRespuesta.Valor = r.FormValue("valor")
	parametrosRespuesta.Descripcion = r.FormValue("descripcion")
	parametrosRespuesta.TipoDato = r.FormValue("tipo_dato")
	parametrosRespuesta.IsActive = r.FormValue("is_active") == "on"
	parametrosRespuesta.RazonI = r.FormValue("razon_i")

	return parametrosRespuesta, nil
}

func ParametroFromFormUpdate(r *http.Request) (parametros.Parametro, error) {
	var parametrosRespuesta = parametros.Parametro{}
	var err error

	if err = r.ParseForm(); err != nil {
		return parametrosRespuesta, fmt.Errorf("error al parsear formulario: %w", err)
	}

	parametrosRespuesta.ParametroId, err = db.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		return parametrosRespuesta, err
	}

	parametrosRespuesta.Clave = r.FormValue("clave")
	parametrosRespuesta.Valor = r.FormValue("valor")
	parametrosRespuesta.Descripcion = r.FormValue("descripcion")
	parametrosRespuesta.TipoDato = r.FormValue("tipo_dato")
	parametrosRespuesta.IsActive = r.FormValue("is_active") == "on"
	parametrosRespuesta.RazonM = r.FormValue("razon_m")

	return parametrosRespuesta, nil
}

func ParametroFromFormSimple(r *http.Request) (parametros.Parametro, error) {
	var parametrosRespuesta = parametros.Parametro{}
	var err error

	if err = r.ParseForm(); err != nil {
		return parametrosRespuesta, fmt.Errorf("error al parsear formulario: %w", err)
	}

	parametrosRespuesta.Clave = r.FormValue("clave")
	parametrosRespuesta.Valor = r.FormValue("valor")

	return parametrosRespuesta, nil
}
