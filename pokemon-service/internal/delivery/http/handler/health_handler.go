package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthChecker — dependency eksternal yang dicek oleh readiness probe (misalnya database).
type HealthChecker interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	db      HealthChecker
	timeout time.Duration
}

func NewHealthHandler(db HealthChecker) *HealthHandler {
	return &HealthHandler{db: db, timeout: 2 * time.Second}
}

func (h *HealthHandler) Register(r gin.IRoutes) {
	r.GET("/health/live", h.Live)
	r.GET("/health/ready", h.Ready)
	// /health dipertahankan sebagai alias readiness.
	r.GET("/health", h.Ready)
}

// Live — proses hidup dan bisa melayani HTTP. Tidak mengecek dependency.
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Ready — siap menerima traffic: database bisa dihubungi.
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		_ = c.Error(err) // dicatat oleh middleware logger
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"checks": gin.H{"database": "down"},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"checks": gin.H{"database": "up"},
	})
}
