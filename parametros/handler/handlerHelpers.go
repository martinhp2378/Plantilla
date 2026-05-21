package handler

import (
	"PlantillaInstitucional/parametros"
	"PlantillaInstitucional/parametros/view/components"
	"PlantillaInstitucional/parametros/view/layouts"
	"PlantillaInstitucional/shared/pagina"
	"PlantillaInstitucional/shared/utils"
	"PlantillaInstitucional/shared/views/components/compartidos"
	"context"
	"fmt"
	"net/http"
	"strings"
)

const baseUrl = "/configuracion/parametros/tabla"

func (h *ParametrosHandler) CloseModal(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprint(w, `<div id="modal"></div>`)
}

func (h *ParametrosHandler) Mandar_error(ctx context.Context, w http.ResponseWriter, r *http.Request, mensaje string) {
	var sb strings.Builder

	criterios, err := h.service.ListarParametros(ctx)
	if err != nil {
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		utils.RenderIgnore(ctx, &sb, components.TableV2(criterios, baseUrl, false))
	} else {
		utils.RenderIgnore(ctx, &sb, layouts.ViewParametros("Parámetros", pagina.Parametros, criterios, baseUrl))
	}
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("error", mensaje))

	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(sb.String()))
}

func (h *ParametrosHandler) Mandar_errorModalCreate(ctx context.Context, w http.ResponseWriter, mensaje string, errors map[string]string, parametro parametros.Parametro) {
	var sb strings.Builder
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("error", mensaje))
	utils.RenderIgnore(ctx, &sb, components.FormCreateParametroV2(parametro, errors))

	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(sb.String()))
}

func (h *ParametrosHandler) Mandar_errorModalUpdate(ctx context.Context, w http.ResponseWriter, mensaje string, errors map[string]string, parametro parametros.Parametro) {
	var sb strings.Builder
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("error", mensaje))
	utils.RenderIgnore(ctx, &sb, components.FormUpdateParametroV2(&parametro, errors))

	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(sb.String()))
}

func (h *ParametrosHandler) Mandar_success(ctx context.Context, w http.ResponseWriter, mensaje string, baseUrl string) {
	var sb strings.Builder

	criterios, err := h.service.ListarParametros(ctx)
	if err != nil {
		return
	}

	sb.WriteString(`<div id="modal" hx-swap-oob="true"></div>`)
	utils.RenderIgnore(ctx, &sb, components.TableV2(criterios, baseUrl, true))
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("success", mensaje))

	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(sb.String()))
}
