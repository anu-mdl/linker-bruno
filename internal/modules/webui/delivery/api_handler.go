package delivery

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/anu-mdl/linker-bruno/internal/modules/webui/dto"
	"github.com/anu-mdl/linker-bruno/internal/modules/webui/service"
	"github.com/go-chi/chi/v5"
)

// APIHandler handles CRUD operations for requests via HTMX endpoints
type APIHandler struct {
	service   *service.RequestService
	templates *template.Template
	baseDir   string
}

// NewAPIHandler creates a new APIHandler
func NewAPIHandler(service *service.RequestService, templates *template.Template, baseDir string) *APIHandler {
	return &APIHandler{
		service:   service,
		templates: templates,
		baseDir:   baseDir,
	}
}

// RegisterRoutes registers all API routes
func (h *APIHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/requests", h.HandleListRequests)
	r.Get("/api/requests/{id}", h.HandleGetRequest)
	r.Post("/api/requests", h.HandleCreateRequest)
	r.Put("/api/requests/{id}", h.HandleUpdateRequest)
	r.Delete("/api/requests/{id}", h.HandleDeleteRequest)
}

// HandleListRequests returns all requests as a tree structure
func (h *APIHandler) HandleListRequests(w http.ResponseWriter, r *http.Request) {
	// Build folder tree
	tree, err := h.service.BuildRequestTree(h.baseDir)
	if err != nil {
		log.Printf("Error building tree: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Render sidebar template
	err = h.templates.ExecuteTemplate(w, "sidebar.html", map[string]interface{}{
		"Tree": tree,
	})
	if err != nil {
		log.Printf("Error rendering sidebar: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// HandleGetRequest returns a specific request for editing
func (h *APIHandler) HandleGetRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Load the request
	req, err := h.service.GetRequestByID(id)
	if err != nil {
		log.Printf("Error loading request: %v", err)
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}

	// Render editor template
	err = h.templates.ExecuteTemplate(w, "editor.html", map[string]interface{}{
		"Request": req,
		"ID":      id,
	})
	if err != nil {
		log.Printf("Error rendering editor: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// HandleCreateRequest creates a new request
func (h *APIHandler) HandleCreateRequest(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Create input from form
	input := &dto.CreateRequestInput{
		Name:        r.FormValue("name"),
		Method:      r.FormValue("method"),
		URL:         r.FormValue("url"),
		Description: r.FormValue("example_description"),
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
		Body:        r.FormValue("request_body"),
		ResponseHeaders: map[string]string{
			"content-type": "application/json",
		},
		ResponseBody: r.FormValue("response_body"),
	}

	// Set default status if not provided
	input.ResponseStatus.Code = 200
	input.ResponseStatus.Text = "OK"

	// Create request
	if err := h.service.CreateRequest(h.baseDir, input); err != nil {
		log.Printf("Error creating request: %v", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Return success and trigger sidebar reload
	w.Header().Set("HX-Trigger", "requestsChanged")
	w.WriteHeader(http.StatusCreated)
}

// HandleUpdateRequest updates an existing request
func (h *APIHandler) HandleUpdateRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Load existing request to preserve data
	existingReq, err := h.service.GetRequestByID(id)
	if err != nil {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}

	// Create update input from form
	input := &dto.UpdateRequestInput{
		Name:        r.FormValue("name"),
		Method:      r.FormValue("method"),
		URL:         r.FormValue("url"),
		Description: r.FormValue("example_description"),
		Body:        r.FormValue("request_body"),
		ResponseBody: r.FormValue("response_body"),
	}

	// Parse request headers from form arrays
	reqHeaderKeys := r.Form["req_header_key[]"]
	reqHeaderValues := r.Form["req_header_value[]"]
	input.Headers = make(map[string]string)
	for i := 0; i < len(reqHeaderKeys) && i < len(reqHeaderValues); i++ {
		if reqHeaderKeys[i] != "" {
			input.Headers[reqHeaderKeys[i]] = reqHeaderValues[i]
		}
	}

	// Parse query params from form arrays
	paramKeys := r.Form["param_key[]"]
	paramValues := r.Form["param_value[]"]
	input.QueryParams = make(map[string]string)
	for i := 0; i < len(paramKeys) && i < len(paramValues); i++ {
		if paramKeys[i] != "" {
			input.QueryParams[paramKeys[i]] = paramValues[i]
		}
	}

	// Parse response headers from form arrays
	respHeaderKeys := r.Form["resp_header_key[]"]
	respHeaderValues := r.Form["resp_header_value[]"]
	input.ResponseHeaders = make(map[string]string)
	for i := 0; i < len(respHeaderKeys) && i < len(respHeaderValues); i++ {
		if respHeaderKeys[i] != "" {
			input.ResponseHeaders[respHeaderKeys[i]] = respHeaderValues[i]
		}
	}
	// Preserve existing headers if none provided
	if len(input.ResponseHeaders) == 0 {
		input.ResponseHeaders = existingReq.ResponseHeaders
	}

	// Parse status code and text
	statusCode := existingReq.ResponseStatus.Code
	statusText := existingReq.ResponseStatus.Text

	if statusCodeStr := r.FormValue("status_code"); statusCodeStr != "" {
		if code, err := strconv.Atoi(statusCodeStr); err == nil {
			statusCode = code
		}
	}
	if st := r.FormValue("status_text"); st != "" {
		statusText = st
	}

	input.ResponseStatus.Code = statusCode
	input.ResponseStatus.Text = statusText

	// Update request
	if err := h.service.UpdateRequest(id, input); err != nil {
		log.Printf("Error updating request: %v", err)
		http.Error(w, "Failed to update request", http.StatusInternalServerError)
		return
	}

	// Return success and trigger sidebar reload
	w.Header().Set("HX-Trigger", "requestsChanged")
	w.WriteHeader(http.StatusOK)
}

// HandleDeleteRequest deletes a request
func (h *APIHandler) HandleDeleteRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteRequest(id); err != nil {
		log.Printf("Error deleting request: %v", err)
		http.Error(w, "Failed to delete request", http.StatusInternalServerError)
		return
	}

	// Return success and trigger sidebar reload
	w.Header().Set("HX-Trigger", "requestsChanged")
	w.WriteHeader(http.StatusOK)
}
