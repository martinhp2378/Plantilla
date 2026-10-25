package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"PlantillaInstitucional/configuracion/usuario"
	usuarioEvents "PlantillaInstitucional/configuracion/usuario/events"
	"PlantillaInstitucional/pagination"
	sharedEvents "PlantillaInstitucional/shared/events"
	sharedComp "PlantillaInstitucional/shared/views/components"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
)

type failingActivationListener struct{ event sharedEvents.Event }

func (l *failingActivationListener) Handle(_ context.Context, event sharedEvents.Event) error {
	l.event = event
	return errors.New("notification unavailable")
}

type createUserService struct{ deleted uuid.UUID }

func (s *createUserService) ConsultarUsuarios(context.Context, pagination.CriteriosPagina[usuario.Usuario]) (pagination.CriteriosPagina[usuario.Usuario], error) {
	return pagination.CriteriosPagina[usuario.Usuario]{}, nil
}
func (s *createUserService) ConsultarUsuarioPorID(context.Context, uuid.UUID) (usuario.Usuario, error) {
	return usuario.Usuario{}, nil
}
func (s *createUserService) ObtenerUsuario(context.Context, uuid.UUID) (usuario.Usuario, error) {
	return usuario.Usuario{}, nil
}
func (s *createUserService) ListarUsuarios(context.Context, pagination.CriteriosPagina[usuario.Usuario]) (pagination.CriteriosPagina[usuario.Usuario], error) {
	return pagination.CriteriosPagina[usuario.Usuario]{}, nil
}
func (s *createUserService) CrearUsuario(context.Context, usuario.Usuario, string) error { return nil }
func (s *createUserService) ActualizarUsuario(context.Context, usuario.Usuario) error    { return nil }
func (s *createUserService) EliminarUsuario(_ context.Context, id uuid.UUID) error {
	s.deleted = id
	return nil
}
func (s *createUserService) ReportePdf(context.Context) (string, error)   { return "", nil }
func (s *createUserService) ReporteExcel(context.Context) (string, error) { return "", nil }

type createPersonaService struct{}

func (createPersonaService) ObtenerOpcionesPersona(context.Context) ([]sharedComp.SelectOption, error) {
	return nil, nil
}

func TestCreateDoesNotReportSuccessWhenActivationNotificationFails(t *testing.T) {
	userID := uuid.Must(uuid.NewV4())
	service := &createUserService{}
	dispatcher := sharedEvents.NewDispatcher()
	listener := &failingActivationListener{}
	dispatcher.Register(usuarioEvents.SendUserActivationEvent{}.Name(), listener)
	h := NewUsuarioHandler(service, createPersonaService{}, dispatcher, nil, validator.New())

	form := url.Values{"email": {"person@example.test"}, "perfil": {"operador"}, "uuid_persona": {userID.String()}}
	req := httptest.NewRequest(http.MethodPost, "/configuracion/usuarios/", nil)
	req.PostForm = form
	req.Form = form
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if service.deleted != userID {
		t.Fatalf("created user must be removed after delivery failure; deleted = %s, want %s", service.deleted, userID)
	}
	event, ok := listener.event.(usuarioEvents.SendUserActivationEvent)
	if !ok || event.ActivationURL == "" {
		t.Fatalf("event = %#v, want one-time activation event", listener.event)
	}
	if rec.Code == http.StatusOK && rec.Body.String() == "" {
		t.Fatal("notification failure must render an explicit failure state")
	}
}
