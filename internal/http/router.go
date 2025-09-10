package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Server struct {
	Router *chi.Mux
	Log    *zap.Logger
}

func NewServer(log *zap.Logger) *Server {
	r := chi.NewRouter()
	r.Use(LoggingMiddleware(log))

	return &Server{Router: r, Log: log}
}

func (s *Server) RegisterRoutes(h Handlers) {
	s.Router.Route("/api/v1", func(r chi.Router) {
		r.Route("/subscriptions", func(r chi.Router) {
			r.Get("/total", h.Total)
			r.Post("/", h.Create)
			r.Get("/", h.List)
			r.Get("/{id}", h.GetByID)
			r.Put("/{id}", h.Update)
			r.Delete("/{id}", h.Delete)
		})
	})
}

type Handlers interface {
	Create(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	Total(w http.ResponseWriter, r *http.Request)
}


