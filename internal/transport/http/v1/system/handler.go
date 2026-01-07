package system

import (
	"context"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/elprogramadorgt/lucidRAG/internal/domain/system"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/gin-gonic/gin"
)

// DBPinger allows checking database connectivity.
type DBPinger interface {
	Ping(ctx context.Context) error
}

// HandlerConfig contains dependencies for creating a system handler.
type HandlerConfig struct {
	Repo        system.LogRepository
	DB          DBPinger
	Log         *logger.Logger
	StartTime   time.Time
	Environment string
	Version     string
}

// Handler handles system administration HTTP requests.
type Handler struct {
	repo        system.LogRepository
	db          DBPinger
	log         *logger.Logger
	startTime   time.Time
	environment string
	version     string
}

// NewHandler creates a new system handler.
func NewHandler(cfg HandlerConfig) *Handler {
	version := cfg.Version
	if version == "" {
		version = "dev"
	}
	return &Handler{
		repo:        cfg.Repo,
		db:          cfg.DB,
		log:         cfg.Log.With("handler", "system"),
		startTime:   cfg.StartTime,
		environment: cfg.Environment,
		version:     version,
	}
}

func (h *Handler) ListLogs(ctx *gin.Context) {
	adminID := ctx.GetString("user_id")
	filter := system.LogFilter{
		Level:     ctx.Query("level"),
		Search:    ctx.Query("search"),
		RequestID: ctx.Query("request_id"),
		Source:    ctx.Query("source"),
	}

	if limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50")); limit > 0 {
		filter.Limit = limit
	}
	if offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0")); offset > 0 {
		filter.Offset = offset
	}
	if start := ctx.Query("start_time"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			filter.StartTime = t
		}
	}
	if end := ctx.Query("end_time"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			filter.EndTime = t
		}
	}

	logs, total, err := h.repo.List(ctx.Request.Context(), filter)
	if err != nil {
		h.log.Error("failed to list logs", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list logs"})
		return
	}

	h.log.Info("admin_activity", "action", "logs_view", "admin_id", adminID, "filter_level", filter.Level, "result_count", len(logs))

	ctx.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

func (h *Handler) GetStats(ctx *gin.Context) {
	adminID := ctx.GetString("user_id")
	stats, err := h.repo.Stats(ctx.Request.Context())
	if err != nil {
		h.log.Error("failed to get log stats", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	h.log.Info("admin_activity", "action", "logs_stats", "admin_id", adminID)
	ctx.JSON(http.StatusOK, stats)
}

func (h *Handler) CleanupLogs(ctx *gin.Context) {
	adminID := ctx.GetString("user_id")
	days, _ := strconv.Atoi(ctx.DefaultQuery("days", "30"))
	if days < 1 {
		days = 30
	}

	deleted, err := h.repo.DeleteOlderThan(ctx.Request.Context(), days)
	if err != nil {
		h.log.Error("failed to cleanup logs", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cleanup logs"})
		return
	}

	h.log.Info("admin_activity", "action", "logs_cleanup", "admin_id", adminID, "deleted_count", deleted, "days", days)
	ctx.JSON(http.StatusOK, gin.H{"deleted": deleted, "days": days})
}

// ServerInfo contains server status and runtime information.
type ServerInfo struct {
	Status      string            `json:"status"`
	Environment string            `json:"environment"`
	Version     string            `json:"version"`
	Uptime      string            `json:"uptime"`
	UptimeSecs  int64             `json:"uptime_seconds"`
	StartedAt   time.Time         `json:"started_at"`
	Database    DatabaseStatus    `json:"database"`
	Runtime     RuntimeInfo       `json:"runtime"`
	Endpoints   []EndpointInfo    `json:"endpoints"`
}

// DatabaseStatus contains database connectivity information.
type DatabaseStatus struct {
	Status     string `json:"status"`
	Latency    string `json:"latency,omitempty"`
	LatencyMs  int64  `json:"latency_ms,omitempty"`
}

// RuntimeInfo contains Go runtime statistics.
type RuntimeInfo struct {
	GoVersion    string `json:"go_version"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
	MemAllocMB   int64  `json:"mem_alloc_mb"`
	MemSysMB     int64  `json:"mem_sys_mb"`
}

// EndpointInfo describes an API endpoint.
type EndpointInfo struct {
	Path        string `json:"path"`
	Method      string `json:"method"`
	Description string `json:"description"`
}

func (h *Handler) GetServerInfo(ctx *gin.Context) {
	adminID := ctx.GetString("user_id")

	// Calculate uptime
	uptime := time.Since(h.startTime)
	uptimeStr := formatDuration(uptime)

	// Check database connectivity
	dbStatus := DatabaseStatus{Status: "connected"}
	dbStart := time.Now()
	if err := h.db.Ping(ctx.Request.Context()); err != nil {
		dbStatus.Status = "disconnected"
	} else {
		latency := time.Since(dbStart)
		dbStatus.Latency = latency.String()
		dbStatus.LatencyMs = latency.Milliseconds()
	}

	// Get runtime info
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	runtimeInfo := RuntimeInfo{
		GoVersion:    runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		MemAllocMB:   int64(memStats.Alloc / 1024 / 1024),
		MemSysMB:     int64(memStats.Sys / 1024 / 1024),
	}

	// API endpoints info
	endpoints := []EndpointInfo{
		{Path: "/healthz", Method: "GET", Description: "Liveness probe"},
		{Path: "/readyz", Method: "GET", Description: "Readiness probe (checks DB)"},
		{Path: "/api/v1/auth/register", Method: "POST", Description: "User registration"},
		{Path: "/api/v1/auth/login", Method: "POST", Description: "User login"},
		{Path: "/api/v1/auth/me", Method: "GET", Description: "Current user info"},
		{Path: "/api/v1/documents", Method: "GET/POST/PUT/DELETE", Description: "Document CRUD"},
		{Path: "/api/v1/conversations", Method: "GET", Description: "Conversation list"},
		{Path: "/api/v1/rag/query", Method: "POST", Description: "RAG query endpoint"},
		{Path: "/api/v1/whatsapp/webhook", Method: "GET/POST", Description: "WhatsApp webhook"},
		{Path: "/api/v1/system/logs", Method: "GET/DELETE", Description: "System logs (admin)"},
		{Path: "/api/v1/system/info", Method: "GET", Description: "Server info (admin)"},
	}

	info := ServerInfo{
		Status:      "running",
		Environment: h.environment,
		Version:     h.version,
		Uptime:      uptimeStr,
		UptimeSecs:  int64(uptime.Seconds()),
		StartedAt:   h.startTime,
		Database:    dbStatus,
		Runtime:     runtimeInfo,
		Endpoints:   endpoints,
	}

	h.log.Info("admin_activity", "action", "server_info_view", "admin_id", adminID)
	ctx.JSON(http.StatusOK, info)
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return strconv.Itoa(days) + "d " + strconv.Itoa(hours) + "h " + strconv.Itoa(minutes) + "m"
	}
	if hours > 0 {
		return strconv.Itoa(hours) + "h " + strconv.Itoa(minutes) + "m " + strconv.Itoa(seconds) + "s"
	}
	if minutes > 0 {
		return strconv.Itoa(minutes) + "m " + strconv.Itoa(seconds) + "s"
	}
	return strconv.Itoa(seconds) + "s"
}
