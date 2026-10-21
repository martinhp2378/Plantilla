package service_test

import (
	"PlantillaInstitucional/crud"
	"PlantillaInstitucional/crud/service"
	"PlantillaInstitucional/pagination"
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/logs"
)

type mockCrudRepo struct {
	items map[uuid.UUID]crud.Crud
}

func newMockCrudRepo() *mockCrudRepo {
	return &mockCrudRepo{items: make(map[uuid.UUID]crud.Crud)}
}

func (m *mockCrudRepo) ConsultarCrud(ctx context.Context, criterios pagination.CriteriosPagina[crud.Crud]) ([]*crud.Crud, int, error) {
	var res []*crud.Crud
	for _, item := range m.items {
		val := item
		res = append(res, &val)
	}
	return res, len(res), nil
}

func (m *mockCrudRepo) ConsultarCrudPorID(ctx context.Context, id uuid.UUID) (*crud.Crud, error) {
	if item, ok := m.items[id]; ok {
		return &item, nil
	}
	return nil, nil
}

func (m *mockCrudRepo) CrearCrud(ctx context.Context, c crud.Crud) error {
	m.items[c.Id] = c
	return nil
}

func (m *mockCrudRepo) ActualizarCrud(ctx context.Context, c crud.Crud) error {
	m.items[c.Id] = c
	return nil
}

func (m *mockCrudRepo) EliminarCrud(ctx context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockCrudRepo) ObtenerTodos(ctx context.Context) ([]crud.Crud, error) {
	var res []crud.Crud
	for _, item := range m.items {
		res = append(res, item)
	}
	return res, nil
}

func TestCrudServiceCrear(t *testing.T) {
	validate := validator.New()
	logger := logs.NuevoLogger(&logs.Config{Nivel: "ERROR"})
	repo := newMockCrudRepo()
	svc := service.NewCrudService(repo, validate, logger, nil)

	ctx := context.Background()

	t.Run("Fails on invalid email", func(t *testing.T) {
		invalid := crud.Crud{
			TextoCorto: "Prueba",
			Correo:     "not-an-email",
			Fecha:      "2026-10-20",
			Hora:       "12:00",
			Seleccion:  "opcion_a",
		}

		err := svc.CrearCrud(ctx, invalid)
		if err == nil {
			t.Fatal("expected validation error for invalid email, got nil")
		}
	})

	t.Run("Fails on missing required fields", func(t *testing.T) {
		invalid := crud.Crud{
			TextoCorto: "",
			Correo:     "valido@issemym.gob.mx",
		}

		err := svc.CrearCrud(ctx, invalid)
		if err == nil {
			t.Fatal("expected validation error for missing texto_corto, got nil")
		}
	})

	t.Run("Succeeds on valid input and assigns UUID", func(t *testing.T) {
		valid := crud.Crud{
			TextoCorto: "Registro Válido",
			TextoLargo: "Detalles",
			Entero:     10,
			Decimal:    99.50,
			Booleano:   true,
			Fecha:      "2026-10-20",
			Hora:       "10:30",
			FechaHora:  "2026-10-20T10:30",
			Seleccion:  "opcion_a",
			Correo:     "admin@issemym.gob.mx",
			Telefono:   "7221234567",
			URL:        "https://issemym.gob.mx",
			Color:      "#691C32",
		}

		err := svc.CrearCrud(ctx, valid)
		if err != nil {
			t.Fatalf("unexpected error creating crud: %v", err)
		}

		if len(repo.items) != 1 {
			t.Fatalf("expected 1 item in repo, got %d", len(repo.items))
		}

		for id, saved := range repo.items {
			if id == uuid.Nil {
				t.Fatal("expected non-nil UUID generated")
			}
			if saved.TextoCorto != "Registro Válido" {
				t.Errorf("expected TextoCorto 'Registro Válido', got '%s'", saved.TextoCorto)
			}
		}
	})
}
