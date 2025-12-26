package system

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/elprogramadorgt/lucidRAG/internal/domain/system"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/gin-gonic/gin"
)

// mockLogRepository implements system.LogRepository for testing
type mockLogRepository struct {
	listFn            func(ctx context.Context, filter system.LogFilter) ([]system.LogEntry, int64, error)
	statsFn           func(ctx context.Context) (*system.LogStats, error)
	deleteOlderThanFn func(ctx context.Context, days int) (int64, error)
}

func (m *mockLogRepository) Insert(ctx context.Context, entry *system.LogEntry) error {
	return nil
}

func (m *mockLogRepository) List(ctx context.Context, filter system.LogFilter) ([]system.LogEntry, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return []system.LogEntry{}, 0, nil
}

func (m *mockLogRepository) Stats(ctx context.Context) (*system.LogStats, error) {
	if m.statsFn != nil {
		return m.statsFn(ctx)
	}
	return &system.LogStats{}, nil
}

func (m *mockLogRepository) DeleteOlderThan(ctx context.Context, days int) (int64, error) {
	if m.deleteOlderThanFn != nil {
		return m.deleteOlderThanFn(ctx, days)
	}
	return 0, nil
}

// mockDBPinger implements DBPinger for testing
type mockDBPinger struct {
	pingFn func(ctx context.Context) error
}

func (m *mockDBPinger) Ping(ctx context.Context) error {
	if m.pingFn != nil {
		return m.pingFn(ctx)
	}
	return nil
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func createTestHandler(repo *mockLogRepository, db *mockDBPinger) *Handler {
	log := logger.New(logger.Options{Level: "error"})
	return NewHandler(HandlerConfig{
		Repo:        repo,
		DB:          db,
		Log:         log,
		StartTime:   time.Now(),
		Environment: "test",
		Version:     "1.0.0",
	})
}

func TestNewHandler(t *testing.T) {
	log := logger.New(logger.Options{Level: "error"})
	handler := NewHandler(HandlerConfig{
		Repo:        &mockLogRepository{},
		DB:          &mockDBPinger{},
		Log:         log,
		StartTime:   time.Now(),
		Environment: "test",
		Version:     "",
	})

	if handler == nil {
		t.Fatal("Expected non-nil handler")
	}
	if handler.version != "dev" {
		t.Errorf("Expected default version 'dev', got '%s'", handler.version)
	}
}

func TestListLogs(t *testing.T) {
	now := time.Now()
	repo := &mockLogRepository{
		listFn: func(ctx context.Context, filter system.LogFilter) ([]system.LogEntry, int64, error) {
			return []system.LogEntry{
				{ID: "log-1", Level: "INFO", Message: "test1", Timestamp: now},
				{ID: "log-2", Level: "ERROR", Message: "test2", Timestamp: now},
			}, 2, nil
		},
	}
	handler := createTestHandler(repo, &mockDBPinger{})

	router := setupTestRouter()
	router.GET("/logs", func(c *gin.Context) {
		c.Set("user_id", "admin-123")
		handler.ListLogs(c)
	})

	req, _ := http.NewRequest("GET", "/logs", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	logs, ok := result["logs"].([]interface{})
	if !ok {
		t.Fatal("Expected logs array in response")
	}
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs, got %d", len(logs))
	}
}

func TestListLogsWithFilters(t *testing.T) {
	var capturedFilter system.LogFilter
	repo := &mockLogRepository{
		listFn: func(ctx context.Context, filter system.LogFilter) ([]system.LogEntry, int64, error) {
			capturedFilter = filter
			return []system.LogEntry{}, 0, nil
		},
	}
	handler := createTestHandler(repo, &mockDBPinger{})

	router := setupTestRouter()
	router.GET("/logs", func(c *gin.Context) {
		c.Set("user_id", "admin-123")
		handler.ListLogs(c)
	})

	req, _ := http.NewRequest("GET", "/logs?level=ERROR&search=test&limit=25&offset=10", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if capturedFilter.Level != "ERROR" {
		t.Errorf("Expected level 'ERROR', got '%s'", capturedFilter.Level)
	}
	if capturedFilter.Search != "test" {
		t.Errorf("Expected search 'test', got '%s'", capturedFilter.Search)
	}
	if capturedFilter.Limit != 25 {
		t.Errorf("Expected limit 25, got %d", capturedFilter.Limit)
	}
	if capturedFilter.Offset != 10 {
		t.Errorf("Expected offset 10, got %d", capturedFilter.Offset)
	}
}

func TestListLogsError(t *testing.T) {
	repo := &mockLogRepository{
		listFn: func(ctx context.Context, filter system.LogFilter) ([]system.LogEntry, int64, error) {
			return nil, 0, errors.New("database error")
		},
	}
	handler := createTestHandler(repo, &mockDBPinger{})

	router := setupTestRouter()
	router.GET("/logs", func(c *gin.Context) {
		c.Set("user_id", "admin-123")
		handler.ListLogs(c)
	})

	req, _ := http.NewRequest("GET", "/logs", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.Code)
	}
}

func TestGetStats(t *testing.T) {
	repo := &mockLogRepository{
		statsFn: func(ctx context.Context) (*system.LogStats, error) {
			return &system.LogStats{
				TotalCount: 100,
				LevelCounts: map[string]int64{
					"INFO":  80,
					"ERROR": 20,
				},
			}, nil
		},
	}
	handler := createTestHandler(repo, &mockDBPinger{})

	router := setupTestRouter()
	router.GET("/stats", func(c *gin.Context) {
		c.Set("user_id", "admin-123")
		handler.GetStats(c)
	})

	req, _ := http.NewRequest("GET", "/stats", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}

	var result system.LogStats
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if result.TotalCount != 100 {
		t.Errorf("Expected total count 100, got %d", result.TotalCount)
	}
}

func TestGetStatsError(t *testing.T) {
	repo := &mockLogRepository{
		statsFn: func(ctx context.Context) (*system.LogStats, error) {
			return nil, errors.New("database error")
		},
	}
	handler := createTestHandler(repo, &mockDBPinger{})

	router := setupTestRouter()
	router.GET("/stats", func(c *gin.Context) {
		c.Set("user_id", "admin-123")
		handler.GetStats(c)
	})

	req, _ := http.NewRequest("GET", "/stats", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.Code)
	}
}

func TestCleanupLogs(t *testing.T) {
	var capturedDays int
	repo := &mockLogRepository{
		deleteOlderThanFn: func(ctx context.Context, days int) (int64, error) {
			capturedDays = days
			return 50, nil
		},
	}
	handler := createTestHandler(repo, &mockDBPinger{})

	router := setupTestRouter()
	router.DELETE("/cleanup", func(c *gin.Context) {
		c.Set("user_id", "admin-123")
		handler.CleanupLogs(c)
	})

	req, _ := http.NewRequest("DELETE", "/cleanup?days=7", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}

	if capturedDays != 7 {
		t.Errorf("Expected days 7, got %d", capturedDays)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if result["deleted"].(float64) != 50 {
		t.Errorf("Expected deleted 50, got %v", result["deleted"])
	}
}

func TestCleanupLogsDefaultDays(t *testing.T) {
	var capturedDays int
	repo := &mockLogRepository{
		deleteOlderThanFn: func(ctx context.Context, days int) (int64, error) {
			capturedDays = days
			return 0, nil
		},
	}
	handler := createTestHandler(repo, &mockDBPinger{})

	router := setupTestRouter()
	router.DELETE("/cleanup", func(c *gin.Context) {
		c.Set("user_id", "admin-123")
		handler.CleanupLogs(c)
	})

	req, _ := http.NewRequest("DELETE", "/cleanup", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if capturedDays != 30 {
		t.Errorf("Expected default days 30, got %d", capturedDays)
	}
}

func TestCleanupLogsError(t *testing.T) {
	repo := &mockLogRepository{
		deleteOlderThanFn: func(ctx context.Context, days int) (int64, error) {
			return 0, errors.New("database error")
		},
	}
	handler := createTestHandler(repo, &mockDBPinger{})

	router := setupTestRouter()
	router.DELETE("/cleanup", func(c *gin.Context) {
		c.Set("user_id", "admin-123")
		handler.CleanupLogs(c)
	})

	req, _ := http.NewRequest("DELETE", "/cleanup", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.Code)
	}
}

func TestGetServerInfo(t *testing.T) {
	db := &mockDBPinger{
		pingFn: func(ctx context.Context) error {
			return nil
		},
	}
	handler := createTestHandler(&mockLogRepository{}, db)

	router := setupTestRouter()
	router.GET("/info", func(c *gin.Context) {
		c.Set("user_id", "admin-123")
		handler.GetServerInfo(c)
	})

	req, _ := http.NewRequest("GET", "/info", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}

	var result ServerInfo
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if result.Status != "running" {
		t.Errorf("Expected status 'running', got '%s'", result.Status)
	}
	if result.Environment != "test" {
		t.Errorf("Expected environment 'test', got '%s'", result.Environment)
	}
	if result.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", result.Version)
	}
	if result.Database.Status != "connected" {
		t.Errorf("Expected database status 'connected', got '%s'", result.Database.Status)
	}
}

func TestGetServerInfoDBDisconnected(t *testing.T) {
	db := &mockDBPinger{
		pingFn: func(ctx context.Context) error {
			return errors.New("connection failed")
		},
	}
	handler := createTestHandler(&mockLogRepository{}, db)

	router := setupTestRouter()
	router.GET("/info", func(c *gin.Context) {
		c.Set("user_id", "admin-123")
		handler.GetServerInfo(c)
	})

	req, _ := http.NewRequest("GET", "/info", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	var result ServerInfo
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if result.Database.Status != "disconnected" {
		t.Errorf("Expected database status 'disconnected', got '%s'", result.Database.Status)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{5 * time.Second, "5s"},
		{65 * time.Second, "1m 5s"},
		{3665 * time.Second, "1h 1m 5s"},
		{90061 * time.Second, "1d 1h 1m"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestServerInfoStruct(t *testing.T) {
	info := ServerInfo{
		Status:      "running",
		Environment: "production",
		Version:     "1.0.0",
		Uptime:      "1h 30m 0s",
		UptimeSecs:  5400,
		Database:    DatabaseStatus{Status: "connected"},
		Runtime:     RuntimeInfo{GoVersion: "go1.21"},
		Endpoints:   []EndpointInfo{{Path: "/test", Method: "GET", Description: "Test"}},
	}

	if info.Status != "running" {
		t.Errorf("Expected status 'running', got '%s'", info.Status)
	}
	if info.UptimeSecs != 5400 {
		t.Errorf("Expected uptime 5400, got %d", info.UptimeSecs)
	}
	if len(info.Endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(info.Endpoints))
	}
}
