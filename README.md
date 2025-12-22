# Bruno Mock Server

A Go server that automatically generates HTTP routes from Bruno request definitions (`.bru` files). Use your Bruno collections as a single source of truth for mocking APIs!

## Features

- 🚀 Automatically scans and loads all `.bru` files
- 🔄 Generates HTTP routes dynamically based on Bruno requests
- 📝 Returns mock JSON responses from separate `.response.json` files
- ✅ Keeps `.bru` files fully compatible with Bruno desktop app
- 🎯 Supports path parameters with variable interpolation
- 🌐 Works with both JSON objects and arrays
- 🔧 Auto-generates default responses when `.response.json` is missing
- ⚡ Lightweight and fast with chi router

## Quick Start

### Prerequisites

- Go 1.16 or higher
- Bruno collection with `.bru` files

### Installation

1. Clone this repository:
```bash
git clone <your-repo-url>
cd linker-bruno
```

2. Install dependencies:
```bash
go mod download
```

### Running the Server

Start the server with default settings (port 8080, current directory):
```bash
go run server/main.go
```

Or customize with flags:
```bash
go run server/main.go --port 3000 --dir requests --env local
```

**Available Flags:**
- `--port` - Port to run the server on (default: 8080)
- `--dir` - Directory containing Bruno collection (default: current directory)
- `--env` - Environment name to load (default: "local")

### Building

Build a standalone binary:
```bash
go build -o bruno-mock-server server/main.go
./bruno-mock-server --port 8080
```

## Bruno Request Format

The server uses standard Bruno `.bru` files (fully compatible with Bruno desktop) and separate `.response.json` files for mock responses:

### File Structure

For each `.bru` request file, create a corresponding `.response.json` file:

```
requests/
├── User/
│   ├── Get User.bru              # Bruno request definition
│   └── Get User.response.json    # Mock response data
```

### Bruno Request File (Get User.bru)

Standard Bruno format - no modifications needed:

```bru
meta {
  name: Get User
  type: http
  seq: 1
}

get {
  url: {{baseUrl}}/users/:id
}
```

### Response File (Get User.response.json)

JSON file with status, headers, and body:

```json
{
  "status": 200,
  "headers": {
    "Content-Type": "application/json",
    "X-Custom-Header": "custom-value"
  },
  "body": {
    "id": "{{id}}",
    "username": "johndoe",
    "email": "john@example.com",
    "createdAt": "2024-01-15T10:30:00Z"
  }
}
```

### Array Response Example

**List Users.bru:**
```bru
meta {
  name: List Users
  type: http
  seq: 2
}

get {
  url: {{baseUrl}}/users
}
```

**List Users.response.json:**
```json
{
  "status": 200,
  "body": [
    {
      "id": "1",
      "username": "alice"
    },
    {
      "id": "2",
      "username": "bob"
    }
  ]
}
```

## Response JSON File Format

The `.response.json` file supports the following fields:

### Status Code
```json
{
  "status": 200
}
```
Default: 200 if not specified

### Custom Headers
```json
{
  "status": 200,
  "headers": {
    "Content-Type": "application/json",
    "X-API-Version": "v1",
    "Cache-Control": "no-cache"
  }
}
```

### Response Body
Supports both JSON objects and arrays:

**Object:**
```json
{
  "status": 200,
  "body": {
    "key": "value",
    "nested": {
      "field": true
    }
  }
}
```

**Array:**
```json
{
  "status": 200,
  "body": [
    {"id": 1, "name": "Item 1"},
    {"id": 2, "name": "Item 2"}
  ]
}
```

### Default Response

If no `.response.json` file exists, the server automatically generates a default response:

```json
{
  "message": "Mock response for [Request Name]",
  "method": "GET",
  "url": "{{baseUrl}}/path"
}
```

## Path Parameter Interpolation

Path parameters in the URL are automatically interpolated into the response body:

**Get Post.bru:**
```bru
meta {
  name: Get Post
  type: http
  seq: 1
}

get {
  url: {{baseUrl}}/users/:userId/posts/:postId
}
```

**Get Post.response.json:**
```json
{
  "status": 200,
  "body": {
    "userId": "{{userId}}",
    "postId": "{{postId}}",
    "title": "Sample Post"
  }
}
```

**HTTP Request:**
```bash
curl http://localhost:8080/users/123/posts/456
```

**Actual Response:**
```json
{
  "userId": "123",
  "postId": "456",
  "title": "Sample Post"
}
```

## Environment Variables

Environment variables can be defined in `environments/*.bru` files:

**environments/local.bru:**
```bru
vars {
  baseUrl: http://localhost:3000
  apiKey: secret-key-123
}
```

These variables are substituted in request URLs during server startup.

## Testing

Test the server with curl:

```bash
# Get user info
curl http://localhost:8080/users/usebruno

# Get user repos
curl http://localhost:8080/users/usebruno/repos

# Pretty print with jq
curl http://localhost:8080/users/usebruno | jq '.'
```

## Project Structure

```
linker-bruno/
├── bruno.json                         # Bruno collection config
├── environments/                      # Environment variables
│   └── local.bru
├── requests/                          # Bruno requests & responses
│   ├── User/
│   │   ├── User Info.bru              # Bruno request
│   │   ├── User Info.response.json    # Mock response
│   │   ├── User Repos.bru
│   │   └── User Repos.response.json
│   └── Repository/
│       ├── Repository Info.bru
│       └── ...
└── server/                            # Go server code
    ├── main.go                        # Entry point
    ├── parser/
    │   ├── types.go                   # Data structures
    │   └── parser.go                  # .bru file parser
    ├── loader/
    │   ├── loader.go                  # File & response loader
    │   └── environment.go             # Environment loader
    └── router/
        └── router.go                  # Route generator
```

## How It Works

1. **Scanning**: Server recursively scans the directory for `.bru` files
2. **Parsing**: Each `.bru` file is parsed to extract:
   - HTTP method (GET, POST, PUT, DELETE, PATCH)
   - URL pattern
3. **Response Loading**: For each `.bru` file, the server looks for a matching `.response.json` file:
   - If found, loads status, headers, and body from JSON
   - If not found, generates a default response
4. **Route Generation**: URLs are converted to chi routes (e.g., `/users/:id` → `/users/{id}`)
5. **Variable Interpolation**: Path parameters are extracted and interpolated into response bodies
6. **Serving**: HTTP server responds with the mock data from `.response.json` files

## Supported HTTP Methods

- GET
- POST
- PUT
- DELETE
- PATCH

## Limitations

- Response data stored separately from `.bru` files (not visible in Bruno desktop)
- Only JSON responses are supported (no XML, plain text, etc.)
- No request body validation (yet)
- Status codes and headers must be defined in `.response.json` files

## Example Use Cases

- **API Mocking**: Mock backend APIs for frontend development
- **Testing**: Create mock servers for integration tests
- **Prototyping**: Quickly prototype APIs without writing backend code
- **Documentation**: Use Bruno files as living API documentation with examples

## Contributing

Contributions are welcome! Feel free to:
- Add support for more response formats
- Implement request validation
- Add response delays for latency simulation
- Create conditional responses based on request data

## License

MIT

## Acknowledgments

Built with:
- [Bruno](https://www.usebruno.com/) - Open-source API client
- [chi](https://github.com/go-chi/chi) - Lightweight Go HTTP router
