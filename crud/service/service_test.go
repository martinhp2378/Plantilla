package service_test

import (
	"PlantillaInstitucional/crud"
	"PlantillaInstitucional/crud/service"
	"PlantillaInstitucional/pagination"
	"context"
	"strings"
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
	res := make([]*crud.Crud, 0, len(m.items))
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

func (m *mockCrudRepo) CrearCrudMasivo(ctx context.Context, items []crud.Crud) error {
	for _, c := range items {
		m.items[c.Id] = c
	}
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
	res := make([]crud.Crud, 0, len(m.items))
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

	t.Run("GenerarPlantillaCSV produces valid header and sample row", func(t *testing.T) {
		bytesCSV, err := svc.GenerarPlantillaCSV()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(bytesCSV) == 0 {
			t.Fatal("expected non-empty template CSV")
		}
	})

	t.Run("CargarMasivo with valid CSV persists records", func(t *testing.T) {
		csvContent := "texto_corto,texto_largo,entero,decimal,booleano,fecha,hora,seleccion,correo,telefono,url,color\n" +
			"Masivo 1,Detalle 1,15,199.99,Sí,2026-10-21,12:00,opcion_a,masivo1@issemym.gob.mx,7221111111,https://issemym.gob.mx,#691C32\n" +
			"Masivo 2,Detalle 2,30,450.50,No,2026-10-22,14:30,opcion_b,masivo2@issemym.gob.mx,7222222222,https://issemym.gob.mx,#10312B\n"

		res, err := svc.CargarMasivo(ctx, strings.NewReader(csvContent))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.EsExitoso() {
			t.Fatalf("expected successful bulk upload, got errors: %+v", res.Errores)
		}
		if res.TotalValidos != 2 {
			t.Errorf("expected 2 valid items, got %d", res.TotalValidos)
		}
	})

	t.Run("CargarMasivo with validation errors returns error list without persisting", func(t *testing.T) {
		csvContent := "texto_corto,texto_largo,entero,decimal,booleano,fecha,hora,seleccion,correo,telefono,url,color\n" +
			",Detalle,abc,invalido,no_bool,fecha_mal,hora_mal,opcion_x,email_mal,7221111111,url_mal,color_mal\n"

		res, err := svc.CargarMasivo(ctx, strings.NewReader(csvContent))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.EsExitoso() {
			t.Fatal("expected validation errors, got success")
		}
		if len(res.Errores) == 0 {
			t.Fatal("expected detailed row errors, got 0")
		}
	})
}
