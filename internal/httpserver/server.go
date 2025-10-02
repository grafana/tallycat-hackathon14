package httpserver

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/tallycat/tallycat/internal/httpserver/api"
	"github.com/tallycat/tallycat/internal/repository"
	"github.com/tallycat/tallycat/ui"
)

type Server struct {
	httpServer  *http.Server
	schemaRepo  repository.TelemetrySchemaRepository
	historyRepo repository.TelemetryHistoryRepository
}

func New(
	addr string,
	schemaRepo repository.TelemetrySchemaRepository,
	historyRepo repository.TelemetryHistoryRepository,
) *Server {
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
		schemaRepo:  schemaRepo,
		historyRepo: historyRepo,
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
			r.Get("/{key}/history", api.HandleTelemetryHistory(srv.historyRepo))
			r.Get("/{key}/entities", api.HandleTelemetryEntityList(srv.schemaRepo))
			r.Get("/{key}/scopes", api.HandleTelemetryScopeList(srv.schemaRepo))
			r.Route("/{key}/schemas", func(r chi.Router) {
				r.Get("/", api.HandleTelemetrySchemas(srv.schemaRepo))
				r.Get("/{schemaId}/weaver-schema.zip", api.HandleWeaverSchemaExport(srv.schemaRepo))
				r.Post("/{schemaId}", api.HandleTelemetrySchemaVersionAssignment(srv.schemaRepo, srv.historyRepo))
				r.Get("/{schemaId}", api.HandleGetTelemetrySchema(srv.schemaRepo))
			})
		})
		r.Route("/entities", func(r chi.Router) {
			r.Get("/", api.HandleEntityList(srv.schemaRepo))
			r.Get("/{entityType}/weaver-schema.zip", api.HandleEntityWeaverSchemaExport(srv.schemaRepo))
			r.Get("/{entityType}/dashboards", api.HandleEntityDashboardExport(srv.schemaRepo))
			r.Get("/{entityType}/{type}", api.HandleEntitySchemaExport(srv.schemaRepo))
		})
		r.Route("/scopes", func(r chi.Router) {
			r.Get("/", api.HandleScopeList(srv.schemaRepo))
			r.Get("/{scope}/weaver-schema.zip", api.HandleScopeWeaverSchemaExport(srv.schemaRepo))
			r.Get("/{scope}/dashboards", api.HandleScopeDashboardExport(srv.schemaRepo))
			r.Get("/{scope}/{type}", api.HandleScopeSchemaExport(srv.schemaRepo))
		})
	})
	r.Handle("/*", SPAHandler())
}

func (s *Server) Start() error {
	hostname, _ := os.Hostname()
	slog.Info("Starting HTTP server", "addr", s.httpServer.Addr, "hostname", hostname)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func SPAHandler() http.HandlerFunc {
	spaFS, err := fs.Sub(ui.DistFiles, "dist")
	if err != nil {
		panic(fmt.Errorf("failed getting the sub tree for the site files: %w", err))
	}
	return func(w http.ResponseWriter, r *http.Request) {
		f, err := spaFS.Open(strings.TrimPrefix(path.Clean(r.URL.Path), "/"))
		if err == nil {
			defer f.Close()
		}
		if os.IsNotExist(err) {
			r.URL.Path = "/"
		}
		http.FileServer(http.FS(spaFS)).ServeHTTP(w, r)
	}
}
