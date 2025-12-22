package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// LoadEnvironment loads environment variables from a .bru environment file
func LoadEnvironment(envName string, baseDir string) (map[string]string, error) {
	envPath := filepath.Join(baseDir, "environments", envName+".bru")

	// If environment file doesn't exist, return empty map (not an error)
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return make(map[string]string), nil
	}

	content, err := os.ReadFile(envPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read environment file %s: %w", envPath, err)
	}

	vars := parseVarsBlock(string(content))
	return vars, nil
}

// parseVarsBlock parses the vars { ... } block from an environment file
func parseVarsBlock(content string) map[string]string {
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
