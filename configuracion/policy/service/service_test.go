package service

import (
	"PlantillaInstitucional/configuracion/policy"
	"PlantillaInstitucional/pagination"
	"context"
	"encoding/json"
	"errors"
	"testing"
)

type fakePolicyRepo struct {
	allowResult  bool
	allowErr     error
	activeResult bool
	activeErr    error
	createErr    error
	updateErr    error
	createCalled bool
	updateCalled bool
	lastPolicy   policy.Politica
	lastName     string
	lastMethod   string
	lastPath     string
}

func (f *fakePolicyRepo) RouteAllowed(ctx context.Context, profile string, path string) (bool, error) {
	f.lastPath = path
	return f.allowResult, f.allowErr
}

func (f *fakePolicyRepo) RouteAllowedWithMethod(ctx context.Context, profile string, method string, path string) (bool, error) {
	f.lastMethod = method
	f.lastPath = path
	return f.allowResult, f.allowErr
}

func (f *fakePolicyRepo) ProfileHasRequiredPolicies(ctx context.Context, profile string) (bool, error) {
	return f.activeResult, f.activeErr
}

func (f *fakePolicyRepo) ListarPoliticas(ctx context.Context, criterios pagination.CriteriosPagina[policy.Politica]) ([]policy.Politica, int, error) {
	return nil, 0, nil
}

func (f *fakePolicyRepo) ObtenerPolitica(ctx context.Context, nombre string) (policy.Politica, error) {
	return policy.Politica{}, nil
}

func (f *fakePolicyRepo) CrearPolitica(ctx context.Context, p policy.Politica) error {
	f.createCalled = true
	f.lastPolicy = p
	return f.createErr
}

func (f *fakePolicyRepo) ActualizarPolitica(ctx context.Context, p policy.Politica, nombreOriginal string) error {
	f.updateCalled = true
	f.lastPolicy = p
	f.lastName = nombreOriginal
	return f.updateErr
}

func (f *fakePolicyRepo) EliminarPolitica(ctx context.Context, nombre string) error {
	return nil
}

func (f *fakePolicyRepo) ListarPerfiles(ctx context.Context) ([]string, error) {
	return []string{"admin"}, nil
}

func (f *fakePolicyRepo) ListarRecursos(ctx context.Context) ([]string, error) {
	return []string{"dashboard"}, nil
}

func (f *fakePolicyRepo) ListarAcciones(ctx context.Context) ([]string, error) {
	return []string{"consultar"}, nil
}

func TestCrearPoliticaNormalizesRules(t *testing.T) {
	repo := &fakePolicyRepo{}
	svc := NewService(repo)

	p := policy.Politica{
		Nombre:      "test-policy",
		Perfil:      "admin",
		Recurso:     "dashboard",
		Accion:      "consultar",
		PathPrefix:  "/dashboard",
		Reglas:      `{"path_prefix": "/dashboard", "routes": [{"method": "GET", "path": "/dashboard", "accion": "consultar"}]}`,
	}

	err := svc.CrearPolitica(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.createCalled {
		t.Fatal("expected repo.CrearPolitica to be called")
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(repo.lastPolicy.Reglas), &parsed); err != nil {
		t.Fatalf("stored rules must be valid JSON: %v", err)
	}

	if parsed["path_prefix"] != "/dashboard" {
		t.Errorf("expected path_prefix to be /dashboard, got %v", parsed["path_prefix"])
	}
}

func TestCrearPoliticaAutoGeneratesDefaultRulesWhenEmpty(t *testing.T) {
	repo := &fakePolicyRepo{}
	svc := NewService(repo)

	p := policy.Politica{
		Nombre:     "test-default",
		Perfil:     "admin",
		Recurso:    "crud",
		Accion:     "consultar",
		PathPrefix: "/crud",
		Reglas:     "",
	}

	err := svc.CrearPolitica(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(repo.lastPolicy.Reglas), &parsed); err != nil {
		t.Fatalf("stored rules must be valid JSON: %v", err)
	}

	routes, ok := parsed["routes"].([]any)
	if !ok || len(routes) == 0 {
		t.Fatal("expected routes array to be auto-generated")
	}
}

func TestCrearPoliticaRejectsInvalidJSON(t *testing.T) {
	repo := &fakePolicyRepo{}
	svc := NewService(repo)

	p := policy.Politica{
		Nombre:     "bad-json",
		Perfil:     "admin",
		Recurso:    "crud",
		Accion:     "consultar",
		PathPrefix: "/crud",
		Reglas:     "{not-valid-json}",
	}

	err := svc.CrearPolitica(context.Background(), p)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestIsRouteAllowedWithMethod(t *testing.T) {
	repo := &fakePolicyRepo{allowResult: true}
	svc := NewService(repo)

	allowed, err := svc.IsRouteAllowedWithMethod(context.Background(), "admin", "GET", "/dashboard")
	if err != nil || !allowed {
		t.Fatalf("expected allowed=true, err=nil; got %v, %v", allowed, err)
	}

	if repo.lastMethod != "GET" || repo.lastPath != "/dashboard" {
		t.Errorf("expected GET /dashboard, got %s %s", repo.lastMethod, repo.lastPath)
	}
}

func TestIsProfileActive(t *testing.T) {
	repo := &fakePolicyRepo{activeResult: true}
	svc := NewService(repo)

	active, err := svc.IsProfileActive(context.Background(), "admin")
	if err != nil || !active {
		t.Fatalf("expected active=true, err=nil; got %v, %v", active, err)
	}
}

func TestRepoErrorPropagates(t *testing.T) {
	expectedErr := errors.New("db error")
	repo := &fakePolicyRepo{createErr: expectedErr}
	svc := NewService(repo)

	p := policy.Politica{
		Nombre:     "policy-error",
		Perfil:     "admin",
		Recurso:    "crud",
		Accion:     "consultar",
		PathPrefix: "/crud",
	}

	err := svc.CrearPolitica(context.Background(), p)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}
