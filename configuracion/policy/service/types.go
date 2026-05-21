package service

import (
	"PlantillaInstitucional/configuracion/policy"
	"PlantillaInstitucional/pagination"
	"context"
)

type PolicyRepository interface {
	RouteAllowed(ctx context.Context, profile string, path string) (bool, error)
	ProfileHasRequiredPolicies(ctx context.Context, profile string) (bool, error)
	ListarPoliticas(ctx context.Context, criterios pagination.CriteriosPagina[policy.Politica]) ([]policy.Politica, int, error)
	ObtenerPolitica(ctx context.Context, nombre string) (policy.Politica, error)
	CrearPolitica(ctx context.Context, p policy.Politica) error
	ActualizarPolitica(ctx context.Context, p policy.Politica, nombreOriginal string) error
	EliminarPolitica(ctx context.Context, nombre string) error
	ListarPerfiles(ctx context.Context) ([]string, error)
	ListarRecursos(ctx context.Context) ([]string, error)
	ListarAcciones(ctx context.Context) ([]string, error)
}

type Service struct {
	repo PolicyRepository
}
