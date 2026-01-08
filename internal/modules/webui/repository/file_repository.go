package repository

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/anu-mdl/linker-bruno/internal/shared/brunoformat"
)

// FileRepository handles file I/O operations for .bru files
type FileRepository struct {
	serializer *brunoformat.Serializer
}

// NewFileRepository creates a new FileRepository
func NewFileRepository(serializer *brunoformat.Serializer) *FileRepository {
	return &FileRepository{
		serializer: serializer,
	}
}

// ReadFile reads and parses a .bru file by its file path
func (r *FileRepository) ReadFile(filePath string) (*brunoformat.BrunoRequest, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	// Parse the file
	req, err := brunoformat.ParseBrunoFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	return req, nil
}

// WriteFile writes a BrunoRequest to a .bru file
func (r *FileRepository) WriteFile(filePath string, req *brunoformat.BrunoRequest) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Serialize request to .bru format
	content := r.serializer.Serialize(req)

	// Write to file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// DeleteFile deletes a .bru file
func (r *FileRepository) DeleteFile(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// ScanDirectory recursively scans a directory for all .bru files
func (r *FileRepository) ScanDirectory(baseDir string) ([]*brunoformat.BrunoRequest, error) {
	var requests []*brunoformat.BrunoRequest

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

		// Skip collection.bru file
		if info.Name() == "collection.bru" {
			return nil
		}

		// Parse the .bru file
		req, err := brunoformat.ParseBrunoFile(path)
		if err != nil {
			log.Printf("Warning: failed to parse %s: %v", path, err)
			return nil // Continue walking
		}

		requests = append(requests, req)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", baseDir, err)
	}

	return requests, nil
}

// GenerateFilePath generates a file path from base directory, URL, and name
func (r *FileRepository) GenerateFilePath(baseDir, url, name string) string {
	// Parse URL into segments
	segments := r.parseURLSegments(url)

	// Build directory path
	dirPath := baseDir
	if len(segments) > 1 {
		// Use URL segments as folder structure (excluding last segment)
		dirPath = filepath.Join(baseDir, filepath.Join(segments[:len(segments)-1]...))
	}

	// Sanitize name for filename
	filename := r.SanitizeFilename(name) + ".bru"

	return filepath.Join(dirPath, filename)
}

// SanitizeFilename removes invalid characters from filename
func (r *FileRepository) SanitizeFilename(name string) string {
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

// parseURLSegments splits a URL into segments for directory structure
func (r *FileRepository) parseURLSegments(url string) []string {
	// Remove leading slash
	url = strings.TrimPrefix(url, "/")

	// Remove environment variables
	url = strings.ReplaceAll(url, "{{baseUrl}}", "")
	url = strings.TrimPrefix(url, "/")

	if url == "" {
		return []string{}
	}

	// Split by slash
	segments := strings.Split(url, "/")

	// Clean up segments
	cleaned := make([]string, 0)
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg != "" {
			cleaned = append(cleaned, seg)
		}
	}

	return cleaned
}
