package web

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anu-mdl/linker-bruno/server/loader"
	"github.com/anu-mdl/linker-bruno/server/parser"
)

// GetRequestByID loads a request by its ID
func GetRequestByID(baseDir, id string) (*parser.BrunoRequest, error) {
	// Convert ID back to file path
	filePath := idToFilePath(id)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("request not found: %s", id)
	}

	// Parse the file
	req, err := parser.ParseBrunoFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	return req, nil
}

// SaveRequest saves a new request to a .bru file
func SaveRequest(baseDir string, req *parser.BrunoRequest) error {
	// Generate file path from URL
	filePath := generateFilePath(baseDir, req.URL, req.Meta.Name)

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Serialize request to .bru format
	content := serializeRequest(req)

	// Write to file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Update FilePath in request
	req.FilePath = filePath

	return nil
}

// UpdateRequest updates an existing request
func UpdateRequest(baseDir, id string, req *parser.BrunoRequest) error {
	// Get original file path
	filePath := idToFilePath(id)

	// Serialize request to .bru format
	content := serializeRequest(req)

	// Write to file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// DeleteRequest deletes a request file
func DeleteRequest(baseDir, id string) error {
	// Get file path
	filePath := idToFilePath(id)

	// Delete file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// generateFilePath generates a file path from URL and name
func generateFilePath(baseDir, url, name string) string {
	// Parse URL into segments
	segments := parseURLSegments(url)

	// Build directory path
	dirPath := baseDir
	if len(segments) > 0 {
		// Use URL segments as folder structure
		dirPath = filepath.Join(baseDir, filepath.Join(segments[:len(segments)-1]...))
	}

	// Sanitize name for filename
	filename := sanitizeFilename(name) + ".bru"

	return filepath.Join(dirPath, filename)
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(name string) string {
	// Replace invalid characters
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, ":", "-")
	name = strings.ReplaceAll(name, "*", "-")
	name = strings.ReplaceAll(name, "?", "-")
	name = strings.ReplaceAll(name, "\"", "-")
	name = strings.ReplaceAll(name, "<", "-")
	name = strings.ReplaceAll(name, ">", "-")
	name = strings.ReplaceAll(name, "|", "-")
	return name
}

// idToFilePath converts an ID back to a file path
func idToFilePath(id string) string {
	// Reverse the ID generation - order matters!
	// Process in reverse order from generateID
	path := strings.NewReplacer(
		"~~", "/",
		"--", ".",
		"__", " ",
		"-_", "-",
		"_-", "_",
	).Replace(id)
	return path
}

// serializeRequest converts a BrunoRequest to .bru file format
func serializeRequest(req *parser.BrunoRequest) string {
	var sb strings.Builder

	// Meta block
	sb.WriteString("meta {\n")
	sb.WriteString(fmt.Sprintf("  name: %s\n", req.Meta.Name))
	sb.WriteString(fmt.Sprintf("  type: %s\n", req.Meta.Type))
	sb.WriteString(fmt.Sprintf("  seq: %d\n", req.Meta.Seq))
	sb.WriteString("}\n\n")

	// HTTP method block
	sb.WriteString(fmt.Sprintf("%s {\n", strings.ToLower(req.Method)))
	sb.WriteString(fmt.Sprintf("  url: %s\n", req.URL))
	sb.WriteString("}\n\n")

	// Headers block (request headers)
	if len(req.Headers) > 0 {
		sb.WriteString("headers {\n")
		for key, value := range req.Headers {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
		sb.WriteString("}\n\n")
	}

	// Query params block
	if len(req.QueryParams) > 0 {
		sb.WriteString("params:query {\n")
		for key, value := range req.QueryParams {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
		sb.WriteString("}\n\n")
	}

	// Body block (request body)
	if req.Body != "" {
		sb.WriteString("body:json {\n")
		sb.WriteString(req.Body)
		sb.WriteString("\n}\n\n")
	}

	// Example block
	sb.WriteString("example {\n")
	sb.WriteString(fmt.Sprintf("  name: %s\n", req.Example.Name))
	if req.Example.Description != "" {
		sb.WriteString(fmt.Sprintf("  description: %s\n", req.Example.Description))
	}
	sb.WriteString("\n")

	// Request block
	sb.WriteString("  request: {\n")
	sb.WriteString(fmt.Sprintf("    url: %s\n", req.Example.Request.URL))
	sb.WriteString(fmt.Sprintf("    method: %s\n", req.Example.Request.Method))
	sb.WriteString(fmt.Sprintf("    mode: %s\n", req.Example.Request.Mode))
	sb.WriteString("  }\n\n")

	// Response block
	sb.WriteString("  response: {\n")

	// Headers
	if len(req.Example.Response.Headers) > 0 {
		sb.WriteString("    headers: {\n")
		for key, value := range req.Example.Response.Headers {
			sb.WriteString(fmt.Sprintf("      %s: %s\n", key, value))
		}
		sb.WriteString("    }\n\n")
	}

	// Status
	sb.WriteString("    status: {\n")
	sb.WriteString(fmt.Sprintf("      code: %d\n", req.Example.Response.Status.Code))
	sb.WriteString(fmt.Sprintf("      text: %s\n", req.Example.Response.Status.Text))
	sb.WriteString("    }\n\n")

	// Body
	sb.WriteString("    body: {\n")
	sb.WriteString(fmt.Sprintf("      type: %s\n", req.Example.Response.Body.Type))
	sb.WriteString("      content: '''\n")
	sb.WriteString(req.Example.Response.Body.Content)
	sb.WriteString("\n      '''\n")
	sb.WriteString("    }\n")

	sb.WriteString("  }\n")
	sb.WriteString("}\n")

	return sb.String()
}

// ListAllRequests returns all requests in the base directory
func ListAllRequests(baseDir string) ([]*parser.BrunoRequest, error) {
	return loader.LoadAllRequests(baseDir)
}
