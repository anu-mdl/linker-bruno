package dto

// TreeNode represents a folder or request in the UI sidebar tree
type TreeNode struct {
	Name      string      `json:"name"`
	Type      string      `json:"type"` // "folder" or "request"
	Method    string      `json:"method,omitempty"` // HTTP method (for requests only)
	URL       string      `json:"url,omitempty"` // Full URL path
	ID        string      `json:"id,omitempty"` // Unique identifier for the request
	IsDynamic bool        `json:"isDynamic"` // True if this is a dynamic parameter folder/segment
	Children  []*TreeNode `json:"children,omitempty"`
}

// CreateRequestInput represents input for creating a new request
type CreateRequestInput struct {
	Name        string            `json:"name"`
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Description string            `json:"description,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	QueryParams map[string]string `json:"queryParams,omitempty"`
	Body        string            `json:"body,omitempty"`
	ResponseStatus struct {
		Code int    `json:"code"`
		Text string `json:"text"`
	} `json:"responseStatus"`
	ResponseHeaders map[string]string `json:"responseHeaders,omitempty"`
	ResponseBody    string            `json:"responseBody,omitempty"`
}

// UpdateRequestInput represents input for updating an existing request
type UpdateRequestInput struct {
	Name        string            `json:"name"`
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Description string            `json:"description,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	QueryParams map[string]string `json:"queryParams,omitempty"`
	Body        string            `json:"body,omitempty"`
	ResponseStatus struct {
		Code int    `json:"code"`
		Text string `json:"text"`
	} `json:"responseStatus"`
	ResponseHeaders map[string]string `json:"responseHeaders,omitempty"`
	ResponseBody    string            `json:"responseBody,omitempty"`
}

// RequestResponse represents a request returned to the client
type RequestResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Description string            `json:"description,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	QueryParams map[string]string `json:"queryParams,omitempty"`
	Body        string            `json:"body,omitempty"`
	ResponseStatus struct {
		Code int    `json:"code"`
		Text string `json:"text"`
	} `json:"responseStatus"`
	ResponseHeaders map[string]string `json:"responseHeaders,omitempty"`
	ResponseBody    string            `json:"responseBody,omitempty"`
}

// RequestListItem represents a request in the list/tree view
type RequestListItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Method string `json:"method"`
	URL    string `json:"url"`
}
