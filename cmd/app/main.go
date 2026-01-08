package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/anu-mdl/linker-bruno/internal/modules/mockserver"
	"github.com/anu-mdl/linker-bruno/internal/modules/webui"
	"github.com/anu-mdl/linker-bruno/internal/shared/logger"
	"github.com/anu-mdl/linker-bruno/internal/shared/middleware"
	"github.com/go-chi/chi/v5"
)

func main() {
	// Initialize logger
	logger.Setup()

	// Parse CLI flags
	port := flag.Int("port", 8080, "Port to run the server on")
	dir := flag.String("dir", ".", "Directory containing Bruno collection")
	env := flag.String("env", "local", "Environment name to load")
	ui := flag.Bool("ui", false, "Enable web UI for API design")
	flag.Parse()

	log.Printf("Starting Bruno Mock Server")
	log.Printf("Directory: %s", *dir)
	log.Printf("Environment: %s", *env)

	// Create router
	r := chi.NewRouter()
	middleware.SetupDefault(r)

	// Initialize Web UI module (if enabled)
	if *ui {
		log.Println("Web UI enabled - initializing UI module")
		uiModule, err := webui.NewModule(*dir)
		if err != nil {
			log.Fatalf("Failed to initialize UI module: %v", err)
		}
		uiModule.RegisterRoutes(r)
		log.Printf("Web UI available at http://localhost:%d/", *port)
	}

	// Initialize Mock Server module
	mockModule, err := mockserver.NewModule(*dir, *env)
	if err != nil {
		log.Fatalf("Failed to initialize mock server module: %v", err)
	}

	if err := mockModule.RegisterRoutes(r); err != nil {
		log.Fatalf("Failed to register mock routes: %v", err)
	}

	// Add a default 404 handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"error": "Route not found", "path": "%s", "method": "%s"}`, r.URL.Path, r.Method)
	})

	// Start the server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Server listening on http://localhost%s", addr)
	log.Printf("Press Ctrl+C to stop")

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
