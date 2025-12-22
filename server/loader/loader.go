package loader

import (
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

		requests = append(requests, req)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", baseDir, err)
	}

	return requests, nil
}
