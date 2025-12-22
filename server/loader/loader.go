package loader

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/anu-mdl/linker-bruno/server/parser"
)

// LoadAllRequests recursively scans a directory for .bru files and parses them
func LoadAllRequests(baseDir string) ([]*parser.BrunoRequest, error) {
	var requests []*parser.BrunoRequest

	// Walk the directory tree
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Warning: failed to access path %s: %v", path, err)
			return nil // Continue walking
		}

		// Skip directories
		if info.IsDir() {
			// Skip environments directory
			if info.Name() == "environments" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .bru files
		if !strings.HasSuffix(info.Name(), ".bru") {
			return nil
		}

		// Skip collection.bru file (collection-level config)
		if info.Name() == "collection.bru" {
			return nil
		}

		// Parse the .bru file
		req, err := parser.ParseBrunoFile(path)
		if err != nil {
			log.Printf("Warning: failed to parse %s: %v", path, err)
			return nil // Continue walking
		}

		// Only include requests with a valid HTTP method
		if req.Method == "" {
			log.Printf("Warning: skipping %s - no HTTP method found", path)
			return nil
		}

		// Only include requests with a URL
		if req.URL == "" {
			log.Printf("Warning: skipping %s - no URL found", path)
			return nil
		}

		// Load the corresponding .response.json file if it exists
		loadResponseFile(path, req)

		requests = append(requests, req)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", baseDir, err)
	}

	return requests, nil
}

// loadResponseFile loads the .response.json file for a .bru file
func loadResponseFile(bruPath string, req *parser.BrunoRequest) {
	// Convert "Get User.bru" to "Get User.response.json"
	responsePath := strings.TrimSuffix(bruPath, ".bru") + ".response.json"

	// Check if the response file exists
	if _, err := os.Stat(responsePath); os.IsNotExist(err) {
		// No response file, use default
		req.Response = generateDefaultResponse(req)
		return
	}

	// Read the response file
	content, err := os.ReadFile(responsePath)
	if err != nil {
		log.Printf("Warning: failed to read response file %s: %v", responsePath, err)
		req.Response = generateDefaultResponse(req)
		return
	}

	// Parse the JSON
	var response parser.ResponseBlock
	if err := json.Unmarshal(content, &response); err != nil {
		log.Printf("Warning: failed to parse response file %s: %v", responsePath, err)
		req.Response = generateDefaultResponse(req)
		return
	}

	// Ensure headers map is initialized
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}

	req.Response = response
}

// generateDefaultResponse creates a default response when no .response.json file exists
func generateDefaultResponse(req *parser.BrunoRequest) parser.ResponseBlock {
	return parser.ResponseBlock{
		Status: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"message": "Mock response for " + req.Meta.Name,
			"method":  req.Method,
			"url":     req.URL,
		},
	}
}

