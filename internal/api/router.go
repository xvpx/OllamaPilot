package api

import (
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"chat_ollama/internal/api/handlers"
	apiMiddleware "chat_ollama/internal/api/middleware"
	"chat_ollama/internal/config"
	"chat_ollama/internal/database"
	"chat_ollama/internal/utils"
)

// Router holds the router configuration and dependencies
type Router struct {
	db     *database.DB
	cfg    *config.Config
	logger *utils.Logger
}

// NewRouter creates a new router instance
func NewRouter(db *database.DB, cfg *config.Config, logger *utils.Logger) *Router {
	return &Router{
		db:     db,
		cfg:    cfg,
		logger: logger,
	}
}

// SetupRoutes configures all routes and middleware
func (rt *Router) SetupRoutes() http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(apiMiddleware.LoggingMiddleware(rt.logger))
	r.Use(apiMiddleware.RecoveryMiddleware(rt.logger))
	r.Use(apiMiddleware.CORSMiddleware())
	r.Use(middleware.Heartbeat("/ping"))

	// Serve static files for the web UI
	workDir, _ := filepath.Abs("./web")
	filesDir := http.Dir(workDir)
	rt.logger.Info().Str("web_dir", workDir).Msg("Serving static files")
	
	// Serve static files
	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(filesDir)))
	
	// Serve the main UI at root
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(workDir, "index.html"))
	})
	
	// Serve individual static files
	r.Get("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(workDir, "style.css"))
	})
	r.Get("/app.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(workDir, "app.js"))
	})

	// Health check handlers
	healthHandler := handlers.NewHealthHandler(rt.db, rt.cfg, rt.logger)
	r.Get("/health", healthHandler.HealthCheck)
	r.Get("/ready", healthHandler.ReadinessCheck)
	r.Get("/live", healthHandler.LivenessCheck)

	// API v1 routes
	r.Route("/v1", func(r chi.Router) {
		// Chat handlers
		chatHandler := handlers.NewChatHandler(rt.db, rt.cfg, rt.logger)
		
		// Chat endpoints
		r.Post("/chat", chatHandler.Chat)
		
		// Session endpoints
		r.Get("/sessions", chatHandler.GetSessions)
		r.Get("/sessions/{sessionID}/messages", chatHandler.GetSessionMessages)
		r.Delete("/sessions/{sessionID}", chatHandler.DeleteSession)
		
		// Model management handlers
		modelsHandler := handlers.NewModelsHandler(rt.db, rt.cfg, rt.logger)
		
		// Model endpoints
		r.Get("/models", modelsHandler.GetModels)
		r.Get("/models/{modelID}", modelsHandler.GetModel)
		r.Put("/models/{modelID}", modelsHandler.UpdateModel)
		r.Delete("/models/{modelID}", modelsHandler.DeleteModel) // Soft delete
		r.Post("/models/sync", modelsHandler.SyncModels)
		
		// Model download endpoints
		r.Post("/models/download", modelsHandler.DownloadModel)
		r.Get("/models/available", modelsHandler.GetAvailableModels)
		r.Post("/models/available/refresh", modelsHandler.RefreshAvailableModels)
		r.Get("/models/cache-info", modelsHandler.GetCacheInfo)
		r.Get("/models/{modelID}/download-status", modelsHandler.GetModelDownloadStatus)
		
		// Model configuration endpoints
		r.Get("/models/{modelID}/config", modelsHandler.GetModelConfig)
		r.Put("/models/{modelID}/config", modelsHandler.UpdateModelConfig)
		
		// Model management endpoints
		r.Post("/models/{modelID}/default", modelsHandler.SetDefaultModel)
		r.Get("/models/{modelID}/stats", modelsHandler.GetModelStats)
		
		// Model deletion endpoints
		r.Delete("/models/{modelID}/hard", modelsHandler.HardDeleteModel) // Hard delete
		r.Post("/models/{modelID}/restore", modelsHandler.RestoreModel)   // Restore soft-deleted model
	})

	// Add a catch-all route for undefined endpoints
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		apiErr := utils.NewNotFoundError(
			"The requested endpoint was not found",
			r.URL.Path,
		)
		utils.WriteError(w, apiErr)
	})

	// Add method not allowed handler
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		apiErr := utils.APIError{
			Type:      utils.ErrorTypeValidation,
			Title:     "Method Not Allowed",
			Status:    http.StatusMethodNotAllowed,
			Detail:    "The requested method is not allowed for this endpoint",
			Instance:  r.URL.Path,
			Timestamp: utils.GetCurrentTime(),
		}
		utils.WriteError(w, apiErr)
	})

	return r
}

// GetHandler returns the configured HTTP handler
func (rt *Router) GetHandler() http.Handler {
	return rt.SetupRoutes()
}