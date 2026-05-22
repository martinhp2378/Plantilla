package handler

import (
	"context"
	"fmt"
	"net/http"

	"PlantillaInstitucional/configuracion/usuario"
	"PlantillaInstitucional/institucion"
	"PlantillaInstitucional/pagination"
	shared "PlantillaInstitucional/shared/errores"
	sharedEvents "PlantillaInstitucional/shared/events"
	"PlantillaInstitucional/shared/pagina"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/logs"
)

// PersonaServiceProvider define el contrato del servicio de personas necesario para usuarios.
type PersonaServiceProvider interface {
	ObtenerTodasCorto(ctx context.Context) ([]institucion.Persona, error)
}

// UsuarioServiceProvider define el contrato del servicio de usuarios.
type UsuarioServiceProvider interface {
	ConsultarUsuarios(ctx context.Context, criterios pagination.CriteriosPagina[usuario.Usuario]) (pagination.CriteriosPagina[usuario.Usuario], error)
	ConsultarUsuarioPorID(ctx context.Context, uuid_persona uuid.UUID) (usuario.Usuario, error)
	ObtenerUsuario(ctx context.Context, uuid_persona uuid.UUID) (usuario.Usuario, error)
	ListarUsuarios(ctx context.Context, criterios pagination.CriteriosPagina[usuario.Usuario]) (pagination.CriteriosPagina[usuario.Usuario], error)
	CrearUsuario(ctx context.Context, u usuario.Usuario, rawPassword string) error
	ActualizarUsuario(ctx context.Context, u usuario.Usuario) error
	EliminarUsuario(ctx context.Context, uuid_persona uuid.UUID) error
	ReportePdf(ctx context.Context) (string, error)
	ReporteExcel(ctx context.Context) (string, error)
}

// UsuarioHandler es la implementación del handler de usuarios.
type UsuarioHandler struct {
	validate        *validator.Validate
	sqlErrorHandler *shared.SQLErrorHandler
	service         UsuarioServiceProvider
	personaService  PersonaServiceProvider
	eventDispatcher *sharedEvents.Dispatcher
	logger          *logs.Logger
}

func NewUsuarioHandler(svc UsuarioServiceProvider, personaService PersonaServiceProvider, eventDispatcher *sharedEvents.Dispatcher, logger *logs.Logger, validate *validator.Validate) *UsuarioHandler {
	return &UsuarioHandler{
		service:         svc,
		personaService:  personaService,
		eventDispatcher: eventDispatcher,
		logger:          logger,
		validate:        validate,
		sqlErrorHandler: shared.NewSQLErrorHandler(logger),
	}
}

// --- Helpers de Parseo de Requests (Privados) ---

func (h *UsuarioHandler) RegisterRoutes(r chi.Router) {
	rutas := pagina.GlobalRutas.Usuarios
	r.Route(rutas.Prefix, func(r chi.Router) {
		r.Get("/", h.Index)
		r.Get("/tabla", h.SearchTable)
		r.Get("/modales/nuevo", h.ModalCreate)
		r.Post("/", h.Create)
		r.Get("/modales/editar/{id}", h.ModalUpdate)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)

		r.Route("/reportes", func(r chi.Router) {
			r.Get("/pdf", h.DownloadPDF)
			r.Get("/excel", h.DownloadExcel)
			r.Get("/notificacion-exito", h.SuccessFileMessage)
		})
	})
}

func (h *UsuarioHandler) parseFormToUsuario(r *http.Request) (usuario.Usuario, error) {
	if err := r.ParseForm(); err != nil {
		return usuario.Usuario{}, fmt.Errorf("error parseando form: %w", err)
	}

	u := usuario.Usuario{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
		Perfil:   r.FormValue("perfil"),
	}

	if personaID := r.FormValue("uuid_persona"); personaID != "" {
		uid, err := uuid.FromString(personaID)
		if err == nil {
			u.UuidPersona = uid
		}
	}

	return u, nil
}

func (h *UsuarioHandler) parseIDParam(r *http.Request) (uuid.UUID, error) {
	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		return uuid.Nil, fmt.Errorf("id requerido")
	}
	// Validamos formato brevemente
	id, err := uuid.FromString(idParam)
	if err != nil {
		return uuid.Nil, fmt.Errorf("formato de id inválido")
	}

	return id, nil
}
