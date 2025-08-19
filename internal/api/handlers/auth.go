package handlers

import (
	"encoding/json"
	"net/http"

	"chat_ollama/internal/api/middleware"
	"chat_ollama/internal/config"
	"chat_ollama/internal/database"
	"chat_ollama/internal/models"
	"chat_ollama/internal/services"
	"chat_ollama/internal/utils"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	authService *services.AuthService
	db          database.Database
	cfg         *config.Config
	logger      *utils.Logger
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(db database.Database, cfg *config.Config, logger *utils.Logger) *AuthHandler {
	logger.Info().Msg("Creating new auth handler")
	authService := services.NewAuthService(db, cfg, logger)
	logger.Info().Msg("Auth service created successfully")
	return &AuthHandler{
		authService: authService,
		db:          db,
		cfg:         cfg,
		logger:      logger,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.UserRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiErr := utils.NewValidationError("Invalid request body", "body")
		utils.WriteError(w, apiErr)
		return
	}

	// Validate request
	if err := h.validateRegistrationRequest(req); err.Type != "" {
		utils.WriteError(w, err)
		return
	}

	// Register user
	user, err := h.authService.RegisterUser(req)
	if err != nil {
		if apiErr, ok := err.(utils.APIError); ok {
			utils.WriteError(w, apiErr)
		} else {
			h.logger.Error().Err(err).Msg("Failed to register user")
			apiErr := utils.NewInternalError("Failed to register user", "registration")
			utils.WriteError(w, apiErr)
		}
		return
	}

	// Return user profile (without sensitive data)
	profile := user.ToProfile()
	utils.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully",
		"user":    profile,
	})
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Msg("Login handler called")
	var req models.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode request body")
		apiErr := utils.NewValidationError("Invalid request body", "body")
		utils.WriteError(w, apiErr)
		return
	}

	h.logger.Info().Str("email", req.Email).Msg("Login request received")

	// Validate request
	if err := h.validateLoginRequest(req); err.Type != "" {
		h.logger.Warn().Str("email", req.Email).Msg("Login request validation failed")
		utils.WriteError(w, err)
		return
	}

	h.logger.Info().Str("email", req.Email).Msg("Calling auth service login")
	// Login user
	response, err := h.authService.LoginUser(req)
	if err != nil {
		h.logger.Error().Err(err).Str("email", req.Email).Msg("Auth service login failed")
		if apiErr, ok := err.(utils.APIError); ok {
			utils.WriteError(w, apiErr)
		} else {
			h.logger.Error().Err(err).Msg("Failed to login user")
			apiErr := utils.NewInternalError("Failed to login user", "login")
			utils.WriteError(w, apiErr)
		}
		return
	}

	h.logger.Info().Str("email", req.Email).Msg("Login successful, returning response")
	// Return login response with token
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"user":    response.User.ToProfile(),
		"token":   response.Token,
	})
}

// GetProfile returns the current user's profile
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	authContext, ok := middleware.GetUserFromContext(r)
	if !ok {
		apiErr := utils.NewUnauthorizedError("Authentication required", "authentication")
		utils.WriteError(w, apiErr)
		return
	}

	// Get user details
	user, err := h.authService.GetUserByID(authContext.UserID)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", authContext.UserID).Msg("Failed to get user profile")
		apiErr := utils.NewNotFoundError("User not found", authContext.UserID)
		utils.WriteError(w, apiErr)
		return
	}

	// Return user profile
	profile := user.ToProfile()
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"user": profile,
	})
}

// Logout handles user logout (client-side token invalidation)
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Since we're using stateless JWT tokens, logout is primarily handled client-side
	// The client should remove the token from storage
	// In a more advanced implementation, we could maintain a token blacklist
	
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Logout successful",
	})
}

// RefreshToken handles token refresh (optional endpoint)
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	authContext, ok := middleware.GetUserFromContext(r)
	if !ok {
		apiErr := utils.NewUnauthorizedError("Authentication required", "authentication")
		utils.WriteError(w, apiErr)
		return
	}

	// Generate new token
	token, err := h.authService.GenerateJWT(authContext.UserID, authContext.Username)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", authContext.UserID).Msg("Failed to refresh token")
		apiErr := utils.NewInternalError("Failed to refresh token", "token_refresh")
		utils.WriteError(w, apiErr)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Token refreshed successfully",
		"token":   token,
	})
}

// validateRegistrationRequest validates user registration request
func (h *AuthHandler) validateRegistrationRequest(req models.UserRegistrationRequest) utils.APIError {
	if req.Username == "" {
		return utils.NewValidationError("Username is required", "username")
	}
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return utils.NewValidationError("Username must be between 3 and 50 characters", "username")
	}
	if req.Email == "" {
		return utils.NewValidationError("Email is required", "email")
	}
	if req.Password == "" {
		return utils.NewValidationError("Password is required", "password")
	}
	if len(req.Password) < 8 {
		return utils.NewValidationError("Password must be at least 8 characters long", "password")
	}
	return utils.APIError{}
}

// GetAuthService returns the authentication service (for middleware)
func (h *AuthHandler) GetAuthService() *services.AuthService {
	return h.authService
}

// validateLoginRequest validates user login request
func (h *AuthHandler) validateLoginRequest(req models.UserLoginRequest) utils.APIError {
	if req.Email == "" {
		return utils.NewValidationError("Email is required", "email")
	}
	if req.Password == "" {
		return utils.NewValidationError("Password is required", "password")
	}
	return utils.APIError{}
}