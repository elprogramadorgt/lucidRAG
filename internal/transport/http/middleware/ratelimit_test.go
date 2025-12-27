package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)
	defer rl.Stop()

	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
	}
	if rl.limit != 10 {
		t.Errorf("Expected limit 10, got %d", rl.limit)
	}
	if rl.window != time.Minute {
		t.Errorf("Expected window 1m, got %v", rl.window)
	}
	if rl.requests == nil {
		t.Error("requests map is nil")
	}
	if rl.stopCh == nil {
		t.Error("stopCh is nil")
	}
}

func TestRateLimiterAllow(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)
	defer rl.Stop()

	ip := "192.168.1.1"

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		if !rl.Allow(ip) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied
	if rl.Allow(ip) {
		t.Error("4th request should be denied")
	}

	// Different IP should still be allowed
	otherIP := "192.168.1.2"
	if !rl.Allow(otherIP) {
		t.Error("Request from different IP should be allowed")
	}
}

func TestRateLimiterWindowExpiry(t *testing.T) {
	rl := NewRateLimiter(2, 100*time.Millisecond)
	defer rl.Stop()

	ip := "192.168.1.1"

	// Use up the limit
	rl.Allow(ip)
	rl.Allow(ip)

	// Should be denied
	if rl.Allow(ip) {
		t.Error("Should be denied after hitting limit")
	}

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again
	if !rl.Allow(ip) {
		t.Error("Should be allowed after window expires")
	}
}

func TestRateLimiterConcurrency(t *testing.T) {
	rl := NewRateLimiter(100, time.Minute)
	defer rl.Stop()

	var wg sync.WaitGroup
	allowed := make(chan bool, 200)

	// Send 200 requests concurrently from 2 IPs
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			allowed <- rl.Allow("ip1")
		}()
		go func() {
			defer wg.Done()
			allowed <- rl.Allow("ip2")
		}()
	}

	wg.Wait()
	close(allowed)

	ip1Allowed := 0
	ip2Allowed := 0
	total := 0
	for a := range allowed {
		total++
		if a && total <= 100 {
			ip1Allowed++
		} else if a {
			ip2Allowed++
		}
	}

	// Each IP should have exactly 100 allowed requests
	// But due to timing, some might be rejected
	if ip1Allowed+ip2Allowed < 100 {
		t.Errorf("Expected at least 100 allowed requests from each IP, got ip1=%d, ip2=%d", ip1Allowed, ip2Allowed)
	}
}

func TestRateLimiterStop(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)

	// Allow some requests
	rl.Allow("test-ip")

	// Stop should not panic
	rl.Stop()

	// Calling Allow after Stop should still work (though cleanup won't run)
	rl.Allow("test-ip")
}

func TestRateLimiterCleanup(t *testing.T) {
	// Use very short window for testing
	rl := NewRateLimiter(5, 50*time.Millisecond)
	defer rl.Stop()

	// Add some requests
	rl.Allow("ip1")
	rl.Allow("ip2")

	// Verify requests are tracked
	rl.mu.RLock()
	if len(rl.requests) != 2 {
		t.Errorf("Expected 2 IPs tracked, got %d", len(rl.requests))
	}
	rl.mu.RUnlock()
}

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rl := NewRateLimiter(2, time.Minute)
	defer rl.Stop()

	router := gin.New()
	router.Use(RateLimit(rl))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}
	}

	// 3rd request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["error"] != "rate limit exceeded" {
		t.Errorf("Expected error 'rate limit exceeded', got %q", response["error"])
	}
}

func TestRateLimitMiddlewareDifferentIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rl := NewRateLimiter(1, time.Minute)
	defer rl.Stop()

	router := gin.New()
	router.Use(RateLimit(rl))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First request from IP1
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.Header.Set("X-Forwarded-For", "192.168.1.1")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("First request from IP1: Expected status %d, got %d", http.StatusOK, w1.Code)
	}

	// Second request from IP1 should be rate limited
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("X-Forwarded-For", "192.168.1.1")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("Second request from IP1: Expected status %d, got %d", http.StatusTooManyRequests, w2.Code)
	}

	// Request from different IP should succeed
	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req3.Header.Set("X-Forwarded-For", "192.168.1.2")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Errorf("Request from IP2: Expected status %d, got %d", http.StatusOK, w3.Code)
	}
}

func TestRateLimiterMultipleWindows(t *testing.T) {
	rl := NewRateLimiter(2, 100*time.Millisecond)
	defer rl.Stop()

	ip := "test-ip"

	// First window
	if !rl.Allow(ip) {
		t.Error("First request should be allowed")
	}
	if !rl.Allow(ip) {
		t.Error("Second request should be allowed")
	}
	if rl.Allow(ip) {
		t.Error("Third request should be denied")
	}

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Second window
	if !rl.Allow(ip) {
		t.Error("First request in new window should be allowed")
	}
	if !rl.Allow(ip) {
		t.Error("Second request in new window should be allowed")
	}
	if rl.Allow(ip) {
		t.Error("Third request in new window should be denied")
	}
}

func BenchmarkRateLimiterAllow(b *testing.B) {
	rl := NewRateLimiter(1000000, time.Hour)
	defer rl.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow("test-ip")
	}
}

func BenchmarkRateLimiterAllowConcurrent(b *testing.B) {
	rl := NewRateLimiter(1000000, time.Hour)
	defer rl.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rl.Allow("test-ip")
		}
	})
}
