package handler

import (
	shared "PlantillaInstitucional/shared/errores"
	"PlantillaInstitucional/shared/utils"
	"PlantillaInstitucional/shared/views/components/compartidos"
	"net/http"
)

func (h *PerfilHandler) DownloadPDF(w http.ResponseWriter, r *http.Request) {
	url, err := h.service.ReportePdf(r.Context())
	if err != nil {
		h.renderErrorGeneric(w, r, shared.TraducirError(err))
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (h *PerfilHandler) DownloadExcel(w http.ResponseWriter, r *http.Request) {
	url, err := h.service.ReporteExcel(r.Context())
	if err != nil {
		h.renderErrorGeneric(w, r, shared.TraducirError(err))
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (h *PerfilHandler) SuccessFileMessage(w http.ResponseWriter, r *http.Request) {
	utils.RenderIgnore(r.Context(), w, compartidos.Mensaje("success", shared.MensajeOperacionExitosaArchivoGenerado))
}
