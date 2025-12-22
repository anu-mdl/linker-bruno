package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
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

	// Parse response block
	if responseContent, ok := blocks["response"]; ok {
		resp, err := parseResponseBlock(responseContent)
		if err == nil {
			req.Response = resp
		}
	}

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

// parseResponseBlock parses the response block
func parseResponseBlock(content string) (ResponseBlock, error) {
	resp := ResponseBlock{
		Status:  200,
		Headers: make(map[string]string),
	}

	// Extract status
	statusRe := regexp.MustCompile(`status:\s*(\d+)`)
	if match := statusRe.FindStringSubmatch(content); match != nil {
		if status, err := strconv.Atoi(match[1]); err == nil {
			resp.Status = status
		}
	}

	// Extract headers block
	headersRe := regexp.MustCompile(`headers\s*\{([^}]*)\}`)
	if match := headersRe.FindStringSubmatch(content); match != nil {
		resp.Headers = parseHeaders(match[1])
	}

	// Extract body block - this is tricky because it can contain nested braces/brackets (JSON)
	bodyStart := strings.Index(content, "body")
	if bodyStart != -1 {
		// Find the opening delimiter after "body" (either { or [)
		restOfContent := content[bodyStart:]
		openBraceIdx := strings.Index(restOfContent, "{")
		openBracketIdx := strings.Index(restOfContent, "[")

		// Determine which comes first (and exists)
		var delimStart int
		var openDelim, closeDelim byte

		if openBraceIdx != -1 && (openBracketIdx == -1 || openBraceIdx < openBracketIdx) {
			delimStart = openBraceIdx
			openDelim = '{'
			closeDelim = '}'
		} else if openBracketIdx != -1 {
			delimStart = openBracketIdx
			openDelim = '['
			closeDelim = ']'
		} else {
			return resp, nil // No body content found
		}

		bodyStart += delimStart // Don't skip the opening delimiter

		// Count delimiters to find the matching closing delimiter
		delimCount := 0
		bodyEnd := bodyStart

		for bodyEnd < len(content) {
			if content[bodyEnd] == openDelim {
				delimCount++
			} else if content[bodyEnd] == closeDelim {
				delimCount--
				if delimCount == 0 {
					bodyEnd++
					break
				}
			}
			bodyEnd++
		}

		if delimCount == 0 {
			bodyContent := strings.TrimSpace(content[bodyStart:bodyEnd])

			// Try to parse as JSON
			var jsonBody interface{}
			if err := json.Unmarshal([]byte(bodyContent), &jsonBody); err == nil {
				resp.Body = jsonBody
			} else {
				// If not valid JSON, store as string
				resp.Body = bodyContent
			}
		}
	}

	return resp, nil
}

// parseHeaders parses the headers block content
func parseHeaders(content string) map[string]string {
	headers := make(map[string]string)

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			headers[key] = value
		}
	}

	return headers
}
