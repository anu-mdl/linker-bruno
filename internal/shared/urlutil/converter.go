package urlutil

import (
	"regexp"
	"strings"
)

// Converter handles URL pattern conversions and ID encoding/decoding
type Converter struct{}

// NewConverter creates a new URL Converter
func NewConverter() *Converter {
	return &Converter{}
}

// ConvertPattern converts a Bruno URL pattern to a chi route pattern
// Replaces environment variables and converts :param to {param}
func (c *Converter) ConvertPattern(brunoURL string, envVars map[string]string) string {
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

// EncodeID generates a unique URL-safe identifier from file path
// Uses multi-character replacements to avoid conflicts
func (c *Converter) EncodeID(filePath string) string {
	// Use URL-safe encoding to handle all special characters
	// Order matters - process in this specific sequence
	id := strings.NewReplacer(
		"/", "~~",
		".", "--",
		" ", "__",
		"-", "-_",
		"_", "_-",
	).Replace(filePath)
	return id
}

// DecodeID converts a URL-safe ID back to file path
// Reverses the encoding done by EncodeID
func (c *Converter) DecodeID(id string) string {
	// Reverse the ID generation - order matters!
	// Process in reverse order from EncodeID
	path := strings.NewReplacer(
		"~~", "/",
		"--", ".",
		"__", " ",
		"-_", "-",
		"_-", "_",
	).Replace(id)
	return path
}

// DisplayURL converts URL path parameters from {param} to [param] for UI display
// Also ensures the URL starts with a leading slash
func (c *Converter) DisplayURL(url string) string {
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

// DisplayParam wraps a parameter name in brackets for display
func (c *Converter) DisplayParam(param string) string {
	return "[" + param + "]"
}

// CleanPath normalizes a path by ensuring it starts with / and removing double slashes
func (c *Converter) CleanPath(path string) string {
	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Clean up double slashes
	path = regexp.MustCompile(`/+`).ReplaceAllString(path, "/")

	return path
}
