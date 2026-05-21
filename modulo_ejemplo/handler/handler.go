package handler

import (
	"PlantillaInstitucional/modulo_ejemplo"
	"PlantillaInstitucional/modulo_ejemplo/view"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type handler struct {
	service modulo_ejemplo.Service
}

func NewHandler(service modulo_ejemplo.Service) modulo_ejemplo.Handler {
	return &handler{service: service}
}

func (h *handler) RegisterRoutes(mux http.Handler) {
	r := mux.(chi.Router)
	r.Get("/examples", h.HandleList)
	r.Get("/examples/tabla", h.HandleSearchTable)
	r.Post("/examples", h.HandleCreate)
}

func (h *handler) HandleList(w http.ResponseWriter, r *http.Request) {
	examples, err := h.service.ListExamples(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.List(examples).Render(r.Context(), w)
}

func (h *handler) HandleSearchTable(w http.ResponseWriter, r *http.Request) {
	examples, err := h.service.ListExamples(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// En un escenario real, aquí filtraríamos y paginaríamos los examples
	// usando r.FormValue("id"), r.FormValue("name"), etc.
	// Por ahora devolvemos el componente de tabla actualizado.
	// Nota: DataTableV2 en view.templ ya tiene las props, aquí solo renderizamos el body
	// o el componente completo si queremos actualizar filtros/paginacion.
	// Para este ejemplo, renderizamos la vista completa de nuevo o una versión parcial.
	
	// Usamos el helper de la vista para renderizar solo el componente de tabla
	// (Tendríamos que extraer DataTableV2 a un componente exportado en view si fuera complejo)
	// Para simplicidad en este ejemplo, dejamos que HandleList maneje el render inicial
	// y aquí devolvemos lo que pide el hx-target.
	
	view.List(examples).Render(r.Context(), w)
}

func (h *handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	err := h.service.CreateExample(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// For HTMX requests, we might want to return just the new row
	// Or we can just list all again. Let's return all for simplicity or the row if htmx.
	if r.Header.Get("HX-Request") == "true" {
		// Ideally we fetch the newly created example to get its ID, 
		// but since CreateExample doesn't return it currently (it could), 
		// we'll just list again or return a simple row if we had the ID.
		// Let's assume we want to see the update.
		examples, _ := h.service.ListExamples(r.Context())
		if len(examples) > 0 {
			newest := examples[len(examples)-1]
			view.ExampleRow(newest).Render(r.Context(), w)
		}
		return
	}

	http.Redirect(w, r, "/examples", http.StatusSeeOther)
}
