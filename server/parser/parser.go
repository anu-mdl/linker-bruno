package parser

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ParseBrunoFile parses a .bru file and returns a BrunoRequest
func ParseBrunoFile(filepath string) (*BrunoRequest, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filepath, err)
	}

	req := &BrunoRequest{
		FilePath: filepath,
		Response: ResponseBlock{
			Status:  200, // default status
			Headers: make(map[string]string),
		},
	}

	// Extract all blocks from the file
	blocks := extractBlocks(string(content))

	// Parse meta block
	if metaContent, ok := blocks["meta"]; ok {
		req.Meta = parseMetaBlock(metaContent)
	}

	// Parse HTTP method blocks
	for _, method := range []string{"get", "post", "put", "delete", "patch"} {
		if methodContent, ok := blocks[method]; ok {
			req.Method = strings.ToUpper(method)
			req.URL = parseMethodBlock(methodContent)
			break
		}
	}

	// Response will be loaded separately from .response.json files

	return req, nil
}

// extractBlocks extracts all top-level blocks from the .bru file content
func extractBlocks(content string) map[string]string {
	blocks := make(map[string]string)

	// Regex to match block_name { ... content ... }
	// This handles nested braces by counting them
	lines := strings.Split(content, "\n")
	var currentBlock string
	var blockContent strings.Builder
	braceCount := 0
	inBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if this line starts a new block
		if !inBlock && strings.Contains(trimmed, "{") {
			parts := strings.SplitN(trimmed, "{", 2)
			blockName := strings.TrimSpace(parts[0])
			if blockName != "" {
				currentBlock = blockName
				inBlock = true
				braceCount = 1

				// Add the rest of the line after the opening brace
				rest := strings.TrimSpace(parts[1])
				if rest != "" {
					blockContent.WriteString(rest)
					blockContent.WriteString("\n")
				}

				// Count additional braces on this line
				braceCount += strings.Count(rest, "{")
				braceCount -= strings.Count(rest, "}")

				if braceCount == 0 {
					blocks[currentBlock] = strings.TrimSpace(blockContent.String())
					blockContent.Reset()
					inBlock = false
				}
				continue
			}
		}

		// If we're in a block, accumulate content
		if inBlock {
			blockContent.WriteString(line)
			blockContent.WriteString("\n")

			braceCount += strings.Count(line, "{")
			braceCount -= strings.Count(line, "}")

			if braceCount == 0 {
				blocks[currentBlock] = strings.TrimSpace(blockContent.String())
				blockContent.Reset()
				inBlock = false
			}
		}
	}

	return blocks
}

// parseMetaBlock parses the meta block
func parseMetaBlock(content string) MetaBlock {
	meta := MetaBlock{}

	// Parse key: value pairs
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "}" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "name":
			meta.Name = value
		case "type":
			meta.Type = value
		case "seq":
			if seq, err := strconv.Atoi(value); err == nil {
				meta.Seq = seq
			}
		}
	}

	return meta
}

// parseMethodBlock parses the HTTP method block (get, post, etc.)
func parseMethodBlock(content string) string {
	// Look for url: value
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "url:") {
			url := strings.TrimSpace(strings.TrimPrefix(line, "url:"))
			return url
		}
	}
	return ""
}

