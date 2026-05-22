package handler

import (
	"net/http"

	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/parametros"
	"PlantillaInstitucional/parametros/view/components"
	"PlantillaInstitucional/parametros/view/layouts"
	errores "PlantillaInstitucional/shared/errores"
	"PlantillaInstitucional/shared/pagina"
	"PlantillaInstitucional/shared/utils"
	"PlantillaInstitucional/shared/validacion"
	shared_components "PlantillaInstitucional/shared/views/components/compartidos"

	"github.com/go-chi/chi/v5"
)

func (h *ParametrosHandler) Create_templ(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		h.Mandar_errorModalCreate(ctx, w, err.Error(), map[string]string{}, parametros.Parametro{})
		return
	}

	Parametro, err := ParametroFromFormCreate(r)
	if err != nil {
		h.Mandar_errorModalCreate(ctx, w, err.Error(), map[string]string{}, parametros.Parametro{})
		return
	}

	if err := h.service.CrearParametro(ctx, Parametro); err != nil {
		mapa := validacion.ValidationErrorsMap(err)
		h.Mandar_errorModalCreate(ctx, w, err.Error(), mapa, Parametro)
		return
	}

	h.Mandar_success(ctx, w, errores.MensajeOperacionExitosaRegistroCreado, pagina.GlobalRutas.Parametros.Tabla)
}

func (h *ParametrosHandler) Update_templ(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		h.Mandar_errorModalUpdate(ctx, w, err.Error(), map[string]string{}, parametros.Parametro{})
		return
	}

	Parametro, err := ParametroFromFormUpdate(r)
	if err != nil {
		mapa := validacion.ValidationErrorsMap(err)
		h.Mandar_errorModalUpdate(ctx, w, err.Error(), mapa, parametros.Parametro{})
		return
	}

	if err := h.service.ActualizarParametro(ctx, Parametro); err != nil {
		mapa := validacion.ValidationErrorsMap(err)
		h.Mandar_errorModalUpdate(ctx, w, err.Error(), mapa, Parametro)
		return
	}

	h.Mandar_success(ctx, w, errores.MensajeOperacionExitosaRegistroActualizado, pagina.GlobalRutas.Parametros.Tabla)
}

func (h *ParametrosHandler) Delete_templ(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idParam := chi.URLParam(r, "id")

	if err := h.service.EliminarParametro(ctx, idParam); err != nil {
		h.Mandar_error(ctx, w, r, err.Error())
		return
	}

	h.Mandar_success(ctx, w, errores.MensajeOperacionExitosaRegistroEliminado, pagina.GlobalRutas.Parametros.Tabla)
}

func (h *ParametrosHandler) CargarPagina_templ(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	criterios, err := h.service.ListarParametros(ctx)
	if err != nil {
		_, msg := h.sqlErrorHandler.Handle(err)
		h.Mandar_error(ctx, w, r, msg)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		utils.RenderIgnore(ctx, w, components.TableV2(criterios, pagina.GlobalRutas.Parametros.Tabla, false))
	} else {
		utils.RenderIgnore(ctx, w, layouts.ViewParametros("Parámetros", pagina.Parametros, criterios, pagina.GlobalRutas.Parametros.Tabla))
	}
}

func (h *ParametrosHandler) CargarTabla_templ(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			h.Mandar_error(ctx, w, r, "Error en parámetros")
			return
		}
	}

	c := pagination.FromQuery[parametros.Parametro](r.URL.Query())
	criterios, err := h.service.ListarParametrosFiltrados(ctx, c)
	if err != nil {
		_, msg := h.sqlErrorHandler.Handle(err)
		h.Mandar_error(ctx, w, r, msg)
		return
	}

	utils.RenderIgnore(ctx, w, components.TableV2(criterios, pagina.GlobalRutas.Parametros.Tabla, false))
}

func (h *ParametrosHandler) ModalCreate_templ(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	utils.RenderIgnore(ctx, w, components.FormCreateParametroV2(parametros.Parametro{}, map[string]string{}))
}

func (h *ParametrosHandler) ModalUpdate_templ(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idParam := chi.URLParam(r, "id")

	parametro, err := h.service.ObtenerParametro(ctx, idParam)
	if err != nil {
		h.Mandar_error(ctx, w, r, err.Error())
		return
	}

	utils.RenderIgnore(ctx, w, components.FormUpdateParametroV2(&parametro, map[string]string{}))
}

func (h *ParametrosHandler) ReportePDF_templ(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	url, err := h.service.ReportePdf(ctx)
	if err != nil {
		h.Mandar_error(ctx, w, r, err.Error())
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *ParametrosHandler) ReporteExcel_templ(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	url, err := h.service.ReporteExcel(ctx)
	if err != nil {
		h.Mandar_error(ctx, w, r, err.Error())
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *ParametrosHandler) SuccessFileMessage(w http.ResponseWriter, r *http.Request) {
	utils.RenderIgnore(r.Context(), w, shared_components.Mensaje("success", errores.MensajeOperacionExitosaArchivoGenerado))
}
