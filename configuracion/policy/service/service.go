package service

import (
	"PlantillaInstitucional/configuracion/policy"
	"PlantillaInstitucional/pagination"
	"context"
	"fmt"
	"strings"
)

func NewService(repo PolicyRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) IsRouteAllowed(ctx context.Context, profile string, path string) (bool, error) {
	return s.IsRouteAllowedWithMethod(ctx, profile, "*", path)
}

func (s *Service) IsRouteAllowedWithMethod(ctx context.Context, profile string, method string, path string) (bool, error) {
	if method == "" {
		method = "*"
	}
	if withMethod, ok := s.repo.(interface {
		RouteAllowedWithMethod(ctx context.Context, profile string, method string, path string) (bool, error)
	}); ok {
		return withMethod.RouteAllowedWithMethod(ctx, profile, method, path)
	}
	return s.repo.RouteAllowed(ctx, profile, path)
}

func (s *Service) IsProfileActive(ctx context.Context, profile string) (bool, error) {
	return s.repo.ProfileHasRequiredPolicies(ctx, profile)
}

func (s *Service) ListarPoliticas(ctx context.Context, criterios pagination.CriteriosPagina[policy.Politica]) (pagination.CriteriosPagina[policy.Politica], error) {
	rows, count, err := s.repo.ListarPoliticas(ctx, criterios)
	if err != nil {
		return pagination.CriteriosPagina[policy.Politica]{}, fmt.Errorf("error al listar políticas: %w", err)
	}
	criterios.Data = rows
	criterios.Count = count
	return criterios, nil
}

func (s *Service) ObtenerPolitica(ctx context.Context, nombre string) (policy.Politica, error) {
	return s.repo.ObtenerPolitica(ctx, nombre)
}

func (s *Service) CrearPolitica(ctx context.Context, p policy.Politica) error {
	if err := validatePolicy(p); err != nil {
		return err
	}
	return s.repo.CrearPolitica(ctx, p)
}

func (s *Service) ActualizarPolitica(ctx context.Context, p policy.Politica, nombreOriginal string) error {
	if strings.TrimSpace(nombreOriginal) == "" {
		return fmt.Errorf("nombre original requerido")
	}
	if err := validatePolicy(p); err != nil {
		return err
	}
	return s.repo.ActualizarPolitica(ctx, p, nombreOriginal)
}

func (s *Service) EliminarPolitica(ctx context.Context, nombre string) error {
	if strings.TrimSpace(nombre) == "" {
		return fmt.Errorf("nombre requerido")
	}
	return s.repo.EliminarPolitica(ctx, nombre)
}

func (s *Service) CatalogosPolitica(ctx context.Context) ([]string, []string, []string, error) {
	perfiles, err := s.repo.ListarPerfiles(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error listando perfiles: %w", err)
	}
	recursos, err := s.repo.ListarRecursos(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error listando recursos: %w", err)
	}
	acciones, err := s.repo.ListarAcciones(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error listando acciones: %w", err)
	}
	return perfiles, recursos, acciones, nil
}

func validatePolicy(p policy.Politica) error {
	if strings.TrimSpace(p.Nombre) == "" {
		return fmt.Errorf("nombre requerido")
	}
	if strings.TrimSpace(p.Perfil) == "" || strings.TrimSpace(p.Recurso) == "" || strings.TrimSpace(p.Accion) == "" {
		return fmt.Errorf("perfil, recurso y acción son obligatorios")
	}
	if strings.TrimSpace(p.PathPrefix) == "" {
		return fmt.Errorf("path_prefix requerido")
	}
	if !strings.HasPrefix(strings.TrimSpace(p.PathPrefix), "/") {
		return fmt.Errorf("path_prefix debe iniciar con /")
	}
	return nil
}
