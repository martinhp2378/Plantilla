package handler

import (
	shared "PlantillaInstitucional/shared/errores"
	"PlantillaInstitucional/shared/utils"
	"PlantillaInstitucional/shared/views/components/compartidos"
	"net/http"
)

func (h *personaHandler) DownloadPDF(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	url, err := h.service.ReportePdf(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "Error al generar PDF", err)
		h.renderErrorGeneric(w, r, shared.TraducirError(err))
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (h *personaHandler) DownloadExcel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	url, err := h.service.ReporteExcel(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "Error al generar Excel", err)
		h.renderErrorGeneric(w, r, shared.TraducirError(err))
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (h *personaHandler) SuccessFileMessage(w http.ResponseWriter, r *http.Request) {
	utils.RenderIgnore(r.Context(), w, compartidos.Mensaje("success", shared.MensajeOperacionExitosaArchivoGenerado))
}
