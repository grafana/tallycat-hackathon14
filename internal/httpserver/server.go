package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/tallycat/tallycat/internal/httpserver/api"
	"github.com/tallycat/tallycat/internal/repository"
)

type Server struct {
	httpServer *http.Server
	schemaRepo repository.TelemetrySchemaRepository
}

func New(addr string, schemaRepo repository.TelemetrySchemaRepository) *Server {
	r := chi.NewRouter()

	// Register middlewares
	registerMiddlewares(r)

	// Register routes
	registerHealthCheck(r)

	srv := &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: r,
		},
		schemaRepo: schemaRepo,
	}

	// Register API routes
	registerAPIRoutes(r, srv)

	return srv
}

func registerMiddlewares(r chi.Router) {
	// Middlewares must be registered before routes or mounts
	r.Use(middleware.RealIP)
	r.Use(middleware.CleanPath)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Second * 60))
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	// Now mount profiler and add routes
	r.Mount("/debug", middleware.Profiler())
}

func registerHealthCheck(r chi.Router) {
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}

func registerAPIRoutes(r chi.Router, srv *Server) {
	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/telemetries", func(r chi.Router) {
			r.Get("/", api.HandleTelemetryList(srv.schemaRepo))
			r.Get("/{key}", api.HandleGetTelemetry(srv.schemaRepo))
			r.Route("/{key}/schemas", func(r chi.Router) {
				r.Get("/", api.HandleTelemetrySchemas(srv.schemaRepo))
				r.Post("/{schemaId}", api.HandleTelemetrySchemaVersionAssignment(srv.schemaRepo))
				r.Get("/{schemaId}", api.HandleGetTelemetrySchema(srv.schemaRepo))
			})
		})
	})
}

func (s *Server) Start() error {
	hostname, _ := os.Hostname()
	slog.Info("Starting HTTP server", "addr", s.httpServer.Addr, "hostname", hostname)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
