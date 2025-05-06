/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tallycat/tallycat/internal/grpcserver"
	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

var (
	serverAddr           string
	maxConcurrentStreams uint32
	connectionTimeout    time.Duration
	shutdownTimeout      time.Duration
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

		slog.Info("Starting OpenTelemetry logs collector server",
			"addr", serverAddr,
			"maxConcurrentStreams", maxConcurrentStreams,
			"connectionTimeout", connectionTimeout,
			"shutdownTimeout", shutdownTimeout,
		)

		opts := []grpc.ServerOption{
			grpc.MaxConcurrentStreams(maxConcurrentStreams),
			grpc.ConnectionTimeout(connectionTimeout),
		}

		srv := grpcserver.NewServer(serverAddr, opts...)

		logsService := grpcserver.NewLogsServiceServer()
		srv.RegisterService(&logspb.LogsService_ServiceDesc, logsService)

		g, _ := errgroup.WithContext(ctx)

		g.Go(func() error {
			return srv.Start()
		})

		g.Go(func() error {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

			for sig := range sigChan {
				switch sig {
				case syscall.SIGTERM, syscall.SIGINT:
					slog.Info("Received shutdown signal", "signal", sig)
					cancel()
				case syscall.SIGHUP:
					slog.Info("Received reload signal", "signal", sig)
					// TODO: Implement configuration reload
				}
			}
			return nil
		})

		if err := g.Wait(); err != nil {
			return err
		}

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()

		shutdownDone := make(chan struct{})
		go func() {
			defer close(shutdownDone)
			srv.Stop()
		}()

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
	serverCmd.Flags().StringVarP(&serverAddr, "addr", "a", ":4317", "Address to listen on (default: :4317)")
	serverCmd.Flags().Uint32Var(&maxConcurrentStreams, "max-streams", 1000, "Maximum number of concurrent streams")
	serverCmd.Flags().DurationVar(&connectionTimeout, "connection-timeout", 10*time.Second, "Connection timeout duration")
	serverCmd.Flags().DurationVar(&shutdownTimeout, "shutdown-timeout", 30*time.Second, "Graceful shutdown timeout duration")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
