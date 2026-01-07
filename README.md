# Bruno Mock Server

A Go server that automatically generates HTTP routes from Bruno request definitions (`.bru` files). Use your Bruno collections as a single source of truth for mocking APIs!

## Features

- 🚀 Automatically scans and loads all `.bru` files
- 🔄 Generates HTTP routes dynamically based on Bruno requests
- 📝 Returns mock JSON responses from inline `example` blocks
- ✅ Single-file format with request and response in one place
- 🎯 Supports path parameters with variable interpolation
- 🌐 Works with both JSON objects and arrays
- 🔧 Clean separation of concerns with example block format
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
  url: {{baseUrl}}/users/:id
}

example {
  name: User Response Example
  description: Mock response for user retrieval

  request: {
    url: /users/:id
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

example {
  name: Post Example

  request: {
    url: /users/:userId/posts/:postId
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
├── requests/                          # Bruno requests with examples
│   ├── User/
│   │   ├── User Info.bru              # Request + response example
│   │   └── User Repos.bru             # Request + response example
│   └── Repository/
│       ├── Repository Info.bru
│       └── ...
└── server/                            # Go server code
    ├── main.go                        # Entry point
    ├── parser/
    │   ├── types.go                   # Data structures
    │   └── parser.go                  # .bru file parser
    ├── loader/
    │   ├── loader.go                  # File loader
    │   └── environment.go             # Environment loader
    └── router/
        └── router.go                  # Route generator
```

## How It Works

1. **Scanning**: Server recursively scans the directory for `.bru` files
2. **Parsing**: Each `.bru` file is parsed to extract:
   - HTTP method (GET, POST, PUT, DELETE, PATCH)
   - URL pattern
   - Example block with response definition
3. **Route Generation**: URLs are converted to chi routes (e.g., `/users/:id` → `/users/{id}`)
4. **Variable Interpolation**: Path parameters are extracted and interpolated into response bodies
5. **Serving**: HTTP server responds with the mock data from the example block

## Supported HTTP Methods

- GET
- POST
- PUT
- DELETE
- PATCH

## Limitations

- Only JSON responses are supported (no XML, plain text, etc.)
- No request body validation (yet)
- All `.bru` files must contain a valid `example` block with response definition

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
