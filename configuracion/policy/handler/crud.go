package handler

import (
	"PlantillaInstitucional/configuracion/policy"
	"PlantillaInstitucional/configuracion/policy/view/components"
	shared "PlantillaInstitucional/shared/errores"
	"PlantillaInstitucional/shared/utils"
	"net/http"
)

func (h *PolicyHandler) Create(w http.ResponseWriter, r *http.Request) {
	p, err := h.parseForm(r)
	if err != nil {
		h.renderErrorModalCreate(w, r, "Error en datos recibidos", p)
		return
	}
	if err := h.service.CrearPolitica(r.Context(), p); err != nil {
		h.renderErrorModalCreate(w, r, err.Error(), p)
		return
	}
	h.renderSuccess(w, r, shared.MensajeOperacionExitosaRegistroCreado)
}

func (h *PolicyHandler) Update(w http.ResponseWriter, r *http.Request) {
	nombreOld, err := h.parseNombre(r)
	if err != nil {
		h.renderErrorGeneric(w, r, shared.TraducirError(err))
		return
	}
	p, err := h.parseForm(r)
	if err != nil {
		h.renderErrorModalUpdate(w, r, "Datos inválidos", p)
		return
	}
	if err := h.service.ActualizarPolitica(r.Context(), p, nombreOld); err != nil {
		h.renderErrorModalUpdate(w, r, err.Error(), p)
		return
	}
	h.renderSuccess(w, r, shared.MensajeOperacionExitosaRegistroActualizado)
}

func (h *PolicyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	nombre, err := h.parseNombre(r)
	if err != nil {
		h.renderErrorGeneric(w, r, shared.TraducirError(err))
		return
	}
	if err := h.service.EliminarPolitica(r.Context(), nombre); err != nil {
		h.renderErrorGeneric(w, r, shared.TraducirError(err))
		return
	}
	h.renderSuccess(w, r, shared.MensajeOperacionExitosaRegistroEliminado)
}

func (h *PolicyHandler) ModalCreate(w http.ResponseWriter, r *http.Request) {
	perfiles, recursos, acciones, err := h.service.CatalogosPolitica(r.Context())
	if err != nil {
		h.renderErrorGeneric(w, r, err.Error())
		return
	}
	utils.RenderIgnore(r.Context(), w, components.FormCreatePolicyV2(policy.Politica{}, nil, perfiles, recursos, acciones))
}

func (h *PolicyHandler) ModalUpdate(w http.ResponseWriter, r *http.Request) {
	nombre, err := h.parseNombre(r)
	if err != nil {
		h.renderErrorGeneric(w, r, "Nombre no proporcionado")
		return
	}
	p, err := h.service.ObtenerPolitica(r.Context(), nombre)
	if err != nil {
		h.renderErrorGeneric(w, r, "No se encontró la política")
		return
	}
	perfiles, recursos, acciones, err := h.service.CatalogosPolitica(r.Context())
	if err != nil {
		h.renderErrorGeneric(w, r, err.Error())
		return
	}
	utils.RenderIgnore(r.Context(), w, components.FormUpdatePolicyV2(&p, nil, perfiles, recursos, acciones))
}

func (h *PolicyHandler) CloseModal(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte(`<div id="modal"></div>`))
}
