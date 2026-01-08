package middleware

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// SetupDefault sets up default middleware for the chi router
// Includes logging and panic recovery
func SetupDefault(r chi.Router) {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
}
