package handler

import (
	"PlantillaInstitucional/configuracion/perfil"
	"PlantillaInstitucional/configuracion/perfil/view/components"
	shared "PlantillaInstitucional/shared/errores"
	"PlantillaInstitucional/shared/utils"
	"PlantillaInstitucional/shared/validacion"
	"net/http"
)

// --- Actions ---

func (h *PerfilHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Parsear
	perfilDTO, err := h.parseFormToPerfil(r)
	if err != nil {
		h.renderErrorModalCreate(w, r, "Error en datos recibidos", nil, perfilDTO)
		return
	}

	// 2. Llamar al servicio (El servicio generará el UUID internamente)
	if err := h.service.CrearPerfil(ctx, perfilDTO); err != nil {
		// Manejo específico si es error de validación
		validationErrors := validacion.ValidationErrorsMap(err)
		h.renderErrorModalCreate(w, r, err.Error(), validationErrors, perfilDTO)
		return
	}

	h.renderSuccess(w, r, shared.MensajeOperacionExitosaRegistroCreado)
}

func (h *PerfilHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Parsear ID y Body

	nombreOld, err := h.parseNombreParam(r)
	if err != nil {
		h.renderErrorGeneric(w, r, shared.TraducirError(err))
		return
	}

	perfilDTO, err := h.parseFormToPerfil(r)
	if err != nil {
		h.renderErrorModalUpdate(w, r, "Datos inválidos", nil, perfilDTO)
		return
	}

	// 2. Actualizar
	if err := h.service.ActualizarPerfil(ctx, perfilDTO, nombreOld); err != nil {
		validationErrors := validacion.ValidationErrorsMap(err)
		h.renderErrorModalUpdate(w, r, err.Error(), validationErrors, perfilDTO)
		return
	}

	h.renderSuccess(w, r, shared.MensajeOperacionExitosaRegistroActualizado)
}

func (h *PerfilHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	nombre, err := h.parseNombreParam(r)
	if err != nil {
		h.renderErrorGeneric(w, r, shared.TraducirError(err))
		return
	}

	if err := h.service.EliminarPerfil(ctx, nombre); err != nil {
		h.renderErrorGeneric(w, r, shared.TraducirError(err))
		return
	}

	h.renderSuccess(w, r, shared.MensajeOperacionExitosaRegistroEliminado)
}

// --- Modals & UI ---

func (h *PerfilHandler) ModalCreate(w http.ResponseWriter, r *http.Request) {
	utils.RenderIgnore(r.Context(), w, components.FormCreatePerfilV2(perfil.Perfil{}, nil))
}

func (h *PerfilHandler) ModalUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	nombre, err := h.parseNombreParam(r)
	if err != nil {
		h.renderErrorGeneric(w, r, "Nombre no proporcionado")
		return
	}

	b, err := h.service.ObtenerPerfil(ctx, nombre)
	if err != nil {
		h.renderErrorGeneric(w, r, "No se encontró el perfil")
		return
	}

	utils.RenderIgnore(ctx, w, components.FormUpdatePerfilV2(&b, nil))
}

func (h *PerfilHandler) CloseModal(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(`<div id="modal"></div>`))
}
