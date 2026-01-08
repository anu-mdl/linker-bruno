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
