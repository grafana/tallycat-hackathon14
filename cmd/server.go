/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tallycat/tallycat/internal/grpcserver"
	"github.com/tallycat/tallycat/internal/httpserver"
	"github.com/tallycat/tallycat/internal/repository/duckdb"
	"github.com/tallycat/tallycat/internal/repository/duckdb/migrator"
	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	profilespb "go.opentelemetry.io/proto/otlp/collector/profiles/v1development"
	tracespb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

var (
	grpcAddr             string
	maxConcurrentStreams uint32
	connectionTimeout    time.Duration
	shutdownTimeout      time.Duration
	httpAddr             string
	databasePath         string
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the OpenTelemetry logs collector server",
	Long: `Start the OpenTelemetry logs collector server that implements the
OpenTelemetry LogsService interface. The server listens for gRPC connections
and processes log data according to the OpenTelemetry protocol.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		slog.SetDefault(logger)

		slog.Info("Starting OpenTelemetry logs collector server",
			"grpcAddr", grpcAddr,
			"httpAddr", httpAddr,
			"maxConcurrentStreams", maxConcurrentStreams,
			"connectionTimeout", connectionTimeout,
			"shutdownTimeout", shutdownTimeout,
		)

		opts := []grpc.ServerOption{
			grpc.MaxConcurrentStreams(maxConcurrentStreams),
			grpc.ConnectionTimeout(connectionTimeout),
		}

		srv := grpcserver.NewServer(grpcAddr, opts...)

		pool, err := duckdb.NewConnectionPool(&duckdb.Config{
			DatabasePath:    databasePath,
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: time.Minute * 5,
		}, logger)

		if err != nil {
			return fmt.Errorf("failed to create connection pool: %w", err)
		}

		schemaRepo := duckdb.NewTelemetrySchemaRepository(pool.(*duckdb.ConnectionPool))
		historyRepo := duckdb.NewTelemetryHistoryRepository(pool.(*duckdb.ConnectionPool))

		// Run migrations using the pool connection
		db := pool.GetConnection()
		if err := migrator.ApplyMigrations(db); err != nil {
			slog.Error("failed to run migrations", "error", err)
		}

		logsService := grpcserver.NewLogsServiceServer(schemaRepo)
		srv.RegisterService(&logspb.LogsService_ServiceDesc, logsService)

		metricsService := grpcserver.NewMetricsServiceServer(schemaRepo)
		srv.RegisterService(&metricspb.MetricsService_ServiceDesc, metricsService)

		tracesService := grpcserver.NewTracesServiceServer(schemaRepo)
		srv.RegisterService(&tracespb.TraceService_ServiceDesc, tracesService)

		profilesService := grpcserver.NewProfilesServiceServer(schemaRepo)
		srv.RegisterService(&profilespb.ProfilesService_ServiceDesc, profilesService)

		httpSrv := httpserver.New(httpAddr, schemaRepo, historyRepo)

		g, _ := errgroup.WithContext(ctx)

		g.Go(func() error {
			if err := srv.Start(); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
				return err
			}
			return nil
		})

		g.Go(func() error {
			if err := httpSrv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		})

		func() {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

			for sig := range sigChan {
				switch sig {
				case syscall.SIGTERM, syscall.SIGINT:
					slog.Info("Received shutdown signal", "signal", sig)
					pool.Close()
					cancel()
					return
				case syscall.SIGHUP:
					slog.Info("Received reload signal", "signal", sig)
					// TODO: Implement configuration reload
				}
			}
		}()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()

		shutdownDone := make(chan struct{})
		go func() {
			defer close(shutdownDone)
			srv.Stop()
			httpSrv.Shutdown(shutdownCtx)
		}()

		if err := g.Wait(); err != nil {
			return err
		}

		select {
		case <-shutdownDone:
			slog.Info("Server stopped gracefully")
		case <-shutdownCtx.Done():
			slog.Warn("Server shutdown timed out, forcing stop")
			srv.ForceStop()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.
	serverCmd.Flags().StringVarP(&grpcAddr, "grpc-addr", "g", ":4317", "Address to listen on for gRPC server (default: :4317)")
	serverCmd.Flags().Uint32Var(&maxConcurrentStreams, "max-streams", 1000, "Maximum number of concurrent streams")
	serverCmd.Flags().DurationVar(&connectionTimeout, "connection-timeout", 10*time.Second, "Connection timeout duration")
	serverCmd.Flags().DurationVar(&shutdownTimeout, "shutdown-timeout", 30*time.Second, "Graceful shutdown timeout duration")
	serverCmd.Flags().StringVarP(&httpAddr, "http-addr", "H", ":8080", "Address to listen on for HTTP server (default: :8080)")
	serverCmd.Flags().StringVarP(&databasePath, "database-path", "d", "tallycat.db", "Path to the database file")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
