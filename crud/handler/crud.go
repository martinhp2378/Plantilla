package handler

import (
	"PlantillaInstitucional/crud"
	"PlantillaInstitucional/crud/view/components"
	"PlantillaInstitucional/shared/db"
	errores "PlantillaInstitucional/shared/errores"
	"PlantillaInstitucional/shared/utils"
	"PlantillaInstitucional/shared/validacion"
	"net/http"
)

func (h *crudHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	crudDTO, err := h.parseFormToCrud(r)
	if err != nil {
		h.renderErrorModalCreate(w, r, "Error en los datos recibidos", nil, crudDTO)
		return
	}

	if err := h.service.CrearCrud(ctx, crudDTO); err != nil {
		validationErrors := validacion.ValidationErrorsMap(err)
		h.renderErrorModalCreate(w, r, err.Error(), validationErrors, crudDTO)
		return
	}

	h.renderSuccess(w, r, errores.MensajeOperacionExitosaRegistroCreado)
}

func (h *crudHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr, err := h.parseIDParam(r)
	if err != nil {
		h.renderErrorGeneric(w, r, errores.TraducirError(err))
		return
	}

	crudDTO, err := h.parseFormToCrud(r)
	if err != nil {
		h.renderErrorModalUpdate(w, r, "Datos inválidos", nil, crudDTO)
		return
	}

	crudDTO.Id, err = db.ParseUUID(idStr)
	if err != nil {
		h.renderErrorModalUpdate(w, r, "ID inválido", nil, crudDTO)
		return
	}

	if err := h.service.ActualizarCrud(ctx, crudDTO); err != nil {
		validationErrors := validacion.ValidationErrorsMap(err)
		h.renderErrorModalUpdate(w, r, err.Error(), validationErrors, crudDTO)
		return
	}

	h.renderSuccess(w, r, errores.MensajeOperacionExitosaRegistroActualizado)
}

func (h *crudHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr, err := h.parseIDParam(r)
	if err != nil {
		h.renderErrorGeneric(w, r, errores.TraducirError(err))
		return
	}

	id, err := db.ParseUUID(idStr)
	if err != nil {
		h.renderErrorGeneric(w, r, errores.TraducirError(err))
		return
	}

	if err := h.service.EliminarCrud(ctx, id); err != nil {
		h.renderErrorGeneric(w, r, errores.TraducirError(err))
		return
	}

	h.renderSuccess(w, r, errores.MensajeOperacionExitosaRegistroEliminado)
}

func (h *crudHandler) ModalCreate(w http.ResponseWriter, r *http.Request) {
	utils.RenderIgnore(r.Context(), w, components.FormCreateCrudV2(crud.Crud{
		Booleano: true,
		Color:    "#691C32",
	}, nil))
}

func (h *crudHandler) ModalUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr, err := h.parseIDParam(r)
	if err != nil {
		h.renderErrorGeneric(w, r, "ID no proporcionado")
		return
	}

	id, err := db.ParseUUID(idStr)
	if err != nil {
		h.renderErrorGeneric(w, r, "ID inválido")
		return
	}

	item, err := h.service.ObtenerCrud(ctx, id)
	if err != nil {
		h.renderErrorGeneric(w, r, "No se encontró el registro solicitado")
		return
	}

	utils.RenderIgnore(ctx, w, components.FormUpdateCrudV2(&item, nil))
}

func (h *crudHandler) CloseModal(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<div id="modal" hx-swap-oob="true"></div>`))
}
