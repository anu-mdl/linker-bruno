package delivery

import (
	"html/template"
	"log"
	"net/http"

	"github.com/anu-mdl/linker-bruno/internal/modules/webui/service"
	"github.com/go-chi/chi/v5"
)

// UIHandler handles web UI page rendering and static file serving
type UIHandler struct {
	templates *template.Template
	service   *service.RequestService
}

// NewUIHandler creates a new UIHandler
func NewUIHandler(templates *template.Template, service *service.RequestService) *UIHandler {
	return &UIHandler{
		templates: templates,
		service:   service,
	}
}

// RegisterRoutes registers all UI routes
func (h *UIHandler) RegisterRoutes(r chi.Router) {
	// Serve static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("server/static"))))

	// Main UI page
	r.Get("/", h.HandleIndex)
}

// HandleIndex serves the main UI page
func (h *UIHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	err := h.templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
