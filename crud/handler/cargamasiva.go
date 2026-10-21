package handler

import (
	"PlantillaInstitucional/crud/view/components"
	"PlantillaInstitucional/shared/utils"
	"fmt"
	"net/http"
)

func (h *crudHandler) ModalCargaMasiva(w http.ResponseWriter, r *http.Request) {
	utils.RenderIgnore(r.Context(), w, components.ModalCargaMasivaV2(nil, ""))
}

func (h *crudHandler) DescargarPlantillaCSV(w http.ResponseWriter, r *http.Request) {
	plantillaBytes, err := h.service.GenerarPlantillaCSV()
	if err != nil {
		h.renderErrorGeneric(w, r, "Error al generar la plantilla CSV")
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="plantilla_crud_ejemplo.csv"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(plantillaBytes)
}

func (h *crudHandler) CargaMasiva(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Protección contra DoS: limitar tamaño del cuerpo a 10 MB
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	mr, err := r.MultipartReader()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.RenderIgnore(ctx, w, components.ModalCargaMasivaV2(nil, "Debe enviar un formulario multipart válido"))
		return
	}

	for {
		part, err := mr.NextPart()
		if err != nil {
			break
		}

		if part.FormName() == "archivo" {
			resultado, cargaErr := h.service.CargarMasivo(ctx, part)
			_ = part.Close()

			if cargaErr != nil {
				w.Header().Set("Content-Type", "text/html")
				utils.RenderIgnore(ctx, w, components.ModalCargaMasivaV2(nil, cargaErr.Error()))
				return
			}

			// Si se detectaron errores de validación en una o más filas
			if !resultado.EsExitoso() {
				w.Header().Set("Content-Type", "text/html")
				utils.RenderIgnore(ctx, w, components.ModalCargaMasivaV2(resultado.Errores, ""))
				return
			}

			// Éxito: cerrar modal, actualizar la tabla con OOB swap y mostrar toast verde
			mensaje := fmt.Sprintf("Carga masiva exitosa: se insertaron %d registros correctamente", resultado.TotalValidos)
			h.renderSuccess(w, r, mensaje)
			return
		}

		_ = part.Close()
	}

	w.WriteHeader(http.StatusBadRequest)
	utils.RenderIgnore(ctx, w, components.ModalCargaMasivaV2(nil, "No se encontró el archivo CSV en el formulario"))
}
