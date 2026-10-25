package handler

import (
	"PlantillaInstitucional/configuracion/usuario"
	"PlantillaInstitucional/configuracion/usuario/events"
	"PlantillaInstitucional/configuracion/usuario/view/components"
	errores "PlantillaInstitucional/shared/errores"
	password "PlantillaInstitucional/shared/password"
	"PlantillaInstitucional/shared/utils"
	validacion "PlantillaInstitucional/shared/validacion"
	"net/http"
)

// --- Actions ---

func (h *UsuarioHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Parsear
	usuarioDTO, err := h.parseFormToUsuario(r)
	if err != nil {
		h.renderErrorModalCreate(w, r, "Error en datos recibidos", nil, usuarioDTO)
		return
	}

	// 2. Generate an opaque one-time activation token. Its hash alone is persisted.
	activationToken, err := password.GenerateOpaqueToken()
	if err != nil {
		h.renderErrorModalCreate(w, r, "Error al generar activación", nil, usuarioDTO)
		return
	}

	// 3. Llamar al servicio (El servicio generará el UUID internamente)
	if err := h.service.CrearUsuario(ctx, usuarioDTO, activationToken); err != nil {
		// Manejo específico si es error de validación
		validationErrors := validacion.ValidationErrorsMap(err)
		h.renderErrorModalCreate(w, r, err.Error(), validationErrors, usuarioDTO)
		return
	}

	// 4. The notification is part of successful provisioning. A delivery error is
	// explicit and never rendered as success.
	emailEvent := events.SendUserActivationEvent{
		Email:         usuarioDTO.Email,
		FullName:      "",
		Username:      usuarioDTO.Email,
		ActivationURL: utils.GetBaseURL(r) + "/login/activar?token=" + activationToken,
	}
	if err := h.eventDispatcher.Dispatch(ctx, emailEvent); err != nil {
		if cleanupErr := h.service.EliminarUsuario(ctx, usuarioDTO.UuidPersona); cleanupErr != nil {
			h.logger.ErrorContext(ctx, "Activación no entregada y limpieza fallida", cleanupErr, "email", usuarioDTO.Email)
			h.renderErrorModalCreate(w, r, "Cuenta creada, pero no se pudo entregar la activación. Requiere atención administrativa.", nil, usuarioDTO)
			return
		}
		h.renderErrorModalCreate(w, r, "No se creó el usuario porque no se pudo entregar la activación.", nil, usuarioDTO)
		return
	}

	h.renderSuccess(w, r, errores.MensajeOperacionExitosaRegistroCreado)
}

func (h *UsuarioHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Parsear ID y Body

	id, err := h.parseIDParam(r)
	if err != nil {
		h.renderErrorGeneric(w, r, errores.TraducirError(err))
		return
	}

	usuarioDTO, err := h.parseFormToUsuario(r)
	if err != nil {
		h.renderErrorModalUpdate(w, r, "Datos inválidos", nil, usuarioDTO)
		return
	}
	// Asignamos el ID parseado de la URL al DTO
	usuarioDTO.UuidPersona = id

	// 2. Actualizar
	if err := h.service.ActualizarUsuario(ctx, usuarioDTO); err != nil {
		validationErrors := validacion.ValidationErrorsMap(err)
		h.renderErrorModalUpdate(w, r, err.Error(), validationErrors, usuarioDTO)
		return
	}

	h.renderSuccess(w, r, errores.MensajeOperacionExitosaRegistroActualizado)
}

func (h *UsuarioHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := h.parseIDParam(r)
	if err != nil {
		h.renderErrorGeneric(w, r, errores.TraducirError(err))
		return
	}

	if err := h.service.EliminarUsuario(ctx, id); err != nil {
		h.renderErrorGeneric(w, r, errores.TraducirError(err))
		return
	}

	h.renderSuccess(w, r, errores.MensajeOperacionExitosaRegistroEliminado)
}

// --- Modals & UI ---

func (h *UsuarioHandler) ModalCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	opcionesPersona, _ := h.personaService.ObtenerOpcionesPersona(ctx)
	utils.RenderIgnore(r.Context(), w, components.FormCreateUsuarioV2(usuario.Usuario{}, opcionesPersona, nil))
}

func (h *UsuarioHandler) ModalUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := h.parseIDParam(r)
	if err != nil {
		h.renderErrorGeneric(w, r, "ID no proporcionado")
		return
	}

	b, err := h.service.ObtenerUsuario(ctx, id)
	if err != nil {
		h.renderErrorGeneric(w, r, "No se encontró el usuario")
		return
	}

	opcionesPersona, _ := h.personaService.ObtenerOpcionesPersona(ctx)
	utils.RenderIgnore(ctx, w, components.FormUpdateUsuarioV2(&b, opcionesPersona, nil))
}

func (h *UsuarioHandler) CloseModal(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(`<div id="modal"></div>`))
}
