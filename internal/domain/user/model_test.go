package user

import (
	"testing"
	"time"
)

func TestRoleConstants(t *testing.T) {
	if RoleUser != "user" {
		t.Errorf("Expected RoleUser to be 'user', got '%s'", RoleUser)
	}
	if RoleAdmin != "admin" {
		t.Errorf("Expected RoleAdmin to be 'admin', got '%s'", RoleAdmin)
	}
}

func TestUserStruct(t *testing.T) {
	now := time.Now()
	user := User{
		ID:              "user-123",
		Email:           "test@example.com",
		PasswordHash:    "hashed",
		FirstName:       "John",
		LastName:        "Doe",
		Role:            RoleUser,
		IsActive:        true,
		OAuthProvider:   "google",
		OAuthProviderID: "oauth-123",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if user.ID != "user-123" {
		t.Errorf("Expected ID 'user-123', got '%s'", user.ID)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected Email 'test@example.com', got '%s'", user.Email)
	}
	if user.FirstName != "John" {
		t.Errorf("Expected FirstName 'John', got '%s'", user.FirstName)
	}
	if user.LastName != "Doe" {
		t.Errorf("Expected LastName 'Doe', got '%s'", user.LastName)
	}
	if user.Role != RoleUser {
		t.Errorf("Expected Role 'user', got '%s'", user.Role)
	}
	if !user.IsActive {
		t.Error("Expected IsActive to be true")
	}
	if user.OAuthProvider != "google" {
		t.Errorf("Expected OAuthProvider 'google', got '%s'", user.OAuthProvider)
	}
}

func TestUserZeroValue(t *testing.T) {
	var user User
	if user.ID != "" {
		t.Errorf("Expected empty ID, got '%s'", user.ID)
	}
	if user.IsActive {
		t.Error("Expected IsActive to be false by default")
	}
	if user.Role != "" {
		t.Errorf("Expected empty Role, got '%s'", user.Role)
	}
}

func TestRoleComparison(t *testing.T) {
	adminUser := User{Role: RoleAdmin}
	regularUser := User{Role: RoleUser}

	if adminUser.Role == regularUser.Role {
		t.Error("Admin and regular user should have different roles")
	}

	if adminUser.Role != RoleAdmin {
		t.Error("Admin user should have RoleAdmin")
	}

	if regularUser.Role != RoleUser {
		t.Error("Regular user should have RoleUser")
	}
}
