package web

import (
	"sort"
	"strings"

	"github.com/anu-mdl/linker-bruno/server/parser"
)

// TreeNode represents a folder or request in the tree
type TreeNode struct {
	Name      string
	Type      string // "folder" or "request"
	Method    string // HTTP method (for requests only)
	URL       string // Full URL path
	ID        string // Unique identifier for the request
	IsDynamic bool   // True if this is a dynamic parameter folder/segment
	Children  []*TreeNode
}

// BuildRequestTree builds a folder tree from requests based on URL paths
func BuildRequestTree(requests []*parser.BrunoRequest) *TreeNode {
	root := &TreeNode{
		Name:     "root",
		Type:     "folder",
		Children: make([]*TreeNode, 0),
	}

	// Group requests by URL path
	for _, req := range requests {
		// Generate unique ID from file path
		id := generateID(req.FilePath)

		// Parse URL into segments
		segments := parseURLSegments(req.URL)

		// Insert into tree
		insertIntoTree(root, segments, req, id)
	}

	// Sort tree
	sortTree(root)

	return root
}

// parseURLSegments splits a URL into segments, handling placeholders
func parseURLSegments(url string) []string {
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

// insertIntoTree inserts a request into the tree at the appropriate location
func insertIntoTree(node *TreeNode, segments []string, req *parser.BrunoRequest, id string) {
	if len(segments) == 0 {
		return
	}

	// If this is the last segment, add as a request node
	if len(segments) == 1 {
		// Check if request already exists
		for _, child := range node.Children {
			if child.Type == "request" && child.URL == req.URL && child.Method == req.Method {
				return // Already exists
			}
		}

		// Add request node
		requestNode := &TreeNode{
			Name:     req.Meta.Name,
			Type:     "request",
			Method:   req.Method,
			URL:      req.URL,
			ID:       id,
			Children: nil,
		}
		node.Children = append(node.Children, requestNode)
		return
	}

	// Find or create folder for this segment
	segment := segments[0]
	isDynamic := false

	// Check if this is a dynamic segment
	if strings.HasPrefix(segment, ":") || strings.HasPrefix(segment, "{") {
		segment = extractParamName(segment)
		isDynamic = true
	}

	var folderNode *TreeNode
	for _, child := range node.Children {
		if child.Type == "folder" && child.Name == segment {
			folderNode = child
			break
		}
	}

	if folderNode == nil {
		folderNode = &TreeNode{
			Name:      segment,
			Type:      "folder",
			IsDynamic: isDynamic,
			Children:  make([]*TreeNode, 0),
		}
		node.Children = append(node.Children, folderNode)
	}

	// Recursively insert into subfolder
	insertIntoTree(folderNode, segments[1:], req, id)
}

// extractParamName extracts parameter name from :param or {param}
func extractParamName(segment string) string {
	segment = strings.TrimPrefix(segment, ":")
	segment = strings.TrimPrefix(segment, "{")
	segment = strings.TrimSuffix(segment, "}")
	return segment
}

// sortTree sorts the tree nodes alphabetically (folders first, then requests)
func sortTree(node *TreeNode) {
	if node.Children == nil {
		return
	}

	// Sort children
	sort.Slice(node.Children, func(i, j int) bool {
		// Folders come before requests
		if node.Children[i].Type != node.Children[j].Type {
			return node.Children[i].Type == "folder"
		}
		// Alphabetical within same type
		return node.Children[i].Name < node.Children[j].Name
	})

	// Recursively sort children
	for _, child := range node.Children {
		sortTree(child)
	}
}

// generateID generates a unique identifier from file path
func generateID(filePath string) string {
	// Use URL-safe encoding to handle all special characters
	id := strings.NewReplacer(
		"/", "~~",
		".", "--",
		" ", "__",
		"-", "-_",
		"_", "_-",
	).Replace(filePath)
	return id
}
