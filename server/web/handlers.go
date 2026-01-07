package web

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/anu-mdl/linker-bruno/server/loader"
	"github.com/anu-mdl/linker-bruno/server/parser"
	"github.com/go-chi/chi/v5"
)

// UIServer handles web UI routes
type UIServer struct {
	baseDir   string
	templates *template.Template
}

// NewUIServer creates a new UI server
func NewUIServer(baseDir string) (*UIServer, error) {
	// Create custom template functions
	funcMap := template.FuncMap{
		"displayURL": displayURL,
		"displayParam": displayParam,
	}

	// Load templates with custom functions
	tmpl, err := template.New("").Funcs(funcMap).ParseGlob(filepath.Join("server/templates", "*.html"))
	if err != nil {
		return nil, err
	}

	return &UIServer{
		baseDir:   baseDir,
		templates: tmpl,
	}, nil
}

// displayURL converts URL path parameters from {param} to [param] for display
func displayURL(url string) string {
	// Replace {param} with [param]
	result := url
	for i := 0; i < len(result); i++ {
		if result[i] == '{' {
			result = result[:i] + "[" + result[i+1:]
		} else if result[i] == '}' {
			result = result[:i] + "]" + result[i+1:]
		}
	}

	// Ensure URL starts with /
	if len(result) > 0 && result[0] != '/' {
		result = "/" + result
	}

	return result
}

// displayParam converts a parameter name to display format with brackets
func displayParam(param string) string {
	// If it's already a dynamic param extracted from {param}, wrap it in brackets
	return "[" + param + "]"
}

// RegisterRoutes registers all UI routes
func (s *UIServer) RegisterRoutes(r chi.Router) {
	// Serve static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("server/static"))))

	// Main UI page
	r.Get("/", s.handleIndex)

	// API routes for HTMX
	r.Get("/api/requests", s.handleListRequests)
	r.Get("/api/requests/{id}", s.handleGetRequest)
	r.Post("/api/requests", s.handleCreateRequest)
	r.Put("/api/requests/{id}", s.handleUpdateRequest)
	r.Delete("/api/requests/{id}", s.handleDeleteRequest)
}

// handleIndex serves the main UI page
func (s *UIServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	err := s.templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleListRequests returns all requests as a tree structure
func (s *UIServer) handleListRequests(w http.ResponseWriter, r *http.Request) {
	// Load all requests
	requests, err := loader.LoadAllRequests(s.baseDir)
	if err != nil {
		log.Printf("Error loading requests: %v", err)
		http.Error(w, "Failed to load requests", http.StatusInternalServerError)
		return
	}

	// Build folder tree
	tree := BuildRequestTree(requests)

	// Render sidebar template
	err = s.templates.ExecuteTemplate(w, "sidebar.html", map[string]interface{}{
		"Tree": tree,
	})
	if err != nil {
		log.Printf("Error rendering sidebar: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleGetRequest returns a specific request for editing
func (s *UIServer) handleGetRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Load the request
	req, err := GetRequestByID(s.baseDir, id)
	if err != nil {
		log.Printf("Error loading request: %v", err)
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}

	// Render editor template
	err = s.templates.ExecuteTemplate(w, "editor.html", map[string]interface{}{
		"Request": req,
		"ID":      id,
	})
	if err != nil {
		log.Printf("Error rendering editor: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleCreateRequest creates a new request
func (s *UIServer) handleCreateRequest(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Create request from form
	req := &parser.BrunoRequest{
		Meta: parser.MetaBlock{
			Name: r.FormValue("name"),
			Type: "http",
			Seq:  1,
		},
		Method:      r.FormValue("method"),
		URL:         r.FormValue("url"),
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
		Body:        r.FormValue("request_body"),
	}

	// Create example block
	req.Example = parser.ExampleBlock{
		Name:        req.Meta.Name + " Example",
		Description: r.FormValue("example_description"),
		Request: parser.ExampleRequest{
			URL:    r.FormValue("url"),
			Method: r.FormValue("method"),
			Mode:   "none",
		},
		Response: parser.ExampleResponse{
			Status: parser.ExampleStatus{
				Code: 200,
				Text: "OK",
			},
			Headers: map[string]string{
				"content-type": "application/json",
			},
			Body: parser.ExampleBody{
				Type:    "json",
				Content: r.FormValue("response_body"),
			},
		},
	}

	// Save request
	if err := SaveRequest(s.baseDir, req); err != nil {
		log.Printf("Error saving request: %v", err)
		http.Error(w, "Failed to save request", http.StatusInternalServerError)
		return
	}

	// Return success and trigger sidebar reload
	w.Header().Set("HX-Trigger", "requestsChanged")
	w.WriteHeader(http.StatusCreated)
}

// handleUpdateRequest updates an existing request
func (s *UIServer) handleUpdateRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Load existing request
	req, err := GetRequestByID(s.baseDir, id)
	if err != nil {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}

	// Update request fields
	req.Meta.Name = r.FormValue("name")
	req.Method = r.FormValue("method")
	req.URL = r.FormValue("url")
	req.Body = r.FormValue("request_body")

	// Update request headers from form arrays
	reqHeaderKeys := r.Form["req_header_key[]"]
	reqHeaderValues := r.Form["req_header_value[]"]
	newReqHeaders := make(map[string]string)
	for i := 0; i < len(reqHeaderKeys) && i < len(reqHeaderValues); i++ {
		if reqHeaderKeys[i] != "" {
			newReqHeaders[reqHeaderKeys[i]] = reqHeaderValues[i]
		}
	}
	req.Headers = newReqHeaders

	// Update query params from form arrays
	paramKeys := r.Form["param_key[]"]
	paramValues := r.Form["param_value[]"]
	newParams := make(map[string]string)
	for i := 0; i < len(paramKeys) && i < len(paramValues); i++ {
		if paramKeys[i] != "" {
			newParams[paramKeys[i]] = paramValues[i]
		}
	}
	req.QueryParams = newParams

	// Update example block (preserve name, update description)
	if req.Example.Name == "" {
		req.Example.Name = req.Meta.Name + " Example"
	}
	req.Example.Description = r.FormValue("example_description")
	req.Example.Response.Body.Content = r.FormValue("response_body")

	// Update status code and text
	if statusCode := r.FormValue("status_code"); statusCode != "" {
		if code, err := strconv.Atoi(statusCode); err == nil {
			req.Example.Response.Status.Code = code
		}
	}
	req.Example.Response.Status.Text = r.FormValue("status_text")

	// Update response headers from form arrays
	respHeaderKeys := r.Form["resp_header_key[]"]
	respHeaderValues := r.Form["resp_header_value[]"]
	newRespHeaders := make(map[string]string)
	for i := 0; i < len(respHeaderKeys) && i < len(respHeaderValues); i++ {
		if respHeaderKeys[i] != "" {
			newRespHeaders[respHeaderKeys[i]] = respHeaderValues[i]
		}
	}
	if len(newRespHeaders) > 0 {
		req.Example.Response.Headers = newRespHeaders
	}

	// Save updated request
	if err := UpdateRequest(s.baseDir, id, req); err != nil {
		log.Printf("Error updating request: %v", err)
		http.Error(w, "Failed to update request", http.StatusInternalServerError)
		return
	}

	// Return success and trigger sidebar reload
	w.Header().Set("HX-Trigger", "requestsChanged")
	w.WriteHeader(http.StatusOK)
}

// handleDeleteRequest deletes a request
func (s *UIServer) handleDeleteRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := DeleteRequest(s.baseDir, id); err != nil {
		log.Printf("Error deleting request: %v", err)
		http.Error(w, "Failed to delete request", http.StatusInternalServerError)
		return
	}

	// Return success and trigger sidebar reload
	w.Header().Set("HX-Trigger", "requestsChanged")
	w.WriteHeader(http.StatusOK)
}
