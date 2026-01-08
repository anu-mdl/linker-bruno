# Bruno Mock Server

A Go server that automatically generates HTTP routes from Bruno request definitions (`.bru` files). Use your Bruno collections as a single source of truth for mocking APIs!

## Features

- 🚀 Automatically scans and loads all `.bru` files
- 🔄 Generates HTTP routes dynamically based on Bruno requests
- 📝 Returns mock JSON responses from inline `example` blocks
- 🤖 **Auto-generates default responses** for requests without example blocks
- ✅ Single-file format with request and response in one place
- 🎯 Supports path parameters with variable interpolation
- 🌐 Works with both JSON objects and arrays
- 🔧 Clean separation of concerns with example block format
- ⚡ Lightweight and fast with chi router
- 🎨 **Web UI** for visual API design and management (HTMX-based)
- 📁 Nested folder structure with automatic organization by URL paths
- ✏️ Full CRUD operations for requests via web interface
- 🔤 Tab key support and undo/redo in JSON editors

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
go run cmd/app/main.go
```

Or customize with flags:
```bash
go run cmd/app/main.go --port 3000 --dir requests --env local
```

**Available Flags:**
- `--port` - Port to run the server on (default: 8080)
- `--dir` - Directory containing Bruno collection (default: current directory)
- `--env` - Environment name to load (default: "local")
- `--ui` - Enable web UI for API design and management (default: false)

### Web UI

Enable the visual web interface to create, edit, and manage your API requests:

```bash
go run cmd/app/main.go --ui --port 8080
```

Then open your browser to `http://localhost:8080/`

**Web UI Features:**
- **Three-Panel Layout**:
  - Left sidebar with folder tree (auto-organized by URL paths)
  - Center editor with tabs (General, Headers, Params, Body)
  - Right panel for response configuration
- **Visual Editing**: Edit requests, headers, query params, and response bodies
- **Dynamic Parameters**: URL parameters like `{id}` are displayed as `[id]`
- **Nested Folders**: Automatic folder hierarchy with visual indentation
- **JSON Editor**: Tab key for indentation, Ctrl+Z/Ctrl+Y for undo/redo
- **HTMX-Powered**: Partial page updates without full reloads

### Building

Build a standalone binary:
```bash
go build -o bruno-mock-server cmd/app/main.go
./bruno-mock-server --port 8080
```

## Bruno Request Format

The server uses `.bru` files with an inline `example` block for mock responses:

### File Structure

Each request is defined in a single `.bru` file:

```
requests/
├── User/
│   └── Get User.bru              # Bruno request with example response
```

### Complete .bru File Example (Get User.bru)

```bru
meta {
  name: Get User
  type: http
  seq: 1
}

get {
  url: {{baseUrl}}/users/{id}
}

headers {
  Authorization: Bearer {{token}}
  Accept: application/json
}

params:query {
  include: profile
  fields: id,username,email
}

body:json {
  {
    "requestId": "req-123"
  }
}

example {
  name: User Response Example
  description: Mock response for user retrieval

  request: {
    url: /users/{id}
    method: GET
    mode: none
  }

  response: {
    headers: {
      content-type: application/json
      x-custom-header: custom-value
    }

    status: {
      code: 200
      text: OK
    }

    body: {
      type: json
      content: '''
      {
        "id": "{{id}}",
        "username": "johndoe",
        "email": "john@example.com",
        "createdAt": "2024-01-15T10:30:00Z"
      }
      '''
    }
  }
}
```

**Note**: The `headers`, `params:query`, and `body:json` blocks are optional and primarily used when editing via the Web UI. The `example` block is required for the mock server to function.

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

example {
  name: Users List Example
  description: Mock response for listing users

  request: {
    url: /users
    method: GET
    mode: none
  }

  response: {
    status: {
      code: 200
      text: OK
    }

    body: {
      type: json
      content: '''
      [
        {
          "id": "1",
          "username": "alice"
        },
        {
          "id": "2",
          "username": "bob"
        }
      ]
      '''
    }
  }
}
```

## Example Block Format

The `example` block in `.bru` files defines the mock response:

### Example Block Structure

```bru
example {
  name: Example Name (optional)
  description: Example description (optional)

  request: {
    url: /path/to/resource
    method: GET
    mode: none
  }

  response: {
    headers: {
      header-name: header-value
    }

    status: {
      code: 200
      text: OK
    }

    body: {
      type: json
      content: '''
      {
        "data": "value"
      }
      '''
    }
  }
}
```

### Status Codes

Specify the HTTP status code and text:

```bru
status: {
  code: 200
  text: OK
}
```

Common status codes:
- `200 OK` - Success
- `201 Created` - Resource created
- `400 Bad Request` - Client error
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

### Custom Headers

Define response headers in the headers block:

```bru
headers: {
  content-type: application/json
  x-api-version: v1
  cache-control: no-cache
}
```

### Response Body

The body content must be wrapped in triple quotes (`'''`):

**Object Response:**
```bru
body: {
  type: json
  content: '''
  {
    "key": "value",
    "nested": {
      "field": true
    }
  }
  '''
}
```

**Array Response:**
```bru
body: {
  type: json
  content: '''
  [
    {"id": 1, "name": "Item 1"},
    {"id": 2, "name": "Item 2"}
  ]
  '''
}
```

## Default Responses for Requests Without Examples

The server automatically works with **any valid Bruno request**, even if it doesn't have an `example` block. This makes it compatible with:
- ✅ Brand new requests created in Bruno that haven't been executed yet
- ✅ Collections where no one has saved example responses
- ✅ Requests imported from other tools (Postman, Insomnia, etc.)

When a `.bru` file is missing an `example` block, the server automatically generates a default mock response:

**Input (minimal .bru file):**
```bru
meta {
  name: My API Request
  type: http
  seq: 1
}

get {
  url: /api/users
}
```

**Auto-generated response:**
- Status: `200 OK`
- Content-Type: `application/json`
- Body: `{}`

You can later add custom responses by:
1. Executing the request in Bruno desktop app and saving the response as an example
2. Using the Web UI to edit the response
3. Manually adding an `example` block to the `.bru` file

## Path Parameter Interpolation

Path parameters in the URL are automatically interpolated into the response body. Both `:param` and `{param}` syntax are supported.

**Get Post.bru:**
```bru
meta {
  name: Get Post
  type: http
  seq: 1
}

get {
  url: /users/{userId}/posts/{postId}
}

example {
  name: Post Example

  request: {
    url: /users/{userId}/posts/{postId}
    method: GET
    mode: none
  }

  response: {
    status: {
      code: 200
      text: OK
    }

    body: {
      type: json
      content: '''
      {
        "userId": "{{userId}}",
        "postId": "{{postId}}",
        "title": "Sample Post"
      }
      '''
    }
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

**In the Web UI**, dynamic parameters are displayed with brackets for clarity: `/users/[userId]/posts/[postId]`

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
├── cmd/
│   └── app/
│       └── main.go                    # Application entry point with DI
├── internal/                          # Internal packages (not importable externally)
│   ├── modules/                       # Business logic modules (vertical slices)
│   │   ├── mockserver/               # Mock endpoint serving module
│   │   │   ├── repository/           # .bru file loading & environment parsing
│   │   │   ├── service/              # Route registration & response interpolation
│   │   │   └── module.go             # Module initialization
│   │   └── webui/                    # Web UI module
│   │       ├── dto/                  # Request/response structures
│   │       ├── repository/           # File I/O operations
│   │       ├── service/              # CRUD orchestration & tree building
│   │       ├── delivery/             # HTTP handlers (UI + API)
│   │       └── module.go             # Module initialization
│   └── shared/                       # Shared infrastructure (Shared Kernel)
│       ├── brunoformat/              # .bru parsing & serialization
│       ├── urlutil/                  # URL conversion utilities
│       ├── response/                 # Unified API response format
│       ├── middleware/               # HTTP middleware
│       └── logger/                   # Logging configuration
├── environments/                      # Environment variables
│   └── local.bru
├── requests/                          # Bruno requests with examples
│   ├── api/
│   │   ├── users/
│   │   │   ├── Get User.bru          # Request + response example
│   │   │   └── List Users.bru
│   │   └── v1/
│   │       └── products/
│   │           └── categories/
│   │               └── Get Category.bru
│   └── ...
└── server/                            # Legacy directory (DEPRECATED)
    ├── templates/                     # HTML templates
    │   ├── index.html                 # Main page layout
    │   ├── sidebar.html               # Folder tree sidebar
    │   └── editor.html                # Request editor with tabs
    └── static/                        # Static assets
        └── style.css                  # UI styles
```

**Architecture**: The project follows a modular, vertically-sliced architecture with clear separation of concerns:
- **Modules** contain business logic organized by feature (mockserver, webui)
- **Shared** contains infrastructure utilities used across modules
- Each module has **repository** (data access), **service** (business logic), and **delivery** (HTTP) layers
- Dependency injection is handled in each module's `module.go` file

## How It Works

### Mock Server Mode (Default)
1. **Scanning**: Server recursively scans the directory for `.bru` files
2. **Parsing**: Each `.bru` file is parsed to extract:
   - HTTP method (GET, POST, PUT, DELETE, PATCH)
   - URL pattern
   - Request metadata (headers, params, body)
   - Example block with response definition
3. **Route Generation**: URLs are converted to chi routes (e.g., `/users/{id}`)
4. **Variable Interpolation**: Path parameters are extracted and interpolated into response bodies
5. **Serving**: HTTP server responds with the mock data from the example block

### Web UI Mode (--ui flag)
1. All of the above, plus:
2. **Tree Building**: Requests are organized into a nested folder structure based on URL paths
3. **Template Rendering**: HTMX-based interface with three panels (sidebar, editor, response)
4. **CRUD Operations**: Create, read, update, and delete requests via the web interface
5. **File Serialization**: Changes are written back to `.bru` files maintaining the format

## Supported HTTP Methods

- GET
- POST
- PUT
- DELETE
- PATCH

## Limitations

- Only JSON responses are currently supported (no XML, plain text, etc.)
- No request body validation
- All `.bru` files must contain a valid `example` block with response definition
- Web UI requires JavaScript enabled (uses HTMX for dynamic updates)

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
