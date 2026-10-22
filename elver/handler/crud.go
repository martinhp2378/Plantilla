package handler

import (
	"PlantillaInstitucional/elver"
	"PlantillaInstitucional/shared/errores"
	"PlantillaInstitucional/shared/validacion"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
)

func (h *ElverHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	item, err := h.parseForm(r)
	if err != nil {
		h.MandarErrorModalCreate(ctx, w, "Datos invalidos", nil, item)
		return
	}

	if err := h.svc.Crear(ctx, item); err != nil {
		valErrs := validacion.ValidationErrorsMap(err)
		h.MandarErrorModalCreate(ctx, w, "Error al guardar registro", valErrs, item)
		return
	}

	h.MandarSuccess(ctx, w, errores.MensajeOperacionExitosaRegistroCreado, elver.Rutas.Tabla)
}

func (h *ElverHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		h.MandarErrorGeneric(w, r, "ID invalido")
		return
	}

	item, err := h.parseForm(r)
	if err != nil {
		h.MandarErrorModalUpdate(ctx, w, "Datos invalidos", nil, &item)
		return
	}
	item.ID = id

	if err := h.svc.Actualizar(ctx, item); err != nil {
		valErrs := validacion.ValidationErrorsMap(err)
		h.MandarErrorModalUpdate(ctx, w, "Error al actualizar registro", valErrs, &item)
		return
	}

	h.MandarSuccess(ctx, w, errores.MensajeOperacionExitosaRegistroActualizado, elver.Rutas.Tabla)
}

func (h *ElverHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		h.MandarErrorGeneric(w, r, "ID invalido")
		return
	}

	if err := h.svc.Eliminar(ctx, id); err != nil {
		h.MandarErrorGeneric(w, r, "Error al eliminar registro")
		return
	}

	h.MandarSuccess(ctx, w, errores.MensajeOperacionExitosaRegistroEliminado, elver.Rutas.Tabla)
}
