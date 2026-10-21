package handler

import (
	shared "PlantillaInstitucional/shared/errores"
	"PlantillaInstitucional/shared/utils"
	"PlantillaInstitucional/shared/views/components/compartidos"
	"net/http"
)

func (h *crudHandler) DownloadPDF(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	url, err := h.service.ReportePdf(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "Error al generar PDF de crud", err)
		h.renderErrorGeneric(w, r, shared.TraducirError(err))
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (h *crudHandler) DownloadExcel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	url, err := h.service.ReporteExcel(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "Error al generar Excel de crud", err)
		h.renderErrorGeneric(w, r, shared.TraducirError(err))
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (h *crudHandler) SuccessFileMessage(w http.ResponseWriter, r *http.Request) {
	utils.RenderIgnore(r.Context(), w, compartidos.Mensaje("success", shared.MensajeOperacionExitosaArchivoGenerado))
}
