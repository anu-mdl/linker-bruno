package repository

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/anu-mdl/linker-bruno/internal/shared/brunoformat"
)

// BruRepository handles loading .bru files and environment variables from the filesystem
type BruRepository struct{}

// NewBruRepository creates a new BruRepository
func NewBruRepository() *BruRepository {
	return &BruRepository{}
}

// LoadAllRequests recursively scans a directory for .bru files and parses them
func (r *BruRepository) LoadAllRequests(baseDir string) ([]*brunoformat.BrunoRequest, error) {
	var requests []*brunoformat.BrunoRequest

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
		req, err := brunoformat.ParseBrunoFile(path)
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

		// Validate that an example block exists
		if req.Example.Response.Status.Code == 0 {
			log.Printf("Warning: skipping %s - no example block with valid response found", path)
			return nil
		}

		requests = append(requests, req)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", baseDir, err)
	}

	return requests, nil
}

// LoadEnvironment loads environment variables from a .bru environment file
func (r *BruRepository) LoadEnvironment(envName string, baseDir string) (map[string]string, error) {
	envPath := filepath.Join(baseDir, "environments", envName+".bru")

	// If environment file doesn't exist, return empty map (not an error)
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return make(map[string]string), nil
	}

	content, err := os.ReadFile(envPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read environment file %s: %w", envPath, err)
	}

	vars := r.parseVarsBlock(string(content))
	return vars, nil
}

// parseVarsBlock parses the vars { ... } block from an environment file
func (r *BruRepository) parseVarsBlock(content string) map[string]string {
	vars := make(map[string]string)

	// Extract vars { ... } block
	varsRe := regexp.MustCompile(`vars\s*\{([^}]*)\}`)
	match := varsRe.FindStringSubmatch(content)
	if match == nil {
		return vars
	}

	// Parse key: value pairs
	lines := strings.Split(match[1], "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			vars[key] = value
		}
	}

	return vars
}
