package server

import (
	"fmt"
	"net/http"

	"github.com/astgot/forum/internal/controller"
)

// Server ..
type Server struct {
	config *Config
	mux    *controller.Multiplexer
}

// New - generates instance to support service
func New(config *Config) *Server {
	return &Server{
		config: config,
		mux:    controller.NewMux(),
	}
}

// Start - Initializing server
func (s *Server) Start() error {
	s.mux.CreateHandlers()
	if err := s.mux.ConfigureDB(); err != nil {
		return err
	}
	fmt.Println("Server is working on port", s.config.WebPort)
	return http.ListenAndServe(s.config.WebPort, s.mux.Mux)
}
