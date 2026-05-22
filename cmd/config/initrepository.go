package config

import (
	"PlantillaInstitucional/app"
	bd "PlantillaInstitucional/app/db"
	perfil_repo "PlantillaInstitucional/configuracion/perfil/repository"
	policy_repo "PlantillaInstitucional/configuracion/policy/repository"
	usuario_repo "PlantillaInstitucional/configuracion/usuario/repository"
	persona_repo "PlantillaInstitucional/institucion/repository"
	login_repo "PlantillaInstitucional/login/db"
	parametros_repo "PlantillaInstitucional/parametros/repository"
	"PlantillaInstitucional/shared/tenants"
)

func InitRepositories(a *app.App) {
	// Pools
	pools := &bd.AppPools{ReadPool: a.PqRead, WritePool: a.PqWrite}

	a.Repositories.Tenants = tenants.NewStaticRepository()
	a.Repositories.Usuario = usuario_repo.NewUsuarioRepository(pools)
	a.Repositories.Perfil = perfil_repo.NewPerfilRepository(pools)
	a.Repositories.Policy = policy_repo.NewPolicyRepository(pools)
	a.Repositories.Persona = persona_repo.NewPersonaRepository(pools)
	a.Repositories.Login = login_repo.NewRepositorioPsql(pools)
	a.Repositories.Parametro = parametros_repo.NewRepository(pools)
}

func WireResolvers(a *app.App) {
	a.Repositories.Login.PolicyResolver = a.Services.Policy
}
