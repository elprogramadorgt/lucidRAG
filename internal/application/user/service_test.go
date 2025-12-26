package user

import (
	"context"
	"errors"
	"testing"
	"time"

	userDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/user"
)

// mockUserRepo is a mock implementation of the user repository
type mockUserRepo struct {
	users       map[string]*userDomain.User
	emailIndex  map[string]*userDomain.User
	createError error
	getError    error
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:      make(map[string]*userDomain.User),
		emailIndex: make(map[string]*userDomain.User),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, user *userDomain.User) (string, error) {
	if m.createError != nil {
		return "", m.createError
	}
	id := "user_" + user.Email
	user.ID = id
	m.users[id] = user
	m.emailIndex[user.Email] = user
	return id, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*userDomain.User, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	user, exists := m.users[id]
	if !exists {
		return nil, nil
	}
	return user, nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	user, exists := m.emailIndex[email]
	if !exists {
		return nil, nil
	}
	return user, nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *userDomain.User) error {
	m.users[user.ID] = user
	m.emailIndex[user.Email] = user
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id string) error {
	if user, exists := m.users[id]; exists {
		delete(m.emailIndex, user.Email)
		delete(m.users, id)
	}
	return nil
}

func TestNewService(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(ServiceConfig{
		Repo:      repo,
		JWTSecret: "test-secret-key-that-is-long-enough",
		JWTExpiry: 24 * time.Hour,
	})

	if svc == nil {
		t.Fatal("Expected service to be created")
	}
}

func TestNewServiceDefaultExpiry(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(ServiceConfig{
		Repo:      repo,
		JWTSecret: "test-secret-key-that-is-long-enough",
		JWTExpiry: 0, // Should default to 24 hours
	})

	if svc == nil {
		t.Fatal("Expected service to be created with default expiry")
	}
}

func TestRegister(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(ServiceConfig{
		Repo:      repo,
		JWTSecret: "test-secret-key-that-is-long-enough",
		JWTExpiry: 24 * time.Hour,
	})

	ctx := context.Background()
	newUser := userDomain.User{
		Email:        "test@example.com",
		PasswordHash: "password123",
		FirstName:    "Test",
		LastName:     "User",
	}

	user, err := svc.Register(ctx, newUser)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", user.Email)
	}
	if user.Role != userDomain.RoleUser {
		t.Errorf("Expected role user, got %s", user.Role)
	}
	if !user.IsActive {
		t.Error("Expected user to be active")
	}
	if user.PasswordHash == "password123" {
		t.Error("Expected password to be hashed")
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(ServiceConfig{
		Repo:      repo,
		JWTSecret: "test-secret-key-that-is-long-enough",
		JWTExpiry: 24 * time.Hour,
	})

	ctx := context.Background()
	newUser := userDomain.User{
		Email:        "test@example.com",
		PasswordHash: "password123",
		FirstName:    "Test",
		LastName:     "User",
	}

	// First registration should succeed
	_, err := svc.Register(ctx, newUser)
	if err != nil {
		t.Fatalf("First registration failed: %v", err)
	}

	// Second registration with same email should fail
	_, err = svc.Register(ctx, newUser)
	if !errors.Is(err, ErrEmailExists) {
		t.Errorf("Expected ErrEmailExists, got %v", err)
	}
}

func TestLogin(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(ServiceConfig{
		Repo:      repo,
		JWTSecret: "test-secret-key-that-is-long-enough",
		JWTExpiry: 24 * time.Hour,
	})

	ctx := context.Background()

	// Register a user first
	newUser := userDomain.User{
		Email:        "test@example.com",
		PasswordHash: "password123",
		FirstName:    "Test",
		LastName:     "User",
	}
	_, err := svc.Register(ctx, newUser)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Login with correct credentials
	token, user, err := svc.Login(ctx, "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if token == "" {
		t.Error("Expected token to be returned")
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", user.Email)
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(ServiceConfig{
		Repo:      repo,
		JWTSecret: "test-secret-key-that-is-long-enough",
		JWTExpiry: 24 * time.Hour,
	})

	ctx := context.Background()

	// Register a user first
	newUser := userDomain.User{
		Email:        "test@example.com",
		PasswordHash: "password123",
		FirstName:    "Test",
		LastName:     "User",
	}
	_, err := svc.Register(ctx, newUser)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Login with wrong password
	_, _, err = svc.Login(ctx, "test@example.com", "wrongpassword")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}

	// Login with non-existent email
	_, _, err = svc.Login(ctx, "nonexistent@example.com", "password123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginInactiveUser(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(ServiceConfig{
		Repo:      repo,
		JWTSecret: "test-secret-key-that-is-long-enough",
		JWTExpiry: 24 * time.Hour,
	})

	ctx := context.Background()

	// Register a user first
	newUser := userDomain.User{
		Email:        "test@example.com",
		PasswordHash: "password123",
		FirstName:    "Test",
		LastName:     "User",
	}
	user, err := svc.Register(ctx, newUser)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Deactivate the user
	user.IsActive = false
	repo.Update(ctx, user)

	// Login should fail for inactive user
	_, _, err = svc.Login(ctx, "test@example.com", "password123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Expected ErrInvalidCredentials for inactive user, got %v", err)
	}
}

func TestValidateToken(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(ServiceConfig{
		Repo:      repo,
		JWTSecret: "test-secret-key-that-is-long-enough",
		JWTExpiry: 24 * time.Hour,
	})

	ctx := context.Background()

	// Register and login
	newUser := userDomain.User{
		Email:        "test@example.com",
		PasswordHash: "password123",
		FirstName:    "Test",
		LastName:     "User",
	}
	_, err := svc.Register(ctx, newUser)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	token, _, err := svc.Login(ctx, "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	// Validate the token
	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", claims.Email)
	}
	if claims.Role != "user" {
		t.Errorf("Expected role user, got %s", claims.Role)
	}
}

func TestValidateInvalidToken(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(ServiceConfig{
		Repo:      repo,
		JWTSecret: "test-secret-key-that-is-long-enough",
		JWTExpiry: 24 * time.Hour,
	})

	// Test with invalid token
	_, err := svc.ValidateToken("invalid-token")
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}

	// Test with empty token
	_, err = svc.ValidateToken("")
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Expected ErrInvalidToken for empty token, got %v", err)
	}
}

func TestGetUser(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(ServiceConfig{
		Repo:      repo,
		JWTSecret: "test-secret-key-that-is-long-enough",
		JWTExpiry: 24 * time.Hour,
	})

	ctx := context.Background()

	// Register a user
	newUser := userDomain.User{
		Email:        "test@example.com",
		PasswordHash: "password123",
		FirstName:    "Test",
		LastName:     "User",
	}
	registered, err := svc.Register(ctx, newUser)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Get user by ID
	user, err := svc.GetUser(ctx, registered.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", user.Email)
	}
}

func TestGetUserNotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(ServiceConfig{
		Repo:      repo,
		JWTSecret: "test-secret-key-that-is-long-enough",
		JWTExpiry: 24 * time.Hour,
	})

	ctx := context.Background()

	// Get non-existent user
	_, err := svc.GetUser(ctx, "non-existent-id")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestGetUserByEmail(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(ServiceConfig{
		Repo:      repo,
		JWTSecret: "test-secret-key-that-is-long-enough",
		JWTExpiry: 24 * time.Hour,
	})

	ctx := context.Background()

	// Register a user
	newUser := userDomain.User{
		Email:        "test@example.com",
		PasswordHash: "password123",
		FirstName:    "Test",
		LastName:     "User",
	}
	_, err := svc.Register(ctx, newUser)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Get user by email
	user, err := svc.GetUserByEmail(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", user.Email)
	}
}

func TestGetUserByEmailNotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(ServiceConfig{
		Repo:      repo,
		JWTSecret: "test-secret-key-that-is-long-enough",
		JWTExpiry: 24 * time.Hour,
	})

	ctx := context.Background()

	// Get non-existent user by email
	_, err := svc.GetUserByEmail(ctx, "nonexistent@example.com")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestRegisterOAuth(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(ServiceConfig{
		Repo:      repo,
		JWTSecret: "test-secret-key-that-is-long-enough",
		JWTExpiry: 24 * time.Hour,
	})

	ctx := context.Background()

	newUser := userDomain.User{
		Email:     "oauth@example.com",
		FirstName: "OAuth",
		LastName:  "User",
	}

	user, err := svc.RegisterOAuth(ctx, newUser, "google", "google-provider-id-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if user.Email != "oauth@example.com" {
		t.Errorf("Expected email oauth@example.com, got %s", user.Email)
	}
	if user.OAuthProvider != "google" {
		t.Errorf("Expected OAuth provider google, got %s", user.OAuthProvider)
	}
	if user.OAuthProviderID != "google-provider-id-123" {
		t.Errorf("Expected OAuth provider ID google-provider-id-123, got %s", user.OAuthProviderID)
	}
	if user.PasswordHash != "" {
		t.Error("Expected OAuth user to have no password hash")
	}
	if user.Role != userDomain.RoleUser {
		t.Errorf("Expected role user, got %s", user.Role)
	}
}

func TestRegisterOAuthExistingUser(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(ServiceConfig{
		Repo:      repo,
		JWTSecret: "test-secret-key-that-is-long-enough",
		JWTExpiry: 24 * time.Hour,
	})

	ctx := context.Background()

	// Register a regular user first
	newUser := userDomain.User{
		Email:        "existing@example.com",
		PasswordHash: "password123",
		FirstName:    "Existing",
		LastName:     "User",
	}
	existingUser, err := svc.Register(ctx, newUser)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// OAuth registration with same email should return existing user
	oauthUser := userDomain.User{
		Email:     "existing@example.com",
		FirstName: "OAuth",
		LastName:  "User",
	}

	user, err := svc.RegisterOAuth(ctx, oauthUser, "google", "google-provider-id-456")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should return the existing user
	if user.ID != existingUser.ID {
		t.Errorf("Expected existing user ID %s, got %s", existingUser.ID, user.ID)
	}
}

func TestGenerateToken(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(ServiceConfig{
		Repo:      repo,
		JWTSecret: "test-secret-key-that-is-long-enough",
		JWTExpiry: 24 * time.Hour,
	})

	user := &userDomain.User{
		ID:    "user-123",
		Email: "test@example.com",
		Role:  userDomain.RoleAdmin,
	}

	token, err := svc.GenerateToken(user)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token == "" {
		t.Error("Expected token to be generated")
	}

	// Validate the generated token
	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate generated token: %v", err)
	}

	if claims.UserID != "user-123" {
		t.Errorf("Expected user ID user-123, got %s", claims.UserID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", claims.Email)
	}
	if claims.Role != "admin" {
		t.Errorf("Expected role admin, got %s", claims.Role)
	}
}
