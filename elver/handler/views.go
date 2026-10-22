package handler

import (
	"PlantillaInstitucional/elver"
	"PlantillaInstitucional/elver/view/components"
	"PlantillaInstitucional/elver/view/layouts"
	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/shared/utils"
	"PlantillaInstitucional/shared/views/components/compartidos"
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
)

func (h *ElverHandler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	criterios := pagination.FromQuery[elver.Elver](r.URL.Query())
	data, err := h.svc.Listar(ctx, criterios)
	if err != nil {
		h.logger.ErrorContext(ctx, "Error listando elveres", err)
	}

	if r.Header.Get("HX-Request") == "true" {
		utils.RenderIgnore(ctx, w, components.TableV2(data, elver.Rutas.Tabla, false))
		return
	}
	utils.RenderIgnore(ctx, w, layouts.ViewElverV2("Elveres", data, elver.Rutas.Tabla))
}

func (h *ElverHandler) SearchTable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	criterios := pagination.FromQuery[elver.Elver](r.URL.Query())
	data, err := h.svc.Listar(ctx, criterios)
	if err != nil {
		h.logger.ErrorContext(ctx, "Error buscando en tabla elveres", err)
	}
	utils.RenderIgnore(ctx, w, components.TableV2(data, elver.Rutas.Tabla, false))
}

func (h *ElverHandler) ModalCreate(w http.ResponseWriter, r *http.Request) {
	utils.RenderIgnore(r.Context(), w, components.FormCreateElverV2(elver.Elver{IsActive: true}, nil))
}

func (h *ElverHandler) ModalUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		http.Error(w, "ID invalido", http.StatusBadRequest)
		return
	}

	item, err := h.svc.Obtener(ctx, id)
	if err != nil {
		http.Error(w, "No encontrado", http.StatusNotFound)
		return
	}

	utils.RenderIgnore(ctx, w, components.FormUpdateElverV2(&item, nil))
}

func (h *ElverHandler) MandarSuccess(ctx context.Context, w http.ResponseWriter, msg string, targetTable string) {
	var sb strings.Builder
	sb.WriteString("<div id=\"modal\" hx-swap-oob=\"true\"></div>")
	criterios := pagination.CriteriosPagina[elver.Elver]{Size: 10}
	data, _ := h.svc.Listar(ctx, criterios)
	utils.RenderIgnore(ctx, &sb, components.TableV2(data, targetTable, true))
	w.Header().Set("Content-Type", "text/html")
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("success", msg))
	_, _ = w.Write([]byte(sb.String()))
}

func (h *ElverHandler) MandarErrorModalCreate(ctx context.Context, w http.ResponseWriter, msg string, errs map[string]string, item elver.Elver) {
	var sb strings.Builder
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("error", msg))
	utils.RenderIgnore(ctx, &sb, components.FormCreateElverV2(item, errs))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusUnprocessableEntity)
	_, _ = w.Write([]byte(sb.String()))
}

func (h *ElverHandler) MandarErrorModalUpdate(ctx context.Context, w http.ResponseWriter, msg string, errs map[string]string, item *elver.Elver) {
	var sb strings.Builder
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("error", msg))
	utils.RenderIgnore(ctx, &sb, components.FormUpdateElverV2(item, errs))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusUnprocessableEntity)
	_, _ = w.Write([]byte(sb.String()))
}

func (h *ElverHandler) MandarErrorGeneric(w http.ResponseWriter, r *http.Request, msg string) {
	ctx := r.Context()
	var sb strings.Builder
	criterios := pagination.CriteriosPagina[elver.Elver]{Size: 10}
	data, _ := h.svc.Listar(ctx, criterios)
	utils.RenderIgnore(ctx, &sb, components.TableV2(data, elver.Rutas.Tabla, true))
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("error", msg))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(sb.String()))
}
