package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dfedick/gotak/internal/server"
	"github.com/dfedick/gotak/pkg/config"
	"github.com/dfedick/gotak/pkg/logger"
)

var (
	version   = "dev"
	buildDate = "unknown"
	commit    = "unknown"
)

func main() {
	var (
		configPath = flag.String("config", "config/server.yaml", "path to configuration file")
		showVersion = flag.Bool("version", false, "show version information")
		debug      = flag.Bool("debug", false, "enable debug logging")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("GoTAK Server\n")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Build Date: %s\n", buildDate)
		fmt.Printf("Commit: %s\n", commit)
		return
	}

	// Load configuration
	cfg, err := config.LoadServerConfig(*configPath)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	if *debug {
		cfg.Logging.Level = "debug"
	}

	// Initialize structured logger
	loggerConfig := logger.Config{
		Level:      cfg.Logging.Level,
		Format:     cfg.Logging.Format,
		Output:     cfg.Logging.Output,
		Service:    "gotak-server",
		Version:    version,
		TimeFormat: "rfc3339",
	}
	logger.Initialize(loggerConfig)
	log := logger.GetGlobalLogger()

	// Create server instance
	srv, err := server.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create server")
	}

	// Create context that cancels on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		log.Info().Str("version", version).Msg("Starting GoTAK Server")
		if err := srv.Start(ctx); err != nil {
			errCh <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigCh:
		log.Info().Str("signal", sig.String()).Msg("Received signal, initiating graceful shutdown")
		cancel()
		
		// Give the server time to shutdown gracefully
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("Error during shutdown")
		}
		
	case err := <-errCh:
		log.Error().Err(err).Msg("Server error occurred")
		cancel()
	}

	log.Info().Msg("GoTAK Server stopped")
}
