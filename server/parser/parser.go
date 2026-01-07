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
		FilePath:    filepath,
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
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

	// Parse headers block
	if headersContent, ok := blocks["headers"]; ok {
		req.Headers = parseKeyValueBlock(headersContent)
	}

	// Parse params:query block
	if paramsContent, ok := blocks["params:query"]; ok {
		req.QueryParams = parseKeyValueBlock(paramsContent)
	}

	// Parse body:json block
	if bodyContent, ok := blocks["body:json"]; ok {
		req.Body = strings.TrimSpace(bodyContent)
	}

	// Parse example block
	if exampleContent, ok := blocks["example"]; ok {
		example, err := parseExampleBlock(exampleContent)
		if err != nil {
			return nil, fmt.Errorf("failed to parse example block: %w", err)
		}
		req.Example = example
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
			// Strip trailing colon if present (e.g., "request:" -> "request")
			blockName = strings.TrimSuffix(blockName, ":")
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
			// Count braces in this line
			openCount := strings.Count(line, "{")
			closeCount := strings.Count(line, "}")

			// Add line to content ONLY if it won't close the block
			if braceCount+openCount-closeCount > 0 {
				blockContent.WriteString(line)
				blockContent.WriteString("\n")
			}

			braceCount += openCount
			braceCount -= closeCount

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

// parseExampleBlock parses the example block
func parseExampleBlock(content string) (ExampleBlock, error) {
	example := ExampleBlock{
		Response: ExampleResponse{
			Headers: make(map[string]string),
		},
	}

	// Parse top-level key:value pairs and extract nested blocks
	blocks := extractBlocks(content)

	// Parse top-level fields (name, description)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "{") {
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
			example.Name = value
		case "description":
			example.Description = value
		}
	}

	// Parse request block
	if requestContent, ok := blocks["request"]; ok {
		example.Request = parseExampleRequestBlock(requestContent)
	}

	// Parse response block
	if responseContent, ok := blocks["response"]; ok {
		response, err := parseExampleResponseBlock(responseContent)
		if err != nil {
			return example, err
		}
		example.Response = response
	}

	return example, nil
}

// parseExampleRequestBlock parses the request block within example
func parseExampleRequestBlock(content string) ExampleRequest {
	request := ExampleRequest{}

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
		case "url":
			request.URL = value
		case "method":
			request.Method = value
		case "mode":
			request.Mode = value
		}
	}

	return request
}

// parseExampleResponseBlock parses the response block within example
func parseExampleResponseBlock(content string) (ExampleResponse, error) {
	response := ExampleResponse{
		Headers: make(map[string]string),
	}

	// Extract nested blocks (headers, status, body)
	blocks := extractBlocks(content)

	// Parse headers block
	if headersContent, ok := blocks["headers"]; ok {
		response.Headers = parseKeyValueBlock(headersContent)
	}

	// Parse status block
	if statusContent, ok := blocks["status"]; ok {
		response.Status = parseStatusBlock(statusContent)
	}

	// Parse body block
	if bodyContent, ok := blocks["body"]; ok {
		body, err := parseBodyBlock(bodyContent)
		if err != nil {
			return response, err
		}
		response.Body = body
	}

	return response, nil
}

// parseKeyValueBlock parses a block with key: value pairs
func parseKeyValueBlock(content string) map[string]string {
	result := make(map[string]string)

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
		result[key] = value
	}

	return result
}

// parseStatusBlock parses the status block
func parseStatusBlock(content string) ExampleStatus {
	status := ExampleStatus{}

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
		case "code":
			if code, err := strconv.Atoi(value); err == nil {
				status.Code = code
			}
		case "text":
			status.Text = value
		}
	}

	return status
}

// parseBodyBlock parses the body block with triple-quoted content
func parseBodyBlock(content string) (ExampleBody, error) {
	body := ExampleBody{}

	// Look for type and content fields
	lines := strings.Split(content, "\n")
	inTripleQuote := false
	var contentBuilder strings.Builder

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Parse type field
		if strings.HasPrefix(trimmed, "type:") {
			body.Type = strings.TrimSpace(strings.TrimPrefix(trimmed, "type:"))
			continue
		}

		// Check for content field with triple quotes
		if strings.HasPrefix(trimmed, "content:") {
			// Check if triple quotes are on the same line or next line
			if strings.Contains(line, "'''") {
				inTripleQuote = true
				continue
			}
			// Check next line for triple quotes
			if i+1 < len(lines) && strings.Contains(strings.TrimSpace(lines[i+1]), "'''") {
				inTripleQuote = true
				continue
			}
		}

		// Handle triple-quoted content
		if inTripleQuote {
			if strings.Contains(trimmed, "'''") {
				inTripleQuote = false
				continue
			}
			contentBuilder.WriteString(line)
			contentBuilder.WriteString("\n")
		}
	}

	body.Content = strings.TrimSpace(contentBuilder.String())
	return body, nil
}

