package config

import (
	"PlantillaInstitucional/app"
	policy_repo "PlantillaInstitucional/configuracion/policy/repository"
	login_repo "PlantillaInstitucional/login/db"
	"PlantillaInstitucional/shared/tenants"
)

func InitRepositories(a *app.App) {
	pools := a.Pools

	a.Repositories.Tenants = tenants.NewStaticRepository()
	a.Repositories.Policy = policy_repo.NewPolicyRepository(pools)
	a.Repositories.Login = login_repo.NewRepositorioPsql(pools)
}

func WireResolvers(a *app.App) {
	a.Repositories.Login.PolicyResolver = a.Services.Policy
}
