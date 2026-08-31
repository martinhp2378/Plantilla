package handler

import (
	"PlantillaInstitucional/configuracion/usuario"
	"PlantillaInstitucional/configuracion/usuario/view/components"
	"PlantillaInstitucional/configuracion/usuario/view/layouts"
	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/shared/pagina"
	"PlantillaInstitucional/shared/utils"
	sharedComp "PlantillaInstitucional/shared/views/components"
	"PlantillaInstitucional/shared/views/components/compartidos"
	"net/http"
	"strings"
)

func (h *UsuarioHandler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	criterios := pagination.NewCriteriosPaginaDefault[usuario.Usuario]()
	data, err := h.service.ListarUsuarios(ctx, criterios)
	if err != nil {
		_, msg := h.sqlErrorHandler.Handle(err)
		h.renderErrorGeneric(w, r, msg)
		return
	}

	// Si es peticion HTMX (desde el sidebar), mandamos solo el componente de tabla
	if r.Header.Get("HX-Request") == "true" {
		utils.RenderIgnore(ctx, w, layouts.UsuarioContentV2(data, pagina.GlobalRutas.Usuarios.Tabla))
		return
	}

	utils.RenderIgnore(ctx, w, layouts.ViewUsuariosV2("Usuarios", pagina.Usuarios, data, pagina.GlobalRutas.Usuarios.Tabla))
}

func (h *UsuarioHandler) SearchTable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	criterios := pagination.FromQuery[usuario.Usuario](r.URL.Query())
	data, err := h.service.ListarUsuarios(ctx, criterios)
	if err != nil {
		_, msg := h.sqlErrorHandler.Handle(err)
		h.renderErrorGeneric(w, r, msg)
		return
	}

	// Renderizamos solo la tabla fragmento para HTMX
	utils.RenderIgnore(ctx, w, components.TableV2(data, pagina.GlobalRutas.Usuarios.Tabla, false))
}

// --- Render Helpers (Privados) ---

func (h *UsuarioHandler) renderErrorGeneric(w http.ResponseWriter, r *http.Request, msg string) {
	var sb strings.Builder
	ctx := r.Context()

	criterios := pagination.NewCriteriosPaginaDefault[usuario.Usuario]()
	data, _ := h.service.ListarUsuarios(ctx, criterios)

	// Si es una peticion HTMX, mandamos solo el fragmento
	if r.Header.Get("HX-Request") == "true" {
		utils.RenderIgnore(ctx, &sb, components.TableV2(data, pagina.GlobalRutas.Usuarios.Tabla, false))
	} else {
		utils.RenderIgnore(ctx, &sb, layouts.ViewUsuariosV2("Usuarios", pagina.Usuarios, data, pagina.GlobalRutas.Usuarios.Tabla))
	}
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("error", msg))

	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(sb.String()))
}

func (h *UsuarioHandler) renderSuccess(w http.ResponseWriter, r *http.Request, msg string) {
	var sb strings.Builder
	ctx := r.Context()

	// 1. Limpiar Modal (OOB Swap)
	sb.WriteString(`<div id="modal" hx-swap-oob="true"></div>`)

	// 2. Recargar Tabla con datos frescos
	criterios := pagination.NewCriteriosPaginaDefault[usuario.Usuario]()
	data, _ := h.service.ListarUsuarios(ctx, criterios)
	utils.RenderIgnore(ctx, &sb, components.TableV2(data, pagina.GlobalRutas.Usuarios.Tabla, true))

	// 3. Mostrar Toast/Mensaje
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("success", msg))

	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(sb.String()))
}

func (h *UsuarioHandler) renderErrorModalCreate(w http.ResponseWriter, r *http.Request, msg string, errs map[string]string, b usuario.Usuario) {
	var sb strings.Builder
	utils.RenderIgnore(r.Context(), &sb, compartidos.Mensaje("error", msg))

	personas, err := h.personaService.ObtenerTodasCorto(r.Context())
	var opcionesPersona []sharedComp.SelectOption
	if err == nil {
		for _, p := range personas {
			opcionesPersona = append(opcionesPersona, sharedComp.SelectOption{
				Value: p.UuidPersona.String(),
				Text:  p.GetNombreCompleto(),
			})
		}
	}

	utils.RenderIgnore(r.Context(), &sb, components.FormCreateUsuarioV2(b, opcionesPersona, errs))
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(sb.String()))
}

func (h *UsuarioHandler) renderErrorModalUpdate(w http.ResponseWriter, r *http.Request, msg string, errs map[string]string, b usuario.Usuario) {
	var sb strings.Builder
	utils.RenderIgnore(r.Context(), &sb, compartidos.Mensaje("error", msg))

	personas, err := h.personaService.ObtenerTodasCorto(r.Context())
	var opcionesPersona []sharedComp.SelectOption
	if err == nil {
		for _, p := range personas {
			opcionesPersona = append(opcionesPersona, sharedComp.SelectOption{
				Value: p.UuidPersona.String(),
				Text:  p.GetNombreCompleto(),
			})
		}
	}

	utils.RenderIgnore(r.Context(), &sb, components.FormUpdateUsuarioV2(&b, opcionesPersona, errs))
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(sb.String()))
}
