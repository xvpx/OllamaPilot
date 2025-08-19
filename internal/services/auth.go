package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"chat_ollama/internal/config"
	"chat_ollama/internal/database"
	"chat_ollama/internal/models"
	"chat_ollama/internal/utils"
)

// AuthService handles user authentication operations
type AuthService struct {
	db     database.Database
	cfg    *config.Config
	logger *utils.Logger
}

// NewAuthService creates a new authentication service
func NewAuthService(db database.Database, cfg *config.Config, logger *utils.Logger) *AuthService {
	logger.Info().Msg("Creating new auth service")
	service := &AuthService{
		db:     db,
		cfg:    cfg,
		logger: logger,
	}
	logger.Info().Msg("Auth service initialized successfully")
	return service
}

// RegisterUser creates a new user account
func (s *AuthService) RegisterUser(req models.UserRegistrationRequest) (*models.User, error) {
	// Check if username already exists
	existingUser, err := s.GetUserByUsername(req.Username)
	if err == nil && existingUser != nil {
		return nil, utils.NewValidationError("Username already exists", "username")
	}

	// Check if email already exists
	existingUser, err = s.GetUserByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, utils.NewValidationError("Email already exists", "email")
	}

	// Hash password
	hashedPassword, err := s.HashPassword(req.Password)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to hash password")
		return nil, utils.NewInternalError("Failed to create user account", "password_hash")
	}

	// Generate user ID
	userID, err := s.generateID()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate user ID")
		return nil, utils.NewInternalError("Failed to create user account", "user_id")
	}

	// Create user
	user := &models.User{
		ID:        userID,
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}

	// Insert user into database
	query := `
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, username, email, created_at, updated_at, last_login, is_active
	`

	err = s.db.QueryRow(query, user.ID, user.Username, user.Email, user.Password,
		user.CreatedAt, user.UpdatedAt, user.IsActive).Scan(
		&user.ID, &user.Username, &user.Email, &user.CreatedAt,
		&user.UpdatedAt, &user.LastLogin, &user.IsActive)

	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to insert user")
		return nil, utils.NewInternalError("Failed to create user account", "database")
	}

	s.logger.Info().Str("user_id", user.ID).Str("username", user.Username).Msg("User registered successfully")
	return user, nil
}

// LoginUser authenticates a user and returns a JWT token
func (s *AuthService) LoginUser(req models.UserLoginRequest) (*models.UserLoginResponse, error) {
	// Get user by email
	user, err := s.GetUserByEmail(req.Email)
	if err != nil {
		s.logger.Warn().Str("email", req.Email).Msg("Login attempt with invalid email")
		return nil, utils.NewValidationError("Invalid email or password", "credentials")
	}

	if !user.IsActive {
		s.logger.Warn().Str("user_id", user.ID).Msg("Login attempt for inactive user")
		return nil, utils.NewValidationError("Account is inactive", "user_status")
	}

	// Verify password
	if !s.VerifyPassword(req.Password, user.Password) {
		s.logger.Warn().Str("user_id", user.ID).Msg("Login attempt with invalid password")
		return nil, utils.NewValidationError("Invalid email or password", "credentials")
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	s.updateLastLogin(user.ID, now)

	// Generate JWT token
	token, err := s.GenerateJWT(user.ID, user.Username)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", user.ID).Msg("Failed to generate JWT token")
		return nil, utils.NewInternalError("Failed to generate authentication token", "jwt")
	}

	s.logger.Info().Str("user_id", user.ID).Str("username", user.Username).Msg("User logged in successfully")

	return &models.UserLoginResponse{
		User:  *user,
		Token: token,
	}, nil
}

// GetUserByUsername retrieves a user by username
func (s *AuthService) GetUserByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at, last_login, is_active
		FROM users
		WHERE username = $1 AND is_active = true
	`

	user := &models.User{}
	err := s.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin, &user.IsActive)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *AuthService) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at, last_login, is_active
		FROM users
		WHERE email = $1 AND is_active = true
	`

	user := &models.User{}
	err := s.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin, &user.IsActive)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(userID string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at, last_login, is_active
		FROM users
		WHERE id = $1 AND is_active = true
	`

	user := &models.User{}
	err := s.db.QueryRow(query, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin, &user.IsActive)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// HashPassword hashes a password using bcrypt
func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), s.cfg.BCryptCost)
	return string(bytes), err
}

// VerifyPassword verifies a password against its hash
func (s *AuthService) VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT generates a JWT token for a user
func (s *AuthService) GenerateJWT(userID, username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(s.cfg.JWTExpiration).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

// ValidateJWT validates a JWT token and returns the claims
func (s *AuthService) ValidateJWT(tokenString string) (*models.AuthContext, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user_id in token")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid username in token")
	}

	return &models.AuthContext{
		UserID:   userID,
		Username: username,
	}, nil
}

// updateLastLogin updates the user's last login timestamp
func (s *AuthService) updateLastLogin(userID string, loginTime time.Time) {
	query := `UPDATE users SET last_login = $1, updated_at = $2 WHERE id = $3`
	_, err := s.db.Exec(query, loginTime, time.Now(), userID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update last login")
	}
}

// generateID generates a random ID for users
func (s *AuthService) generateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}