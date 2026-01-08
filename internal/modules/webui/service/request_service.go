package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/anu-mdl/linker-bruno/internal/modules/webui/dto"
	"github.com/anu-mdl/linker-bruno/internal/modules/webui/repository"
	"github.com/anu-mdl/linker-bruno/internal/shared/brunoformat"
	"github.com/anu-mdl/linker-bruno/internal/shared/urlutil"
)

// RequestService handles business logic for request CRUD operations and tree building
type RequestService struct {
	repo      *repository.FileRepository
	converter *urlutil.Converter
}

// NewRequestService creates a new RequestService
func NewRequestService(repo *repository.FileRepository, converter *urlutil.Converter) *RequestService {
	return &RequestService{
		repo:      repo,
		converter: converter,
	}
}

// GetRequestByID retrieves a request by its ID
func (s *RequestService) GetRequestByID(id string) (*dto.RequestResponse, error) {
	// Decode ID to file path
	filePath := s.converter.DecodeID(id)

	// Read file
	req, err := s.repo.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read request: %w", err)
	}

	// Convert to response DTO
	response := &dto.RequestResponse{
		ID:          id,
		Name:        req.Meta.Name,
		Method:      req.Method,
		URL:         req.URL,
		Description: req.Example.Description,
		Headers:     req.Headers,
		QueryParams: req.QueryParams,
		Body:        req.Body,
		ResponseHeaders: req.Example.Response.Headers,
		ResponseBody:    req.Example.Response.Body.Content,
	}
	response.ResponseStatus.Code = req.Example.Response.Status.Code
	response.ResponseStatus.Text = req.Example.Response.Status.Text

	return response, nil
}

// CreateRequest creates a new request
func (s *RequestService) CreateRequest(baseDir string, input *dto.CreateRequestInput) error {
	// Create BrunoRequest from input
	req := &brunoformat.BrunoRequest{
		Meta: brunoformat.MetaBlock{
			Name: input.Name,
			Type: "http",
			Seq:  1,
		},
		Method:      input.Method,
		URL:         input.URL,
		Headers:     input.Headers,
		QueryParams: input.QueryParams,
		Body:        input.Body,
		Example: brunoformat.ExampleBlock{
			Name:        input.Name + " Example",
			Description: input.Description,
			Request: brunoformat.ExampleRequest{
				URL:    input.URL,
				Method: input.Method,
				Mode:   "none",
			},
			Response: brunoformat.ExampleResponse{
				Headers: input.ResponseHeaders,
				Status: brunoformat.ExampleStatus{
					Code: input.ResponseStatus.Code,
					Text: input.ResponseStatus.Text,
				},
				Body: brunoformat.ExampleBody{
					Type:    "json",
					Content: input.ResponseBody,
				},
			},
		},
	}

	// Initialize maps if nil
	if req.Headers == nil {
		req.Headers = make(map[string]string)
	}
	if req.QueryParams == nil {
		req.QueryParams = make(map[string]string)
	}
	if req.Example.Response.Headers == nil {
		req.Example.Response.Headers = make(map[string]string)
	}

	// Generate file path
	filePath := s.repo.GenerateFilePath(baseDir, input.URL, input.Name)
	req.FilePath = filePath

	// Write file
	if err := s.repo.WriteFile(filePath, req); err != nil {
		return fmt.Errorf("failed to save request: %w", err)
	}

	return nil
}

// UpdateRequest updates an existing request
func (s *RequestService) UpdateRequest(id string, input *dto.UpdateRequestInput) error {
	// Decode ID to file path
	filePath := s.converter.DecodeID(id)

	// Create updated BrunoRequest from input
	req := &brunoformat.BrunoRequest{
		FilePath: filePath,
		Meta: brunoformat.MetaBlock{
			Name: input.Name,
			Type: "http",
			Seq:  1,
		},
		Method:      input.Method,
		URL:         input.URL,
		Headers:     input.Headers,
		QueryParams: input.QueryParams,
		Body:        input.Body,
		Example: brunoformat.ExampleBlock{
			Name:        input.Name + " Example",
			Description: input.Description,
			Request: brunoformat.ExampleRequest{
				URL:    input.URL,
				Method: input.Method,
				Mode:   "none",
			},
			Response: brunoformat.ExampleResponse{
				Headers: input.ResponseHeaders,
				Status: brunoformat.ExampleStatus{
					Code: input.ResponseStatus.Code,
					Text: input.ResponseStatus.Text,
				},
				Body: brunoformat.ExampleBody{
					Type:    "json",
					Content: input.ResponseBody,
				},
			},
		},
	}

	// Initialize maps if nil
	if req.Headers == nil {
		req.Headers = make(map[string]string)
	}
	if req.QueryParams == nil {
		req.QueryParams = make(map[string]string)
	}
	if req.Example.Response.Headers == nil {
		req.Example.Response.Headers = make(map[string]string)
	}

	// Write file
	if err := s.repo.WriteFile(filePath, req); err != nil {
		return fmt.Errorf("failed to update request: %w", err)
	}

	return nil
}

// DeleteRequest deletes a request
func (s *RequestService) DeleteRequest(id string) error {
	// Decode ID to file path
	filePath := s.converter.DecodeID(id)

	// Delete file
	if err := s.repo.DeleteFile(filePath); err != nil {
		return fmt.Errorf("failed to delete request: %w", err)
	}

	return nil
}

// ListRequests returns all requests from the base directory
func (s *RequestService) ListRequests(baseDir string) ([]*dto.RequestListItem, error) {
	requests, err := s.repo.ScanDirectory(baseDir)
	if err != nil {
		return nil, err
	}

	items := make([]*dto.RequestListItem, 0, len(requests))
	for _, req := range requests {
		items = append(items, &dto.RequestListItem{
			ID:     s.converter.EncodeID(req.FilePath),
			Name:   req.Meta.Name,
			Method: req.Method,
			URL:    req.URL,
		})
	}

	return items, nil
}

// BuildRequestTree builds a folder tree from requests based on URL paths
func (s *RequestService) BuildRequestTree(baseDir string) (*dto.TreeNode, error) {
	// Load all requests
	requests, err := s.repo.ScanDirectory(baseDir)
	if err != nil {
		return nil, err
	}

	root := &dto.TreeNode{
		Name:     "root",
		Type:     "folder",
		Children: make([]*dto.TreeNode, 0),
	}

	// Group requests by URL path
	for _, req := range requests {
		// Generate unique ID from file path
		id := s.converter.EncodeID(req.FilePath)

		// Parse URL into segments
		segments := s.parseURLSegments(req.URL)

		// Insert into tree
		s.insertIntoTree(root, segments, req, id)
	}

	// Sort tree
	s.sortTree(root)

	return root, nil
}

// parseURLSegments splits a URL into segments, handling placeholders
func (s *RequestService) parseURLSegments(url string) []string {
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
func (s *RequestService) insertIntoTree(node *dto.TreeNode, segments []string, req *brunoformat.BrunoRequest, id string) {
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
		requestNode := &dto.TreeNode{
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
		segment = s.extractParamName(segment)
		isDynamic = true
	}

	var folderNode *dto.TreeNode
	for _, child := range node.Children {
		if child.Type == "folder" && child.Name == segment {
			folderNode = child
			break
		}
	}

	if folderNode == nil {
		folderNode = &dto.TreeNode{
			Name:      segment,
			Type:      "folder",
			IsDynamic: isDynamic,
			Children:  make([]*dto.TreeNode, 0),
		}
		node.Children = append(node.Children, folderNode)
	}

	// Recursively insert into subfolder
	s.insertIntoTree(folderNode, segments[1:], req, id)
}

// extractParamName extracts parameter name from :param or {param}
func (s *RequestService) extractParamName(segment string) string {
	segment = strings.TrimPrefix(segment, ":")
	segment = strings.TrimPrefix(segment, "{")
	segment = strings.TrimSuffix(segment, "}")
	return segment
}

// sortTree sorts the tree nodes alphabetically (folders first, then requests)
func (s *RequestService) sortTree(node *dto.TreeNode) {
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
		s.sortTree(child)
	}
}
