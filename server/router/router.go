package router

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/anu-mdl/linker-bruno/server/parser"
	"github.com/go-chi/chi/v5"
)

// CreateRouter creates a new chi router
func CreateRouter() *chi.Mux {
	return chi.NewRouter()
}

// RegisterRoutes registers all Bruno requests as routes on the given router
func RegisterRoutes(router *chi.Mux, requests []*parser.BrunoRequest, envVars map[string]string) {
	for _, req := range requests {
		// Convert Bruno URL pattern to chi route pattern
		path := convertURLPattern(req.URL, envVars)

		// Create handler for this request
		handler := createHandler(req)

		// Register the route with the appropriate method
		router.Method(req.Method, path, handler)

		log.Printf("Registered: %s %s (from %s)", req.Method, path, req.FilePath)
	}
}

// convertURLPattern converts a Bruno URL pattern to a chi route pattern
func convertURLPattern(brunoURL string, envVars map[string]string) string {
	path := brunoURL

	// Replace environment variables like {{baseUrl}}
	for key, value := range envVars {
		placeholder := "{{" + key + "}}"
		path = strings.ReplaceAll(path, placeholder, value)
	}

	// Remove any remaining {{variable}} placeholders (like {{baseUrl}})
	// by stripping them out entirely
	varRe := regexp.MustCompile(`\{\{[^}]+\}\}`)
	path = varRe.ReplaceAllString(path, "")

	// Convert :param to {param} for chi (chi actually supports both, but {param} is more standard)
	// First, handle :param patterns
	paramRe := regexp.MustCompile(`:(\w+)`)
	path = paramRe.ReplaceAllString(path, "{$1}")

	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Clean up double slashes
	path = regexp.MustCompile(`/+`).ReplaceAllString(path, "/")

	return path
}

// createHandler creates an HTTP handler for a Bruno request
func createHandler(req *parser.BrunoRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract path parameters
		params := extractPathParams(r)

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
		body = interpolateVariables(body, params)

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
func extractPathParams(r *http.Request) map[string]string {
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
func interpolateVariables(body interface{}, vars map[string]string) interface{} {
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
