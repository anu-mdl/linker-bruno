package brunoformat

import (
	"fmt"
	"strings"
)

// Serializer handles conversion of BrunoRequest to .bru file format
type Serializer struct{}

// NewSerializer creates a new Serializer
func NewSerializer() *Serializer {
	return &Serializer{}
}

// Serialize converts a BrunoRequest to .bru file format
func (s *Serializer) Serialize(req *BrunoRequest) string {
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
