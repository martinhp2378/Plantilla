package repository

import (
	bd "PlantillaInstitucional/app/db"
)

type PolicyRepository struct {
	*bd.AppPools
}

func NewPolicyRepository(db *bd.AppPools) *PolicyRepository {
	return &PolicyRepository{AppPools: db}
}

type catalogItem struct {
	Nombre string `db:"nombre"`
}
