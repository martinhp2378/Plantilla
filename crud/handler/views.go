package handler

import (
	"PlantillaInstitucional/crud"
	"PlantillaInstitucional/crud/view"
	"PlantillaInstitucional/crud/view/components"
	"PlantillaInstitucional/crud/view/layouts"
	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/shared/pagina"
	"PlantillaInstitucional/shared/utils"
	"PlantillaInstitucional/shared/views/components/compartidos"
	"net/http"
	"strings"
)

func (h *crudHandler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	criterios := pagination.NewCriteriosPaginaDefault[crud.Crud]()
	data, err := h.service.ListarCrud(ctx, criterios)
	if err != nil {
		_, msg := h.sqlErrorHandler.Handle(err)
		h.renderErrorGeneric(w, r, msg)
		return
	}

	params := view.ParamCrud{
		Titulo:    "Catálogo de Ejemplo CRUD",
		Pag:       pagina.Crud,
		Criterios: data,
		BaseURL:   crud.Rutas.Tabla,
	}

	if r.Header.Get("HX-Request") == "true" {
		utils.RenderIgnore(ctx, w, components.TableV2(data, crud.Rutas.Tabla, false))
		return
	}

	utils.RenderIgnore(ctx, w, layouts.PaginaListaCrudV2(params))
}

func (h *crudHandler) SearchTable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	criterios := pagination.FromQuery[crud.Crud](r.URL.Query())
	data, err := h.service.ListarCrud(ctx, criterios)
	if err != nil {
		_, msg := h.sqlErrorHandler.Handle(err)
		h.renderErrorGeneric(w, r, msg)
		return
	}

	utils.RenderIgnore(ctx, w, components.TableV2(data, crud.Rutas.Tabla, false))
}

func (h *crudHandler) renderErrorGeneric(w http.ResponseWriter, r *http.Request, msg string) {
	var sb strings.Builder
	ctx := r.Context()
	criterios := pagination.NewCriteriosPaginaDefault[crud.Crud]()
	data, _ := h.service.ListarCrud(ctx, criterios)

	if r.Header.Get("HX-Request") == "true" {
		utils.RenderIgnore(ctx, &sb, components.TableV2(data, crud.Rutas.Tabla, false))
	} else {
		params := view.ParamCrud{
			Titulo:    "Catálogo de Ejemplo CRUD",
			Pag:       pagina.Crud,
			Criterios: data,
			BaseURL:   crud.Rutas.Tabla,
		}
		utils.RenderIgnore(ctx, &sb, layouts.PaginaListaCrudV2(params))
	}
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("error", msg))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(sb.String()))
}

func (h *crudHandler) renderSuccess(w http.ResponseWriter, r *http.Request, msg string) {
	var sb strings.Builder
	ctx := r.Context()
	sb.WriteString(`<div id="modal" hx-swap-oob="true"></div>`)
	criterios := pagination.NewCriteriosPaginaDefault[crud.Crud]()
	data, _ := h.service.ListarCrud(ctx, criterios)
	utils.RenderIgnore(ctx, &sb, components.TableV2(data, crud.Rutas.Tabla, true))
	w.Header().Set("Content-Type", "text/html")
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("success", msg))
	_, _ = w.Write([]byte(sb.String()))
}

func (h *crudHandler) renderErrorModalCreate(w http.ResponseWriter, r *http.Request, msg string, errs map[string]string, c crud.Crud) {
	var sb strings.Builder
	utils.RenderIgnore(r.Context(), &sb, compartidos.Mensaje("error", msg))
	utils.RenderIgnore(r.Context(), &sb, components.FormCreateCrudV2(c, errs))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusUnprocessableEntity)
	_, _ = w.Write([]byte(sb.String()))
}

func (h *crudHandler) renderErrorModalUpdate(w http.ResponseWriter, r *http.Request, msg string, errs map[string]string, c crud.Crud) {
	var sb strings.Builder
	utils.RenderIgnore(r.Context(), &sb, compartidos.Mensaje("error", msg))
	utils.RenderIgnore(r.Context(), &sb, components.FormUpdateCrudV2(&c, errs))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusUnprocessableEntity)
	_, _ = w.Write([]byte(sb.String()))
}
