package service

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/anu-mdl/linker-bruno/internal/shared/brunoformat"
	"github.com/anu-mdl/linker-bruno/internal/shared/urlutil"
	"github.com/go-chi/chi/v5"
)

// MockService handles business logic for mock endpoint registration and response generation
type MockService struct {
	converter *urlutil.Converter
}

// NewMockService creates a new MockService
func NewMockService(converter *urlutil.Converter) *MockService {
	return &MockService{
		converter: converter,
	}
}

// RegisterRoutes registers all Bruno requests as routes on the given router
func (s *MockService) RegisterRoutes(router *chi.Mux, requests []*brunoformat.BrunoRequest, envVars map[string]string) error {
	for _, req := range requests {
		// Convert Bruno URL pattern to chi route pattern
		path := s.converter.ConvertPattern(req.URL, envVars)

		// Create handler for this request
		handler := s.createHandler(req)

		// Register the route with the appropriate method
		router.Method(req.Method, path, handler)

		log.Printf("Registered: %s %s (from %s)", req.Method, path, req.FilePath)
	}
	return nil
}

// createHandler creates an HTTP handler for a Bruno request
func (s *MockService) createHandler(req *brunoformat.BrunoRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract path parameters
		params := s.extractPathParams(r)

		// Parse the body content from the example block
		var body interface{}
		if req.Example.Response.Body.Content != "" {
			// Unmarshal the body content as JSON
			if err := json.Unmarshal([]byte(req.Example.Response.Body.Content), &body); err != nil {
				log.Printf("Warning: failed to parse response body for %s: %v", req.FilePath, err)
				body = nil
			}
		}

		// Interpolate variables in response body
		body = s.interpolateVariables(body, params)

		// Set custom headers from example block
		for key, value := range req.Example.Response.Headers {
			w.Header().Set(key, value)
		}

		// Set Content-Type if not already set
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "application/json")
		}

		// Set status code from example block
		w.WriteHeader(req.Example.Response.Status.Code)

		// Write response body
		if body != nil {
			json.NewEncoder(w).Encode(body)
		}
	}
}

// extractPathParams extracts all path parameters from the request using chi
func (s *MockService) extractPathParams(r *http.Request) map[string]string {
	params := make(map[string]string)

	// chi stores URL parameters in the request context
	// We can get them using chi.URLParam, but we need to know the param names
	// For now, we'll extract all params from the URL pattern
	rctx := chi.RouteContext(r.Context())
	if rctx != nil {
		for i, key := range rctx.URLParams.Keys {
			if i < len(rctx.URLParams.Values) {
				params[key] = rctx.URLParams.Values[i]
			}
		}
	}

	return params
}

// interpolateVariables replaces {{varName}} placeholders in the response body
func (s *MockService) interpolateVariables(body interface{}, vars map[string]string) interface{} {
	if body == nil {
		return nil
	}

	// Convert body to JSON string
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return body // Return original if we can't marshal
	}

	jsonStr := string(jsonBytes)

	// Replace {{varName}} with actual values
	for key, value := range vars {
		placeholder := "{{" + key + "}}"
		jsonStr = strings.ReplaceAll(jsonStr, placeholder, value)
	}

	// Parse back to interface{}
	var result interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return body // Return original if we can't unmarshal
	}

	return result
}
