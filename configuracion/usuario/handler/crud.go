package handler

import (
	"PlantillaInstitucional/configuracion/usuario"
	"PlantillaInstitucional/configuracion/usuario/events"
	"PlantillaInstitucional/configuracion/usuario/view/components"
	errores "PlantillaInstitucional/shared/errores"
	password "PlantillaInstitucional/shared/password"
	"PlantillaInstitucional/shared/utils"
	validacion "PlantillaInstitucional/shared/validacion"
	sharedComp "PlantillaInstitucional/shared/views/components"
	"net/http"
)

const baseUrl = "/configuracion/usuarios/tabla"

// --- Actions ---

func (h *UsuarioHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Parsear
	usuarioDTO, err := h.parseFormToUsuario(r)
	if err != nil {
		h.renderErrorModalCreate(w, r, "Error en datos recibidos", nil, usuarioDTO)
		return
	}

	// 2. Generar password temporal para enviar por email
	rawPassword, err := password.GenerateRandomPassword()
	if err != nil {
		h.renderErrorModalCreate(w, r, "Error al generar contraseña", nil, usuarioDTO)
		return
	}

	// 3. Llamar al servicio (El servicio generará el UUID internamente)
	if err := h.service.CrearUsuario(ctx, usuarioDTO, rawPassword); err != nil {
		// Manejo específico si es error de validación
		validationErrors := validacion.ValidationErrorsMap(err)
		h.renderErrorModalCreate(w, r, err.Error(), validationErrors, usuarioDTO)
		return
	}

	// 4. Enviar email con credenciales al usuario
	emailEvent := events.SendUserCredentialsEvent{
		Email:          usuarioDTO.Email,
		FullName:       "", // TODO: obtener nombre de la persona si fuera necesario
		Username:       usuarioDTO.Email,
		Password:       rawPassword,
		CredentialsURL: utils.GetBaseURL(r) + "/login",
	}
	_ = h.eventDispatcher.Dispatch(ctx, emailEvent)

	// No verificamos el error del dispatch - es fire-and-forget asíncrono

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
	personas, err := h.personaService.ObtenerTodasCorto(ctx)
	var opcionesPersona []sharedComp.SelectOption
	if err == nil {
		for _, p := range personas {
			opcionesPersona = append(opcionesPersona, sharedComp.SelectOption{
				Value: p.UuidPersona.String(),
				Text:  p.GetNombreCompleto(),
			})
		}
	}

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

	personas, err := h.personaService.ObtenerTodasCorto(ctx)
	var opcionesPersona []sharedComp.SelectOption
	if err == nil {
		for _, p := range personas {
			opcionesPersona = append(opcionesPersona, sharedComp.SelectOption{
				Value: p.UuidPersona.String(),
				Text:  p.GetNombreCompleto(),
			})
		}
	}

	utils.RenderIgnore(ctx, w, components.FormUpdateUsuarioV2(&b, opcionesPersona, nil))
}

func (h *UsuarioHandler) CloseModal(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(`<div id="modal"></div>`))
}
