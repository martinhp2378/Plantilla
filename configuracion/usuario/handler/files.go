package handler

import (
	shared "PlantillaInstitucional/shared/errores"
	"PlantillaInstitucional/shared/utils"
	"PlantillaInstitucional/shared/views/components/compartidos"
	"net/http"
)

func (h *UsuarioHandler) DownloadPDF(w http.ResponseWriter, r *http.Request) {
	url, err := h.service.ReportePdf(r.Context())
	if err != nil {
		h.renderErrorGeneric(w, r, shared.TraducirError(err))
		return
	}
	// Redireccionamos a la URL del archivo generado en el storage/public
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *UsuarioHandler) DownloadExcel(w http.ResponseWriter, r *http.Request) {
	url, err := h.service.ReporteExcel(r.Context())
	if err != nil {
		h.renderErrorGeneric(w, r, shared.TraducirError(err))
		return
	}
	// Redireccionamos a la URL del archivo generado
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *UsuarioHandler) SuccessFileMessage(w http.ResponseWriter, r *http.Request) {
	utils.RenderIgnore(r.Context(), w, compartidos.Mensaje("success", shared.MensajeOperacionExitosaArchivoGenerado))
}
