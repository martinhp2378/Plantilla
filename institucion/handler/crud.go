package handler

import (
	"PlantillaInstitucional/institucion"
	"PlantillaInstitucional/institucion/view/components"
	"PlantillaInstitucional/shared/db"
	errores "PlantillaInstitucional/shared/errores"
	"PlantillaInstitucional/shared/utils"
	"PlantillaInstitucional/shared/validacion"
	"net/http"
)

// --- Actions ---

func (h *personaHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Parsear
	personaDTO, err := h.parseFormToPersona(r)
	if err != nil {
		h.renderErrorModalCreate(w, r, "Error en datos recibidos", nil, personaDTO)
		return
	}

	// 2. Llamar al servicio (El servicio generará el UUID internamente)
	if err := h.service.CrearPersona(ctx, personaDTO); err != nil {
		// Manejo específico si es error de validación
		validationErrors := validacion.ValidationErrorsMap(err)
		h.renderErrorModalCreate(w, r, err.Error(), validationErrors, personaDTO)
		return
	}

	h.renderSuccess(w, r, errores.MensajeOperacionExitosaRegistroCreado)
}

func (h *personaHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Parsear ID y Body
	id, err := h.parseIDParam(r)
	if err != nil {
		h.renderErrorGeneric(w, r, errores.TraducirError(err))
		return
	}

	personaDTO, err := h.parseFormToPersona(r)
	if err != nil {
		h.renderErrorModalUpdate(w, r, "Datos inválidos", nil, personaDTO)
		return
	}
	// Asignamos el ID parseado de la URL al DTO
	personaDTO.UuidPersona, err = db.ParseUUID(id)
	if err != nil {
		h.renderErrorModalUpdate(w, r, "Id invalido", nil, personaDTO)
		return
	}

	// 2. Actualizar
	if err := h.service.ActualizarPersona(ctx, personaDTO); err != nil {
		validationErrors := validacion.ValidationErrorsMap(err)
		h.renderErrorModalUpdate(w, r, err.Error(), validationErrors, personaDTO)
		return
	}

	h.renderSuccess(w, r, errores.MensajeOperacionExitosaRegistroActualizado)
}

func (h *personaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idparam, err := h.parseIDParam(r)
	if err != nil {
		h.renderErrorGeneric(w, r, errores.TraducirError(err))
		return
	}

	id, err := db.ParseUUID(idparam)
	if err != nil {
		h.renderErrorGeneric(w, r, errores.TraducirError(err))
		return
	}

	if err := h.service.EliminarPersona(ctx, id); err != nil {
		h.renderErrorGeneric(w, r, errores.TraducirError(err))
		return
	}

	h.renderSuccess(w, r, errores.MensajeOperacionExitosaRegistroEliminado)
}

// --- Modals & UI ---

func (h *personaHandler) ModalCreate(w http.ResponseWriter, r *http.Request) {
	utils.RenderIgnore(r.Context(), w, components.FormCreatePersonaV2(institucion.Persona{}, nil))
}

func (h *personaHandler) ModalUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idparam, err := h.parseIDParam(r)
	if err != nil {
		h.renderErrorGeneric(w, r, "ID no proporcionado")
		return
	}

	id, err := db.ParseUUID(idparam)
	if err != nil {
		h.renderErrorGeneric(w, r, errores.TraducirError(err))
		return
	}

	b, err := h.service.ObtenerPersona(ctx, id)
	if err != nil {
		h.renderErrorGeneric(w, r, "No se encontró el persona")
		return
	}

	utils.RenderIgnore(ctx, w, components.FormUpdatePersonaV2(&b, nil))
}

func (h *personaHandler) CloseModal(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(`<div id="modal"></div>`))
}
