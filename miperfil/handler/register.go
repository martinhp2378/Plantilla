package handler

import (
	"PlantillaInstitucional/app"
	"PlantillaInstitucional/miperfil/repository"
	"PlantillaInstitucional/miperfil/service"

	"github.com/go-chi/chi/v5"
)

// Register inicializa el repositorio, servicio, handler y monta las rutas de Mi Perfil.
func Register(a *app.App, r chi.Router) {
	repo := repository.NewMiPerfilRepository(a.Pools)
	svc := service.NewMiPerfilService(repo, a.ManejadorArchivos, a.Logger)
	h := NewMiPerfilHandler(svc, a.Logger, a.ConstructorSesion)
	h.RegisterRoutes(r)
}
