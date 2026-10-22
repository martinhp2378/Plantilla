package handler

import (
	shared "PlantillaInstitucional/shared/errores"
	"PlantillaInstitucional/shared/utils"
	"PlantillaInstitucional/shared/views/components/compartidos"
	"net/http"
)

func (h *ElverHandler) DownloadPDF(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	url, err := h.svc.ReportePdf(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "Error al generar PDF de elver", err)
		http.Error(w, shared.TraducirError(err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (h *ElverHandler) DownloadExcel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	url, err := h.svc.ReporteExcel(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "Error al generar Excel de elver", err)
		http.Error(w, shared.TraducirError(err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (h *ElverHandler) SuccessFileMessage(w http.ResponseWriter, r *http.Request) {
	utils.RenderIgnore(r.Context(), w, compartidos.Mensaje("success", shared.MensajeOperacionExitosaArchivoGenerado))
}
