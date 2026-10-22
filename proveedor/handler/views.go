package handler

import (
	"PlantillaInstitucional/proveedor"
	"PlantillaInstitucional/proveedor/view/components"
	"PlantillaInstitucional/proveedor/view/layouts"
	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/shared/utils"
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
)

func (h *ProveedorHandler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	criterios := pagination.FromQuery[proveedor.Proveedor](r.URL.Query())
	data, err := h.svc.Listar(ctx, criterios)
	if err != nil {
		h.logger.ErrorContext(ctx, "Error listando proveedores", err)
	}

	if r.Header.Get("HX-Request") == "true" {
		utils.RenderIgnore(ctx, w, components.TableV2(data, proveedor.Rutas.Tabla, false))
		return
	}
	utils.RenderIgnore(ctx, w, layouts.ViewProveedorV2("Proveedores", data, proveedor.Rutas.Tabla))
}

func (h *ProveedorHandler) SearchTable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	criterios := pagination.FromQuery[proveedor.Proveedor](r.URL.Query())
	data, err := h.svc.Listar(ctx, criterios)
	if err != nil {
		h.logger.ErrorContext(ctx, "Error buscando en tabla proveedores", err)
	}
	utils.RenderIgnore(ctx, w, components.TableV2(data, proveedor.Rutas.Tabla, false))
}

func (h *ProveedorHandler) ModalCreate(w http.ResponseWriter, r *http.Request) {
	utils.RenderIgnore(r.Context(), w, components.FormCreateProveedorV2(proveedor.Proveedor{IsActive: true}, nil))
}

func (h *ProveedorHandler) ModalUpdate(w http.ResponseWriter, r *http.Request) {
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

	utils.RenderIgnore(ctx, w, components.FormUpdateProveedorV2(&item, nil))
}

func (h *ProveedorHandler) MandarSuccess(ctx context.Context, w http.ResponseWriter, msg string, targetTable string) {
	w.Header().Set("HX-Trigger", "cerrarModal")
	criterios := pagination.CriteriosPagina[proveedor.Proveedor]{Size: 10}
	data, _ := h.svc.Listar(ctx, criterios)
	var sb strings.Builder
	utils.RenderIgnore(ctx, &sb, components.TableV2(data, targetTable, true))
	_, _ = w.Write([]byte(sb.String()))
}

func (h *ProveedorHandler) MandarErrorModal(ctx context.Context, w http.ResponseWriter, msg string, err error) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	errMap := map[string]string{"general": err.Error()}
	utils.RenderIgnore(ctx, w, components.FormCreateProveedorV2(proveedor.Proveedor{}, errMap))
}
