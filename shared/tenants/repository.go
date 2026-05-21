package tenants

import "context"

type Tenant struct {
	ID     string
	Schema string
	Host   string
}

type Repository interface {
	GetByHost(host string) (*Tenant, error)
	ListarTenantsActivos(ctx context.Context) (map[string]string, error)
}

type StaticRepository struct {
	tenants []Tenant
}

func NewStaticRepository() *StaticRepository {
	return &StaticRepository{
		tenants: []Tenant{
			{ID: "tenant1", Schema: "tenant1", Host: "tenant1.localhost"},
			{ID: "tenant2", Schema: "tenant2", Host: "tenant2.localhost"},
		},
	}
}

func (r *StaticRepository) GetByHost(host string) (*Tenant, error) {
	for _, t := range r.tenants {
		if t.Host == host {
			return &t, nil
		}
	}
	return nil, nil // Or return a specific error if not found
}

func (r *StaticRepository) ListarTenantsActivos(ctx context.Context) (map[string]string, error) {
	m := make(map[string]string)
	for _, t := range r.tenants {
		m[t.Host] = t.Schema
	}
	return m, nil
}
