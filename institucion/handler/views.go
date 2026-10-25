package handler

import (
	"PlantillaInstitucional/institucion"
	"PlantillaInstitucional/institucion/view"
	"PlantillaInstitucional/institucion/view/components"
	"PlantillaInstitucional/institucion/view/layouts"
	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/shared/pagina"
	"PlantillaInstitucional/shared/utils"
	"PlantillaInstitucional/shared/views/components/compartidos"
	"net/http"
	"strings"
)

func (h *personaHandler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	criterios := pagination.NewCriteriosPaginaDefault[institucion.Persona]()
	data, err := h.service.ListarPersonas(ctx, criterios)
	if err != nil {
		_, msg := h.sqlErrorHandler.Handle(err)
		h.renderErrorGeneric(w, r, msg)
		return
	}

	params := view.ParamPersona{
		Titulo:    "Catálogo de Personal",
		Pag:       pagina.Personas,
		Criterios: data,
		BaseURL:   institucion.Rutas.Tabla,
	}

	if r.Header.Get("HX-Request") == "true" {
		utils.RenderIgnore(ctx, w, components.TableV2(data, institucion.Rutas.Tabla, false))
		return
	}

	utils.RenderIgnore(ctx, w, layouts.PaginaListaPersonasV2(params))
}

func (h *personaHandler) SearchTable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	criterios := pagination.FromQuery[institucion.Persona](r.URL.Query())
	data, err := h.service.ListarPersonas(ctx, criterios)
	if err != nil {
		_, msg := h.sqlErrorHandler.Handle(err)
		h.renderErrorGeneric(w, r, msg)
		return
	}

	utils.RenderIgnore(ctx, w, components.TableV2(data, institucion.Rutas.Tabla, false))
}

// --- Render Helpers (Privados) ---

func (h *personaHandler) renderErrorGeneric(w http.ResponseWriter, r *http.Request, msg string) {
	var sb strings.Builder
	ctx := r.Context()
	criterios := pagination.NewCriteriosPaginaDefault[institucion.Persona]()
	data, _ := h.service.ListarPersonas(ctx, criterios)

	if r.Header.Get("HX-Request") == "true" {
		utils.RenderIgnore(ctx, &sb, components.TableV2(data, institucion.Rutas.Tabla, false))
	} else {
		params := view.ParamPersona{
			Titulo:    "Catálogo de Personal",
			Pag:       pagina.Personas,
			Criterios: data,
			BaseURL:   institucion.Rutas.Tabla,
		}
		utils.RenderIgnore(ctx, &sb, layouts.PaginaListaPersonasV2(params))
	}
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("error", msg))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(sb.String()))
}

func (h *personaHandler) renderSuccess(w http.ResponseWriter, r *http.Request, msg string) {
	var sb strings.Builder
	ctx := r.Context()
	sb.WriteString(`<div id="modal" hx-swap-oob="true"></div>`)
	criterios := pagination.NewCriteriosPaginaDefault[institucion.Persona]()
	data, _ := h.service.ListarPersonas(ctx, criterios)
	utils.RenderIgnore(ctx, &sb, components.TableV2(data, institucion.Rutas.Tabla, true))
	w.Header().Set("Content-Type", "text/html")
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("success", msg))
	_, _ = w.Write([]byte(sb.String()))
}

func (h *personaHandler) renderErrorModalCreate(w http.ResponseWriter, r *http.Request, msg string, errs map[string]string, p institucion.Persona) {
	var sb strings.Builder
	utils.RenderIgnore(r.Context(), &sb, compartidos.Mensaje("error", msg))
	utils.RenderIgnore(r.Context(), &sb, components.FormCreatePersonaV2(p, errs))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusUnprocessableEntity)
	_, _ = w.Write([]byte(sb.String()))
}

func (h *personaHandler) renderErrorModalUpdate(w http.ResponseWriter, r *http.Request, msg string, errs map[string]string, p institucion.Persona) {
	var sb strings.Builder
	utils.RenderIgnore(r.Context(), &sb, compartidos.Mensaje("error", msg))
	utils.RenderIgnore(r.Context(), &sb, components.FormUpdatePersonaV2(&p, errs))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusUnprocessableEntity)
	_, _ = w.Write([]byte(sb.String()))
}
