package mockserver

import (
	"log"

	"github.com/anu-mdl/linker-bruno/internal/modules/mockserver/repository"
	"github.com/anu-mdl/linker-bruno/internal/modules/mockserver/service"
	"github.com/anu-mdl/linker-bruno/internal/shared/brunoformat"
	"github.com/anu-mdl/linker-bruno/internal/shared/urlutil"
	"github.com/go-chi/chi/v5"
)

// Module represents the mock server module with all its dependencies
type Module struct {
	baseDir  string
	envName  string
	service  *service.MockService
	repo     *repository.BruRepository
	requests []*brunoformat.BrunoRequest
	envVars  map[string]string
}

// NewModule creates and initializes a new mock server module
func NewModule(baseDir, envName string) (*Module, error) {
	// Initialize dependencies
	converter := urlutil.NewConverter()

	// Create repository
	repo := repository.NewBruRepository()

	// Load environment variables
	envVars, err := repo.LoadEnvironment(envName, baseDir)
	if err != nil {
		// Non-fatal error, continue with empty env vars
		log.Printf("Warning: failed to load environment: %v", err)
		envVars = make(map[string]string)
	} else {
		log.Printf("Loaded %d environment variables", len(envVars))
	}

	// Load all .bru requests
	log.Printf("Scanning for .bru files...")
	requests, err := repo.LoadAllRequests(baseDir)
	if err != nil {
		return nil, err
	}

	log.Printf("Found %d Bruno requests", len(requests))

	if len(requests) == 0 {
		log.Println("Warning: No valid .bru files found with response blocks")
		log.Println("Make sure your .bru files contain:")
		log.Println("  1. An HTTP method block (get, post, put, delete, patch)")
		log.Println("  2. A URL in the method block")
		log.Println("  3. An example block with mock response data")
	}

	// Create service
	mockService := service.NewMockService(converter)

	return &Module{
		baseDir:  baseDir,
		envName:  envName,
		service:  mockService,
		repo:     repo,
		requests: requests,
		envVars:  envVars,
	}, nil
}

// RegisterRoutes registers all mock routes on the provided router
func (m *Module) RegisterRoutes(router chi.Router) error {
	return m.service.RegisterRoutes(router.(*chi.Mux), m.requests, m.envVars)
}
