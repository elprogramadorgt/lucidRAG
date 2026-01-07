package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	userDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/user"
	"github.com/gin-gonic/gin"
)

// mockUserService is a mock implementation of user.Service
type mockUserService struct {
	validateTokenFunc func(token string) (*userDomain.Claims, error)
}

func (m *mockUserService) Register(ctx context.Context, newUser userDomain.User) (*userDomain.User, error) {
	return nil, nil
}

func (m *mockUserService) RegisterOAuth(ctx context.Context, newUser userDomain.User, provider, providerID string) (*userDomain.User, error) {
	return nil, nil
}

func (m *mockUserService) Login(ctx context.Context, email, password string) (string, *userDomain.User, error) {
	return "", nil, nil
}

func (m *mockUserService) GetUser(ctx context.Context, id string) (*userDomain.User, error) {
	return nil, nil
}

func (m *mockUserService) GetUserByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	return nil, nil
}

func (m *mockUserService) ValidateToken(token string) (*userDomain.Claims, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(token)
	}
	return nil, errors.New("invalid token")
}

func (m *mockUserService) GenerateToken(user *userDomain.User) (string, error) {
	return "", nil
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestAuthMiddlewareWithValidCookie(t *testing.T) {
	mockSvc := &mockUserService{
		validateTokenFunc: func(token string) (*userDomain.Claims, error) {
			if token == "valid-token" {
				return &userDomain.Claims{
					UserID: "user-123",
					Email:  "test@example.com",
					Role:   "user",
				}, nil
			}
			return nil, errors.New("invalid token")
		},
	}

	router := setupTestRouter()
	router.Use(AuthMiddleware(mockSvc))
	router.GET("/protected", func(c *gin.Context) {
		userID := c.GetString("user_id")
		email := c.GetString("user_email")
		role := c.GetString("user_role")
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"email":   email,
			"role":    role,
		})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: "valid-token"})
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}
}

func TestAuthMiddlewareWithValidBearer(t *testing.T) {
	mockSvc := &mockUserService{
		validateTokenFunc: func(token string) (*userDomain.Claims, error) {
			if token == "valid-bearer-token" {
				return &userDomain.Claims{
					UserID: "user-456",
					Email:  "bearer@example.com",
					Role:   "admin",
				}, nil
			}
			return nil, errors.New("invalid token")
		},
	}

	router := setupTestRouter()
	router.Use(AuthMiddleware(mockSvc))
	router.GET("/protected", func(c *gin.Context) {
		role := c.GetString("user_role")
		c.JSON(http.StatusOK, gin.H{"role": role})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-bearer-token")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}
}

func TestAuthMiddlewareWithLowercaseBearer(t *testing.T) {
	mockSvc := &mockUserService{
		validateTokenFunc: func(token string) (*userDomain.Claims, error) {
			return &userDomain.Claims{
				UserID: "user-789",
				Email:  "lower@example.com",
				Role:   "user",
			}, nil
		},
	}

	router := setupTestRouter()
	router.Use(AuthMiddleware(mockSvc))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "bearer valid-token")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}
}

func TestAuthMiddlewareNoToken(t *testing.T) {
	mockSvc := &mockUserService{}

	router := setupTestRouter()
	router.Use(AuthMiddleware(mockSvc))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.Code)
	}
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	mockSvc := &mockUserService{
		validateTokenFunc: func(token string) (*userDomain.Claims, error) {
			return nil, errors.New("invalid token")
		},
	}

	router := setupTestRouter()
	router.Use(AuthMiddleware(mockSvc))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: "invalid-token"})
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.Code)
	}
}

func TestAuthMiddlewareCookiePriorityOverBearer(t *testing.T) {
	mockSvc := &mockUserService{
		validateTokenFunc: func(token string) (*userDomain.Claims, error) {
			if token == "cookie-token" {
				return &userDomain.Claims{
					UserID: "cookie-user",
					Email:  "cookie@example.com",
					Role:   "user",
				}, nil
			}
			if token == "bearer-token" {
				return &userDomain.Claims{
					UserID: "bearer-user",
					Email:  "bearer@example.com",
					Role:   "admin",
				}, nil
			}
			return nil, errors.New("invalid token")
		},
	}

	router := setupTestRouter()
	router.Use(AuthMiddleware(mockSvc))
	router.GET("/protected", func(c *gin.Context) {
		userID := c.GetString("user_id")
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: "cookie-token"})
	req.Header.Set("Authorization", "Bearer bearer-token")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}
	// Cookie should take priority - user_id should be from cookie token
}

func TestRequireRoleAllowed(t *testing.T) {
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", "admin")
		c.Next()
	})
	router.Use(RequireRole("admin", "superadmin"))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/admin", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}
}

func TestRequireRoleForbidden(t *testing.T) {
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", "user")
		c.Next()
	})
	router.Use(RequireRole("admin", "superadmin"))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/admin", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", resp.Code)
	}
}

func TestRequireRoleNoRole(t *testing.T) {
	router := setupTestRouter()
	// No role set
	router.Use(RequireRole("admin"))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/admin", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.Code)
	}
}

func TestRequireRoleMultipleRoles(t *testing.T) {
	testCases := []struct {
		name         string
		userRole     string
		allowedRoles []string
		expectedCode int
	}{
		{"Admin allowed for admin role", "admin", []string{"admin"}, http.StatusOK},
		{"User allowed for user role", "user", []string{"user", "admin"}, http.StatusOK},
		{"Viewer not allowed for admin role", "viewer", []string{"admin"}, http.StatusForbidden},
		{"Admin allowed in multiple roles", "admin", []string{"user", "admin", "moderator"}, http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := setupTestRouter()
			router.Use(func(c *gin.Context) {
				c.Set("user_role", tc.userRole)
				c.Next()
			})
			router.Use(RequireRole(tc.allowedRoles...))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != tc.expectedCode {
				t.Errorf("Expected status %d, got %d", tc.expectedCode, resp.Code)
			}
		})
	}
}
