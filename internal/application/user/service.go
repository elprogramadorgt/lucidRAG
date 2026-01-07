package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	userDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/user"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Sentinel errors for user operations.
var (
	// ErrEmailExists is returned when registering with an email already in use.
	ErrEmailExists = errors.New("email already exists")
	// ErrInvalidCredentials is returned when login credentials are incorrect.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrInvalidToken is returned when a JWT token is malformed or expired.
	ErrInvalidToken = errors.New("invalid token")
	// ErrUserNotFound is returned when a user cannot be found.
	ErrUserNotFound = errors.New("user not found")
)

type jwtClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type service struct {
	repo      userDomain.Repository
	jwtSecret []byte
	jwtExpiry time.Duration
}

// ServiceConfig contains dependencies for creating a user service.
type ServiceConfig struct {
	Repo      userDomain.Repository
	JWTSecret string
	JWTExpiry time.Duration
}

// NewService creates a new user service with the given configuration.
func NewService(cfg ServiceConfig) userDomain.Service {
	expiry := cfg.JWTExpiry
	if expiry == 0 {
		expiry = 24 * time.Hour
	}

	return &service{
		repo:      cfg.Repo,
		jwtSecret: []byte(cfg.JWTSecret),
		jwtExpiry: expiry,
	}
}

func (s *service) Register(ctx context.Context, newUser userDomain.User) (*userDomain.User, error) {
	existing, _ := s.repo.GetByEmail(ctx, newUser.Email)
	if existing != nil {
		return nil, ErrEmailExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newUser.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &userDomain.User{
		Email:        newUser.Email,
		PasswordHash: string(hash),
		FirstName:    newUser.FirstName,
		LastName:     newUser.LastName,
		Role:         userDomain.RoleUser,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	id, err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	user.ID = id

	return user, nil
}

func (s *service) Login(ctx context.Context, email, password string) (string, *userDomain.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return "", nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return "", nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	claims := &jwtClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", nil, fmt.Errorf("sign token: %w", err)
	}
	return tokenStr, user, nil
}

func (s *service) GetUser(ctx context.Context, id string) (*userDomain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *service) ValidateToken(tokenString string) (*userDomain.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid {
		return &userDomain.Claims{
			UserID: claims.UserID,
			Email:  claims.Email,
			Role:   claims.Role,
		}, nil
	}

	return nil, ErrInvalidToken
}

func (s *service) GetUserByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *service) RegisterOAuth(ctx context.Context, newUser userDomain.User, provider, providerID string) (*userDomain.User, error) {
	// Check if user already exists with this email
	existing, _ := s.repo.GetByEmail(ctx, newUser.Email)
	if existing != nil {
		// User exists, update OAuth info if needed and return
		return existing, nil
	}

	user := &userDomain.User{
		Email:           newUser.Email,
		PasswordHash:    "", // OAuth users don't have passwords
		FirstName:       newUser.FirstName,
		LastName:        newUser.LastName,
		Role:            userDomain.RoleUser,
		IsActive:        true,
		OAuthProvider:   provider,
		OAuthProviderID: providerID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	id, err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("create oauth user: %w", err)
	}
	user.ID = id

	return user, nil
}

func (s *service) GenerateToken(user *userDomain.User) (string, error) {
	claims := &jwtClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
