package app

import (
	"PlantillaInstitucional/shared/entorno"
	"context"
	"testing"
)

func TestNewApp(t *testing.T) {
	// We use a dummy env. NewPool will fail to connect but we can verify the error
	// or mock if we had a mocking library. For now, let's just ensure the logic flows.
	env := &entorno.Env{
		PqWriteHost: "localhost",
		PqWritePort: "5432",
		PqReadHost:  "localhost",
		PqReadPort:  "5432",
	}

	ctx := context.Background()
	a, err := NewApp(ctx, env, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("expected nil error for simple constructor, got %v", err)
	}
	if err := a.TenantPolicy.Validate(); err != nil {
		t.Errorf("default tenant policy must be valid, got %v", err)
	}
}
