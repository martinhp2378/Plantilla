package handlers
import (
	"encoding/json"
	"errors"
	"net/http"


	"PlantillaInstitucional/login/views"
	"PlantillaInstitucional/login/views/layouts"
	errores "PlantillaInstitucional/shared/errores"
	"PlantillaInstitucional/shared/utils"

	"github.com/go-chi/chi/v5"
	"github.com/issemym-repos/sesiones"
)

func (h *HandlerLogin) RegisterRoutes(r chi.Router) {
	r.Get("/", h.HandlePaginaInicio)
	r.Get("/login", h.HandlePaginaIngresar)
	r.Post("/login", h.HandlePostIngresar)
	r.Get("/login/salir", h.HandleCerrarSesion)
	r.Get("/sesion", h.HandleMostrarSesion)
}

func (h *HandlerLogin) HandlePaginaInicio(w http.ResponseWriter, r *http.Request) {
	sesion := sesiones.RecuperarSesion(r.Context())
	if sesion.EstaFirmado() {
		http.Redirect(w, r, "/institucion/personas", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (h *HandlerLogin) HandlePaginaIngresar(w http.ResponseWriter, r *http.Request) {
	formaLogin := views.FormaLogin{}
	utils.RenderIgnore(r.Context(), w, layouts.LoginV2(formaLogin))
}

func (h *HandlerLogin) HandleCerrarSesion(w http.ResponseWriter, r *http.Request) {
	sesion := sesiones.RecuperarSesion(r.Context())
	sesion.CerrarSesion()
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (h *HandlerLogin) HandleMostrarSesion(w http.ResponseWriter, r *http.Request) {
	sesion := sesiones.RecuperarSesion(r.Context())

	w.Header().Set("Content-Type", "application/json")
	bs, err := json.Marshal(sesion)
	if err != nil {
		h.Logger.Panic("Error en json.Marshal(sesion)", err)
	}
	_, _ = w.Write(bs)
}

func (h *HandlerLogin) HandlePostIngresar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	formaLogin := &views.FormaLogin{
		Email:    email,
		Password: password,
	}

	datosUsuario, err := h.RepositorioLogin.ValidarEmailPassword(ctx, email, password)
	if err != nil {
		if !errors.Is(err, errores.ErrCredencialesInvalidas) && !errors.Is(err, errores.ErrRegistroNoEncontrado) {
			panic(err)
		}

		if errors.Is(err, errores.ErrCredencialesInvalidas) {
			formaLogin.Error = "Contraseña incorrecta"
		} else {
			formaLogin.Error = "Usuario no encontrado"
		}

		utils.RenderIgnore(r.Context(), w, layouts.LoginV2(*formaLogin))
		return
	}

	sesion := sesiones.RecuperarSesion(ctx)

	datosExtra := make(map[string]any)
	datosExtra["email"] = email

	err = sesion.FirmarPersona(datosUsuario.UuidPersona, datosUsuario.Perfil, datosExtra)
	if err != nil {
		panic(err)
	}

	// Redirigir según perfil: administrador va al dashboard
	if datosUsuario.Perfil == sesiones.PerfilSuperAdmin {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/institucion/personas", http.StatusFound)
}
