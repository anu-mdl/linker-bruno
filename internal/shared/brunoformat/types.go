package brunoformat

// BrunoRequest represents a parsed .bru file
type BrunoRequest struct {
	FilePath    string
	Meta        MetaBlock
	Method      string // GET, POST, PUT, DELETE, PATCH
	URL         string
	Headers     map[string]string
	QueryParams map[string]string
	Body        string
	Example     ExampleBlock
}

// MetaBlock contains metadata
type MetaBlock struct {
	Name string
	Type string
	Seq  int
}

// ExampleBlock contains the example response definition
type ExampleBlock struct {
	Name        string
	Description string
	Request     ExampleRequest
	Response    ExampleResponse
}

// ExampleRequest contains request details in the example block
type ExampleRequest struct {
	URL    string
	Method string
	Mode   string
}

// ExampleResponse contains response details in the example block
type ExampleResponse struct {
	Headers map[string]string
	Status  ExampleStatus
	Body    ExampleBody
}

// ExampleStatus contains HTTP status information
type ExampleStatus struct {
	Code int
	Text string
}

// ExampleBody contains response body information
type ExampleBody struct {
	Type    string
	Content string
}

// NewDefaultExampleBlock creates a default example block for requests without one
func NewDefaultExampleBlock(method, url string) ExampleBlock {
	return ExampleBlock{
		Name:        "Default Response",
		Description: "Auto-generated default response",
		Request: ExampleRequest{
			URL:    url,
			Method: method,
			Mode:   "none",
		},
		Response: ExampleResponse{
			Headers: map[string]string{
				"content-type": "application/json",
			},
			Status: ExampleStatus{
				Code: 200,
				Text: "OK",
			},
			Body: ExampleBody{
				Type:    "json",
				Content: "{}",
			},
		},
	}
}
