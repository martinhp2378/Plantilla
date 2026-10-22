package handler

import (
	"PlantillaInstitucional/miperfil"
	"PlantillaInstitucional/miperfil/view/components"
	"PlantillaInstitucional/miperfil/view/layouts"
	"PlantillaInstitucional/shared/pagina"
	"PlantillaInstitucional/shared/sesionutil"
	"PlantillaInstitucional/shared/utils"
	"net/http"
	"path/filepath"

	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/archivos"
	"github.com/issemym-repos/sesiones"
)

// HandleVerPerfil despliega la pantalla principal de Mi Perfil.
func (h *MiPerfilHandler) HandleVerPerfil(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := sesionutil.UsuarioDeContexto(ctx)
	if u == nil || u.UuidPersona == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	uuidPersona, err := uuid.FromString(u.UuidPersona)
	if err != nil {
		http.Error(w, "Identificador de sesión inválido", http.StatusBadRequest)
		return
	}

	perfil, err := h.service.ObtenerPerfil(ctx, uuidPersona)
	if err != nil {
		h.logger.ErrorContext(ctx, "Error al obtener perfil de usuario", err, "id", uuidPersona)
		http.Error(w, "Error al cargar datos del perfil", http.StatusInternalServerError)
		return
	}

	imagenURL := archivos.Asset(ctx, filepath.Join("img", "perfil", sesionutil.ImagenPerfilValida(perfil.ImagenPerfil)))

	if r.Header.Get("HX-Request") == "true" {
		utils.RenderIgnore(ctx, w, layouts.MiPerfilContent(perfil, imagenURL))
		return
	}

	utils.RenderIgnore(ctx, w, layouts.ViewMiPerfil(perfil, imagenURL, pagina.MiPerfil))
}

// HandleCambiarPassword procesa el formulario de actualización de contraseña.
func (h *MiPerfilHandler) HandleCambiarPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := sesionutil.UsuarioDeContexto(ctx)
	if u == nil || u.UuidPersona == "" {
		http.Error(w, "Sesión no válida", http.StatusUnauthorized)
		return
	}

	uuidPersona, err := uuid.FromString(u.UuidPersona)
	if err != nil {
		http.Error(w, "Identificador de sesión inválido", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		utils.RenderIgnore(ctx, w, components.FeedbackMessage(miperfil.MensajeFeedback{
			Tipo:    "error",
			Mensaje: "Datos del formulario inválidos.",
		}))
		return
	}

	form := miperfil.CambioPasswordForm{
		PasswordActual:       r.FormValue("password_actual"),
		PasswordNueva:        r.FormValue("password_nueva"),
		PasswordConfirmacion: r.FormValue("password_confirmacion"),
	}

	if err := h.service.CambiarPassword(ctx, uuidPersona, form); err != nil {
		h.logger.WarnContext(ctx, "Fallo al cambiar contraseña de usuario", "error", err.Error(), "id", uuidPersona)
		utils.RenderIgnore(ctx, w, components.FeedbackMessage(miperfil.MensajeFeedback{
			Tipo:    "error",
			Mensaje: err.Error(),
		}))
		return
	}

	h.logger.InfoContext(ctx, "Contraseña actualizada exitosamente", "id", uuidPersona)
	utils.RenderIgnore(ctx, w, components.FeedbackMessage(miperfil.MensajeFeedback{
		Tipo:    "success",
		Mensaje: "¡Contraseña actualizada exitosamente!",
	}))
}

// HandleCambiarImagen procesa la subida de una nueva foto de perfil.
func (h *MiPerfilHandler) HandleCambiarImagen(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := sesionutil.UsuarioDeContexto(ctx)
	if u == nil || u.UuidPersona == "" {
		http.Error(w, "Sesión no válida", http.StatusUnauthorized)
		return
	}

	uuidPersona, err := uuid.FromString(u.UuidPersona)
	if err != nil {
		http.Error(w, "Identificador de sesión inválido", http.StatusBadRequest)
		return
	}

	imagenURLActual := sesionutil.UrlImagenPerfil(ctx)

	// Límite de 5MB
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		utils.RenderIgnore(ctx, w, components.AvatarContainer(imagenURLActual, miperfil.MensajeFeedback{
			Tipo:    "error",
			Mensaje: "El archivo es demasiado grande (Máximo 5MB).",
		}))
		return
	}

	file, header, err := r.FormFile("imagen")
	if err != nil {
		utils.RenderIgnore(ctx, w, components.AvatarContainer(imagenURLActual, miperfil.MensajeFeedback{
			Tipo:    "error",
			Mensaje: "Debe seleccionar un archivo de imagen.",
		}))
		return
	}
	defer file.Close()

	nuevoNombre, err := h.service.CambiarImagen(ctx, uuidPersona, header.Filename, header.Header.Get("Content-Type"), file, header.Size)
	if err != nil {
		h.logger.WarnContext(ctx, "Fallo al cambiar imagen de perfil", "error", err.Error(), "id", uuidPersona)
		utils.RenderIgnore(ctx, w, components.AvatarContainer(imagenURLActual, miperfil.MensajeFeedback{
			Tipo:    "error",
			Mensaje: err.Error(),
		}))
		return
	}

	// Actualizar la instancia de sesión en memoria
	if u != nil && h.constructorSesion != nil {
		u.ArchivoImagenPerfil = nuevoNombre
		ses := sesiones.RecuperarSesion(ctx)
		if ses != nil {
			_ = h.constructorSesion.AgregarInstancia(ses, *u)
		}
	}

	nuevaURL := archivos.Asset(ctx, filepath.Join("img", "perfil", sesionutil.ImagenPerfilValida(nuevoNombre)))
	h.logger.InfoContext(ctx, "Avatar de usuario actualizado exitosamente", "id", uuidPersona, "archivo", nuevoNombre)

	utils.RenderIgnore(ctx, w, components.AvatarContainer(nuevaURL, miperfil.MensajeFeedback{
		Tipo:    "success",
		Mensaje: "¡Foto de perfil actualizada exitosamente!",
	}))
}
