package handler

import (
	"PlantillaInstitucional/proveedor"
	"PlantillaInstitucional/shared/errores"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
)

func (h *ProveedorHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	item, err := h.parseForm(r)
	if err != nil {
		h.MandarErrorModal(ctx, w, "Datos invalidos", err)
		return
	}

	if err := h.svc.Crear(ctx, item); err != nil {
		h.MandarErrorModal(ctx, w, "Error al guardar registro", err)
		return
	}

	h.MandarSuccess(ctx, w, errores.MensajeOperacionExitosaRegistroCreado, proveedor.Rutas.Tabla)
}

func (h *ProveedorHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		h.MandarErrorModal(ctx, w, "ID invalido", err)
		return
	}

	item, err := h.parseForm(r)
	if err != nil {
		h.MandarErrorModal(ctx, w, "Datos invalidos", err)
		return
	}
	item.ID = id

	if err := h.svc.Actualizar(ctx, item); err != nil {
		h.MandarErrorModal(ctx, w, "Error al actualizar registro", err)
		return
	}

	h.MandarSuccess(ctx, w, errores.MensajeOperacionExitosaRegistroActualizado, proveedor.Rutas.Tabla)
}

func (h *ProveedorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		http.Error(w, "ID invalido", http.StatusBadRequest)
		return
	}

	if err := h.svc.Eliminar(ctx, id); err != nil {
		http.Error(w, "Error al eliminar registro", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
