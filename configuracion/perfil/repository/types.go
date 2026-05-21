package repository

import (
	bd "PlantillaInstitucional/app/db"
)

// PerfilRepository es la implementación del repositorio de perfiles.
type PerfilRepository struct {
	db *bd.AppPools
}

func NewPerfilRepository(db *bd.AppPools) *PerfilRepository {
	return &PerfilRepository{
		db: db,
	}
}
