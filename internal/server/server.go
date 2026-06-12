package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/infralens/infralens/internal/server/handler"
)

type Server struct {
	router *chi.Mux
	port   string
}

func New(projectHandler *handler.ProjectHandler, crawlHandler *handler.CrawlHandler, analyticsHandler *handler.AnalyticsHandler, port string) *Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/projects", func(r chi.Router) {
			r.Get("/", projectHandler.List)
			r.Get("/{id}", projectHandler.Get)
			r.Get("/{id}/changes", projectHandler.Changes)
		})
		r.Get("/crawls", crawlHandler.List)
		r.Route("/analytics", func(r chi.Router) {
			r.Get("/status-distribution", analyticsHandler.StatusDistribution)
			r.Get("/top-builders", analyticsHandler.TopBuilders)
			r.Get("/by-district", analyticsHandler.ByDistrict)
		})
	})

	// Swagger UI
	r.Get("/docs", handler.ServeSwaggerUI)
	r.Get("/docs/openapi.yaml", handler.ServeSpec)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	return &Server{router: r, port: port}
}

func (s *Server) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:         ":" + s.port,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		log.Println("shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Printf("server listening on :%s", s.port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server: %w", err)
	}
	return nil
}
