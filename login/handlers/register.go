package handlers

import (
	"PlantillaInstitucional/app"
	login_repo "PlantillaInstitucional/login/db"

	"github.com/go-chi/chi/v5"
)

// Register inicializa el repositorio, inyecta PolicyResolver, construye el handler y monta las rutas públicas de Login.
func Register(a *app.App, r chi.Router) {
	repo := login_repo.NewRepositorioPsql(a.Pools)
	repo.PolicyResolver = a.Services.Policy
	h := NuevoHandlerLogin(repo, repo, a.ServicioEmail, a.Logger, a.ConstructorSesion, a.ConstructorSesionJWT)
	h.RegisterRoutes(r)
}
