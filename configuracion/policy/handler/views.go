package handler

import (
	"PlantillaInstitucional/configuracion/policy"
	"PlantillaInstitucional/configuracion/policy/view"
	"PlantillaInstitucional/configuracion/policy/view/components"
	"PlantillaInstitucional/configuracion/policy/view/layouts"
	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/shared/pagina"
	"PlantillaInstitucional/shared/utils"
	"PlantillaInstitucional/shared/views/components/compartidos"
	"net/http"
	"strings"
)

func (h *PolicyHandler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	criterios := pagination.NewCriteriosPaginaDefault[policy.Politica]()
	data, err := h.service.ListarPoliticas(ctx, criterios)
	if err != nil {
		h.renderErrorGeneric(w, r, err.Error())
		return
	}

	params := view.ParamsPoliticas{Titulo: "Políticas ABAC", Pag: pagina.Politicas, Criterios: data}
	if r.Header.Get("HX-Request") == "true" {
		utils.RenderIgnore(ctx, w, layouts.PolicyContentV2(params))
		return
	}
	utils.RenderIgnore(ctx, w, layouts.ViewPoliciesV2(params))
}

func (h *PolicyHandler) SearchTable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	criterios := pagination.FromQuery[policy.Politica](r.URL.Query())
	data, err := h.service.ListarPoliticas(ctx, criterios)
	if err != nil {
		h.renderErrorGeneric(w, r, err.Error())
		return
	}
	utils.RenderIgnore(ctx, w, components.TableV2(data, false))
}

func (h *PolicyHandler) renderErrorGeneric(w http.ResponseWriter, r *http.Request, msg string) {
	var sb strings.Builder
	ctx := r.Context()
	criterios := pagination.NewCriteriosPaginaDefault[policy.Politica]()
	data, _ := h.service.ListarPoliticas(ctx, criterios)
	if r.Header.Get("HX-Request") == "true" {
		utils.RenderIgnore(ctx, &sb, components.TableV2(data, false))
	} else {
		utils.RenderIgnore(ctx, &sb, layouts.ViewPoliciesV2(view.ParamsPoliticas{Titulo: "Políticas ABAC", Pag: pagina.Politicas, Criterios: data}))
	}
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("error", msg))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(sb.String()))
}

func (h *PolicyHandler) renderSuccess(w http.ResponseWriter, r *http.Request, msg string) {
	var sb strings.Builder
	ctx := r.Context()
	sb.WriteString(`<div id="modal" hx-swap-oob="true"></div>`)
	criterios := pagination.NewCriteriosPaginaDefault[policy.Politica]()
	data, _ := h.service.ListarPoliticas(ctx, criterios)
	utils.RenderIgnore(ctx, &sb, components.TableV2(data, true))
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("success", msg))
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(sb.String()))
}

func (h *PolicyHandler) renderErrorModalCreate(w http.ResponseWriter, r *http.Request, msg string, p policy.Politica) {
	var sb strings.Builder
	perfiles, recursos, acciones, _ := h.service.CatalogosPolitica(r.Context())
	utils.RenderIgnore(r.Context(), &sb, compartidos.Mensaje("error", msg))
	utils.RenderIgnore(r.Context(), &sb, components.FormCreatePolicyV2(p, nil, perfiles, recursos, acciones))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusUnprocessableEntity)
	_, _ = w.Write([]byte(sb.String()))
}

func (h *PolicyHandler) renderErrorModalUpdate(w http.ResponseWriter, r *http.Request, msg string, p policy.Politica) {
	var sb strings.Builder
	perfiles, recursos, acciones, _ := h.service.CatalogosPolitica(r.Context())
	utils.RenderIgnore(r.Context(), &sb, compartidos.Mensaje("error", msg))
	utils.RenderIgnore(r.Context(), &sb, components.FormUpdatePolicyV2(&p, nil, perfiles, recursos, acciones))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusUnprocessableEntity)
	_, _ = w.Write([]byte(sb.String()))
}
