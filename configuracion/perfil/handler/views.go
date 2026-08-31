package handler

import (
	"PlantillaInstitucional/configuracion/perfil"
	"PlantillaInstitucional/configuracion/perfil/view"
	"PlantillaInstitucional/configuracion/perfil/view/components"
	"PlantillaInstitucional/configuracion/perfil/view/layouts"
	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/shared/pagina"
	"PlantillaInstitucional/shared/utils"
	"PlantillaInstitucional/shared/views/components/compartidos"
	"net/http"
	"strings"
)

func (h *PerfilHandler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Criterios vacíos por defecto trae la primera página
	criterios := pagination.NewCriteriosPaginaDefault[perfil.Perfil]()

	data, err := h.service.ListarPerfiles(ctx, criterios)
	if err != nil {
		_, msg := h.sqlErrorHandler.Handle(err)
		h.renderErrorGeneric(w, r, msg)
		return
	}

	params := view.ParamsPerfiles{
		Titulo:    "Perfiles",
		Pag:       pagina.Perfil,
		Criterios: data,
	}

	if r.Header.Get("HX-Request") == "true" {
		utils.RenderIgnore(ctx, w, layouts.PerfilContentV2(params))
		return
	}

	utils.RenderIgnore(ctx, w, layouts.ViewPerfilesV2(params))
}

func (h *PerfilHandler) SearchTable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	criterios := pagination.FromQuery[perfil.Perfil](r.URL.Query())
	data, err := h.service.ListarPerfiles(ctx, criterios)
	if err != nil {
		_, msg := h.sqlErrorHandler.Handle(err)
		h.renderErrorGeneric(w, r, msg)
		return
	}

	// Renderizamos solo la tabla fragmento para HTMX
	utils.RenderIgnore(ctx, w, components.TableV2(data, false))
}

// --- Render Helpers (Privados) ---

func (h *PerfilHandler) renderErrorGeneric(w http.ResponseWriter, r *http.Request, msg string) {
	var sb strings.Builder
	ctx := r.Context()

	criterios := pagination.NewCriteriosPaginaDefault[perfil.Perfil]()
	data, _ := h.service.ListarPerfiles(ctx, criterios)

	if r.Header.Get("HX-Request") == "true" {
		utils.RenderIgnore(ctx, &sb, components.TableV2(data, false))
	} else {
		params := view.ParamsPerfiles{
			Titulo:    "Perfiles",
			Pag:       pagina.Perfil,
			Criterios: data,
		}
		utils.RenderIgnore(ctx, &sb, layouts.ViewPerfilesV2(params))
	}
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("error", msg))

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(sb.String()))
}

func (h *PerfilHandler) renderSuccess(w http.ResponseWriter, r *http.Request, msg string) {
	var sb strings.Builder
	ctx := r.Context()

	sb.WriteString(`<div id="modal" hx-swap-oob="true"></div>`)

	criterios := pagination.NewCriteriosPaginaDefault[perfil.Perfil]()
	data, _ := h.service.ListarPerfiles(ctx, criterios)
	utils.RenderIgnore(ctx, &sb, components.TableV2(data, true))

	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("success", msg))

	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(sb.String()))
}

func (h *PerfilHandler) renderErrorModalCreate(w http.ResponseWriter, r *http.Request, msg string, errs map[string]string, b perfil.Perfil) {
	var sb strings.Builder
	utils.RenderIgnore(r.Context(), &sb, compartidos.Mensaje("error", msg))
	utils.RenderIgnore(r.Context(), &sb, components.FormCreatePerfilV2(b, errs))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusUnprocessableEntity)
	_, _ = w.Write([]byte(sb.String()))
}

func (h *PerfilHandler) renderErrorModalUpdate(w http.ResponseWriter, r *http.Request, msg string, errs map[string]string, b perfil.Perfil) {
	var sb strings.Builder
	utils.RenderIgnore(r.Context(), &sb, compartidos.Mensaje("error", msg))
	utils.RenderIgnore(r.Context(), &sb, components.FormUpdatePerfilV2(&b, errs))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusUnprocessableEntity)
	_, _ = w.Write([]byte(sb.String()))
}
