package models

import (
	"testing"
	"time"
)

func TestUser_ToProfile(t *testing.T) {
	// Test the ToProfile method that actually exists
	user := User{
		ID:        "test-id",
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		LastLogin: nil,
		IsActive:  true,
	}

	profile := user.ToProfile()

	if profile.ID != user.ID {
		t.Errorf("Expected ID %s, got %s", user.ID, profile.ID)
	}
	if profile.Username != user.Username {
		t.Errorf("Expected Username %s, got %s", user.Username, profile.Username)
	}
	if profile.Email != user.Email {
		t.Errorf("Expected Email %s, got %s", user.Email, profile.Email)
	}
	if !profile.CreatedAt.Equal(user.CreatedAt) {
		t.Errorf("Expected CreatedAt %v, got %v", user.CreatedAt, profile.CreatedAt)
	}
	if profile.LastLogin != user.LastLogin {
		t.Errorf("Expected LastLogin %v, got %v", user.LastLogin, profile.LastLogin)
	}
}

func TestUserRegistrationRequest_Validation(t *testing.T) {
	// Test that the struct fields are properly defined
	req := UserRegistrationRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "testpassword123",
	}

	// Basic validation that fields are accessible
	if req.Username == "" {
		t.Error("Username should not be empty")
	}
	if req.Email == "" {
		t.Error("Email should not be empty")
	}
	if req.Password == "" {
		t.Error("Password should not be empty")
	}
}

func TestUserLoginRequest_Validation(t *testing.T) {
	// Test that the struct fields are properly defined
	req := UserLoginRequest{
		Email:    "test@example.com",
		Password: "testpassword123",
	}

	// Basic validation that fields are accessible
	if req.Email == "" {
		t.Error("Email should not be empty")
	}
	if req.Password == "" {
		t.Error("Password should not be empty")
	}
}

func TestUserLoginResponse_Structure(t *testing.T) {
	// Test that the response structure is correct
	user := User{
		ID:       "test-id",
		Username: "testuser",
		Email:    "test@example.com",
	}

	response := UserLoginResponse{
		User:  user,
		Token: "test-token",
	}

	if response.User.ID != user.ID {
		t.Errorf("Expected User ID %s, got %s", user.ID, response.User.ID)
	}
	if response.Token != "test-token" {
		t.Errorf("Expected Token %s, got %s", "test-token", response.Token)
	}
}

func TestAuthContext_Structure(t *testing.T) {
	// Test that the auth context structure is correct
	ctx := AuthContext{
		UserID:   "test-user-id",
		Username: "testuser",
	}

	if ctx.UserID != "test-user-id" {
		t.Errorf("Expected UserID %s, got %s", "test-user-id", ctx.UserID)
	}
	if ctx.Username != "testuser" {
		t.Errorf("Expected Username %s, got %s", "testuser", ctx.Username)
	}
}