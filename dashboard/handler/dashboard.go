package handler

import (
	"encoding/json"
	"net/http"

	"PlantillaInstitucional/dashboard"
	"PlantillaInstitucional/dashboard/view/layouts"

	"github.com/go-chi/chi/v5"
	"github.com/issemym-repos/logs"
)

type DashboardHandler struct {
	logger *logs.Logger
}

func NewDashboardHandler(logger *logs.Logger) *DashboardHandler {
	return &DashboardHandler{
		logger: logger,
	}
}

func (h *DashboardHandler) RegisterRoutes(r chi.Router) {
	r.Get(dashboard.Rutas.Prefix, h.Index)
	r.Route(dashboard.Rutas.Prefix, func(r chi.Router) {
		r.Get("/", h.Index)
	})
}

func (h *DashboardHandler) Index(w http.ResponseWriter, r *http.Request) {
	lineData := []map[string]any{
		{"label": "Ene", "value": 185},
		{"label": "Feb", "value": 240},
		{"label": "Mar", "value": 310},
		{"label": "Abr", "value": 295},
		{"label": "May", "value": 420},
		{"label": "Jun", "value": 538},
	}
	lineJSON, err := json.Marshal(lineData)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "error al serializar line data", err)
		lineJSON = []byte("[]")
	}

	pieData := []map[string]any{
		{"name": "Completados", "value": 650},
		{"name": "En Proceso", "value": 340},
		{"name": "En Revisión", "value": 180},
		{"name": "Cancelados", "value": 78},
	}
	pieJSON, err := json.Marshal(pieData)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "error al serializar pie data", err)
		pieJSON = []byte("[]")
	}

	data := dashboard.DashboardData{
		TotalRegistros:  "1,248",
		UsuariosActivos: "48",
		Eficiencia:      "94.8%",
		Tramites:        "3,420",
		LineChartJSON:   string(lineJSON),
		PieChartJSON:    string(pieJSON),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := layouts.DashboardV2(data).Render(r.Context(), w); err != nil {
		h.logger.ErrorContext(r.Context(), "error al renderizar dashboard", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}
