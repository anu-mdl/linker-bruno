package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/anu-mdl/linker-bruno/server/loader"
	"github.com/anu-mdl/linker-bruno/server/router"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Parse CLI flags
	port := flag.Int("port", 8080, "Port to run the server on")
	dir := flag.String("dir", ".", "Directory containing Bruno collection")
	env := flag.String("env", "local", "Environment name to load")
	flag.Parse()

	log.Printf("Starting Bruno Mock Server")
	log.Printf("Directory: %s", *dir)
	log.Printf("Environment: %s", *env)

	// Load environment variables
	envVars, err := loader.LoadEnvironment(*env, *dir)
	if err != nil {
		log.Printf("Warning: failed to load environment: %v", err)
		envVars = make(map[string]string)
	} else {
		log.Printf("Loaded %d environment variables", len(envVars))
	}

	// Load all .bru files
	log.Printf("Scanning for .bru files...")
	requests, err := loader.LoadAllRequests(*dir)
	if err != nil {
		log.Fatalf("Failed to load requests: %v", err)
	}

	log.Printf("Found %d Bruno requests", len(requests))

	if len(requests) == 0 {
		log.Println("Warning: No valid .bru files found with response blocks")
		log.Println("Make sure your .bru files contain:")
		log.Println("  1. An HTTP method block (get, post, put, delete, patch)")
		log.Println("  2. A URL in the method block")
		log.Println("  3. Optionally, a response block with mock data")
	}

	// Create router and add middleware first
	r := router.CreateRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Then register routes
	router.RegisterRoutes(r, requests, envVars)

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
