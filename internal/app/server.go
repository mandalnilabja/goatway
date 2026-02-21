package app

import (
	"log"
	"net/http"
	"time"

	"github.com/mandalnilabja/goatway/internal/config"
)

// Server wraps the HTTP server with its configuration
type Server struct {
	httpServer *http.Server
	config     *config.Config
}

// NewServer creates a new configured HTTP server instance
func NewServer(cfg *config.Config, handler http.Handler) *Server {
	srv := &http.Server{
		Addr:    cfg.ServerPort,
		Handler: handler,
		// IMPORTANT: ReadTimeout can kill long streams!
		// For LLM streaming responses, we need generous timeouts
		ReadTimeout:  300 * time.Second,
		WriteTimeout: 300 * time.Second,
	}

	return &Server{
		httpServer: srv,
		config:     cfg,
	}
}

// Start begins listening and serving HTTP requests
func (s *Server) Start() error {
	log.Printf("Goatway server starting on http://localhost%s", s.config.ServerPort)

	if err := s.httpServer.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
