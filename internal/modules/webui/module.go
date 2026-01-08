package webui

import (
	"html/template"
	"path/filepath"

	"github.com/anu-mdl/linker-bruno/internal/modules/webui/delivery"
	"github.com/anu-mdl/linker-bruno/internal/modules/webui/repository"
	"github.com/anu-mdl/linker-bruno/internal/modules/webui/service"
	"github.com/anu-mdl/linker-bruno/internal/shared/brunoformat"
	"github.com/anu-mdl/linker-bruno/internal/shared/urlutil"
	"github.com/go-chi/chi/v5"
)

// Module represents the web UI module with all its dependencies
type Module struct {
	baseDir    string
	uiHandler  *delivery.UIHandler
	apiHandler *delivery.APIHandler
	service    *service.RequestService
	repo       *repository.FileRepository
}

// NewModule creates and initializes a new web UI module
func NewModule(baseDir string) (*Module, error) {
	// Create custom template functions
	converter := urlutil.NewConverter()
	funcMap := template.FuncMap{
		"displayURL": func(url string) string {
			return converter.DisplayURL(url)
		},
		"displayParam": func(param string) string {
			return converter.DisplayParam(param)
		},
	}

	// Load templates with custom functions
	templates, err := template.New("").Funcs(funcMap).ParseGlob(filepath.Join("server/templates", "*.html"))
	if err != nil {
		return nil, err
	}

	// Initialize dependencies
	serializer := brunoformat.NewSerializer()

	// Create repository
	fileRepo := repository.NewFileRepository(serializer)

	// Create service
	requestService := service.NewRequestService(fileRepo, converter)

	// Create handlers
	uiHandler := delivery.NewUIHandler(templates, requestService)
	apiHandler := delivery.NewAPIHandler(requestService, templates, baseDir)

	return &Module{
		baseDir:    baseDir,
		uiHandler:  uiHandler,
		apiHandler: apiHandler,
		service:    requestService,
		repo:       fileRepo,
	}, nil
}

// RegisterRoutes registers all web UI routes on the provided router
func (m *Module) RegisterRoutes(router chi.Router) {
	m.uiHandler.RegisterRoutes(router)
	m.apiHandler.RegisterRoutes(router)
}
