package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/gin-gonic/gin"
)

func setupCommonTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestRequestID(t *testing.T) {
	router := setupCommonTestRouter()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		requestID := c.GetString("request_id")
		c.String(http.StatusOK, requestID)
	})

	t.Run("generates new request ID when not provided", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.Code)
		}

		requestID := resp.Header().Get("X-Request-ID")
		if requestID == "" {
			t.Error("Expected X-Request-ID header to be set")
		}

		bodyID := resp.Body.String()
		if bodyID == "" {
			t.Error("Expected request_id to be set in context")
		}
		if bodyID != requestID {
			t.Errorf("Context ID '%s' should match header ID '%s'", bodyID, requestID)
		}
	})

	t.Run("uses provided request ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", "custom-request-id")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		requestID := resp.Header().Get("X-Request-ID")
		if requestID != "custom-request-id" {
			t.Errorf("Expected X-Request-ID 'custom-request-id', got '%s'", requestID)
		}

		bodyID := resp.Body.String()
		if bodyID != "custom-request-id" {
			t.Errorf("Expected body 'custom-request-id', got '%s'", bodyID)
		}
	})
}

func TestRequestIDUniqueness(t *testing.T) {
	router := setupCommonTestRouter()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	ids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		id := resp.Header().Get("X-Request-ID")
		if ids[id] {
			t.Errorf("Duplicate request ID generated: %s", id)
		}
		ids[id] = true
	}
}

func TestLogger(t *testing.T) {
	log := logger.New(logger.Options{Level: "error"})
	router := setupCommonTestRouter()
	router.Use(RequestID())
	router.Use(Logger(log))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}
}

func TestCORS(t *testing.T) {
	origins := []string{"http://localhost:4200", "https://example.com"}
	router := setupCommonTestRouter()
	router.Use(CORS(origins))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	t.Run("allows configured origin", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:4200")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.Code)
		}

		allowedOrigin := resp.Header().Get("Access-Control-Allow-Origin")
		if allowedOrigin != "http://localhost:4200" {
			t.Errorf("Expected allowed origin 'http://localhost:4200', got '%s'", allowedOrigin)
		}

		credentials := resp.Header().Get("Access-Control-Allow-Credentials")
		if credentials != "true" {
			t.Errorf("Expected credentials 'true', got '%s'", credentials)
		}
	})

	t.Run("allows second configured origin", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		allowedOrigin := resp.Header().Get("Access-Control-Allow-Origin")
		if allowedOrigin != "https://example.com" {
			t.Errorf("Expected allowed origin 'https://example.com', got '%s'", allowedOrigin)
		}
	})

	t.Run("does not allow unconfigured origin", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://malicious.com")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		allowedOrigin := resp.Header().Get("Access-Control-Allow-Origin")
		if allowedOrigin != "" {
			t.Errorf("Expected no allowed origin for unauthorized site, got '%s'", allowedOrigin)
		}
	})

	t.Run("handles OPTIONS preflight request", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:4200")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", resp.Code)
		}

		methods := resp.Header().Get("Access-Control-Allow-Methods")
		if methods != "GET, POST, PUT, DELETE, OPTIONS" {
			t.Errorf("Expected methods 'GET, POST, PUT, DELETE, OPTIONS', got '%s'", methods)
		}

		headers := resp.Header().Get("Access-Control-Allow-Headers")
		if headers != "Content-Type, Authorization, X-Request-ID" {
			t.Errorf("Expected headers 'Content-Type, Authorization, X-Request-ID', got '%s'", headers)
		}
	})

	t.Run("sets headers without origin match", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		// No Origin header set
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		// Should still set the allowed methods and headers
		methods := resp.Header().Get("Access-Control-Allow-Methods")
		if methods != "GET, POST, PUT, DELETE, OPTIONS" {
			t.Errorf("Expected methods header to be set, got '%s'", methods)
		}
	})
}

func TestCORSEmptyOrigins(t *testing.T) {
	router := setupCommonTestRouter()
	router.Use(CORS([]string{}))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:4200")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	allowedOrigin := resp.Header().Get("Access-Control-Allow-Origin")
	if allowedOrigin != "" {
		t.Errorf("Expected no allowed origin for empty config, got '%s'", allowedOrigin)
	}
}
