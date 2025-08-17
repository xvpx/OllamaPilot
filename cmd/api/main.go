package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chat_ollama/internal/api"
	"chat_ollama/internal/config"
	"chat_ollama/internal/database"
	"chat_ollama/internal/utils"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := utils.NewLogger(cfg.LogLevel, cfg.LogFormat)
	logger.Info().
		Str("version", "1.0.0").
		Str("environment", cfg.Environment).
		Str("port", cfg.Port).
		Msg("Starting Chat Ollama API server")

	// Initialize database
	db, err := database.NewDatabase(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error().Err(err).Msg("Failed to close database")
		}
	}()

	if cfg.IsPostgreSQL() {
		logger.Info().
			Str("db_type", "postgresql").
			Str("db_host", cfg.DBHost).
			Str("db_name", cfg.DBName).
			Msg("Database initialized")
	} else {
		logger.Info().
			Str("db_type", "sqlite").
			Str("db_path", cfg.DBPath).
			Msg("Database initialized")
	}

	// Run database migrations
	if err := database.RunMigrations(db, cfg); err != nil {
		logger.Fatal().Err(err).Msg("Failed to run database migrations")
	}

	logger.Info().Msg("Database migrations completed")

	// Initialize router
	router := api.NewRouter(db, cfg, logger)
	handler := router.GetHandler()

	// Configure HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info().
			Str("address", server.Addr).
			Msg("Starting HTTP server")
		
		serverErrors <- server.ListenAndServe()
	}()

	// Wait for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Fatal().Err(err).Msg("Server failed to start")

	case sig := <-shutdown:
		logger.Info().
			Str("signal", sig.String()).
			Msg("Shutdown signal received")

		// Give outstanding requests a deadline for completion
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(ctx); err != nil {
			logger.Error().Err(err).Msg("Graceful shutdown failed")
			
			// Force shutdown
			if err := server.Close(); err != nil {
				logger.Error().Err(err).Msg("Failed to force close server")
			}
		}

		logger.Info().Msg("Server shutdown completed")
	}
}