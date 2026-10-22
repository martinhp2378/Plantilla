package config

import (
	"PlantillaInstitucional/app"
	policy_repo "PlantillaInstitucional/configuracion/policy/repository"
	"PlantillaInstitucional/shared/tenants"
)

func InitRepositories(a *app.App) {
	pools := a.Pools

	a.Repositories.Tenants = tenants.NewStaticRepository()
	a.Repositories.Policy = policy_repo.NewPolicyRepository(pools)
}
