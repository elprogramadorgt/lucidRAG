package health

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DBPinger defines the interface for database health checks.
type DBPinger interface {
	Ping(ctx context.Context) error
}

// Handler handles health check endpoints.
type Handler struct {
	db DBPinger
}

// NewHandler creates a new health handler.
func NewHandler(db DBPinger) *Handler {
	return &Handler{db: db}
}

// Liveness returns OK if the service is running.
func (h *Handler) Liveness(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Readiness returns OK if all dependencies are available.
func (h *Handler) Readiness(ctx *gin.Context) {
	if err := h.db.Ping(ctx.Request.Context()); err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": "error"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Register registers health check routes on the router.
func Register(r *gin.Engine, h *Handler) {
	r.GET("/healthz", h.Liveness)
	r.GET("/readyz", h.Readiness)
}
