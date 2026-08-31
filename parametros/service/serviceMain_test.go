package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/parametros"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/logs"
)

type stubParametrosRepo struct {
	actualizarFn func(context.Context, parametros.Parametro) error
}

func (s *stubParametrosRepo) ConsultarParametros(context.Context, pagination.CriteriosPagina[parametros.Parametro]) ([]parametros.Parametro, int, error) {
	return nil, 0, nil
}
func (s *stubParametrosRepo) ConsultarParametroPorID(context.Context, uuid.UUID) (parametros.Parametro, error) {
	return parametros.Parametro{}, nil
}
func (s *stubParametrosRepo) ActualizarParametro(ctx context.Context, p parametros.Parametro) error {
	if s.actualizarFn != nil {
		return s.actualizarFn(ctx, p)
	}
	return nil
}
func (s *stubParametrosRepo) EliminarParametro(context.Context, uuid.UUID) error { return nil }
func (s *stubParametrosRepo) CrearParametro(context.Context, parametros.Parametro) error {
	return nil
}
func (s *stubParametrosRepo) CargarParametros() error             { return nil }
func (s *stubParametrosRepo) GetJSONParamContext(context.Context, string) (string, error) {
	return "", nil
}
func (s *stubParametrosRepo) MonitorParametros(<-chan string) {}

func TestActualizarParametroPropagaErrorRepo(t *testing.T) {
	repoErr := errors.New("db caída")
	svc := NewService(logs.NuevoLogger(&logs.Config{Ambiente: logs.AmbienteDesarrollo, Servicio: "parametros-service-test"}), &stubParametrosRepo{
		actualizarFn: func(context.Context, parametros.Parametro) error { return repoErr },
	}, validator.New(), nil)

	err := svc.ActualizarParametro(context.Background(), parametros.Parametro{Clave: "CFG_KEY", Valor: "1", TipoDato: "INT"})
	if err == nil {
		t.Fatal("error = nil, want propagated repo error")
	}
	if !strings.Contains(err.Error(), "error actualizando Parametro") {
		t.Fatalf("error = %q, want wraps update context", err.Error())
	}
	if !strings.Contains(err.Error(), "db caída") {
		t.Fatalf("error = %q, want wrapped repo error", err.Error())
	}
}
