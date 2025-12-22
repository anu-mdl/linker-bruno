package parser

// BrunoRequest represents a parsed .bru file
type BrunoRequest struct {
	FilePath string
	Meta     MetaBlock
	Method   string // GET, POST, PUT, DELETE, PATCH
	URL      string
	Response ResponseBlock
}

// MetaBlock contains metadata
type MetaBlock struct {
	Name string
	Type string
	Seq  int
}

// ResponseBlock contains mock response data
type ResponseBlock struct {
	Status  int
	Headers map[string]string
	Body    interface{} // can be JSON object, array, or string
}
