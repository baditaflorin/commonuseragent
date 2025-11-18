package web

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed templates/*
var templates embed.FS

// Server handles web UI requests
type Server struct {
	tmpl *template.Template
}

// NewServer creates a new web server
func NewServer() (*Server, error) {
	tmpl, err := template.ParseFS(templates, "templates/*.html")
	if err != nil {
		return nil, err
	}

	return &Server{
		tmpl: tmpl,
	}, nil
}

// Index serves the main web interface
func (s *Server) Index(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';")

	if err := s.tmpl.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
