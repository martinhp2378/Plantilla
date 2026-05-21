package db

import (
	"context"
	bd "PlantillaInstitucional/app/db"

	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/logs"
)

type RepositorioPsql struct {
	*bd.AppPools
	Logger         *logs.Logger
	PolicyResolver PolicyResolver
}

func NewRepositorioPsql(pools *bd.AppPools) *RepositorioPsql {
	return &RepositorioPsql{
		AppPools: pools,
	}
}

type PolicyResolver interface {
	IsRouteAllowed(ctx context.Context, profile string, path string) (bool, error)
	IsProfileActive(ctx context.Context, profile string) (bool, error)
}

type RowUsuario struct {
	UuidPersona uuid.UUID `db:"uuid_persona"`
	Email       string    `db:"email"`
	Password    string    `db:"password"`
	Perfil      string    `db:"perfil"`
}
