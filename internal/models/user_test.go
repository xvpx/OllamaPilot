package models

import (
	"testing"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name: "valid user",
			user: User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "empty username",
			user: User{
				Username: "",
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "empty email",
			user: User{
				Username: "testuser",
				Email:    "",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "empty password",
			user: User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.user.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_HashPassword(t *testing.T) {
	user := &User{Password: "testpassword"}
	
	err := user.HashPassword()
	if err != nil {
		t.Errorf("User.HashPassword() error = %v", err)
	}
	
	if user.Password == "testpassword" {
		t.Error("Password should be hashed")
	}
	
	if len(user.Password) == 0 {
		t.Error("Hashed password should not be empty")
	}
}

func TestUser_CheckPassword(t *testing.T) {
	user := &User{Password: "testpassword"}
	user.HashPassword()
	
	if !user.CheckPassword("testpassword") {
		t.Error("CheckPassword should return true for correct password")
	}
	
	if user.CheckPassword("wrongpassword") {
		t.Error("CheckPassword should return false for incorrect password")
	}
}