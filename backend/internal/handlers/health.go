package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/inst-dev/webhook/internal/database"
	"github.com/inst-dev/webhook/internal/redis"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db  *database.Pool
	rdb *redis.Client
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *database.Pool, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, rdb: rdb}
}

// Health returns basic health status
func (h *HealthHandler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"service": "webhook.inst.lk",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

// Ready checks all dependencies
func (h *HealthHandler) Ready(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	// Check database
	dbOK := true
	if err := h.db.Ping(ctx); err != nil {
		dbOK = false
	}

	// Check Redis
	redisOK := true
	if err := h.rdb.Ping(ctx).Err(); err != nil {
		redisOK = false
	}

	status := "ready"
	statusCode := fiber.StatusOK
	if !dbOK || !redisOK {
		status = "degraded"
		statusCode = fiber.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(fiber.Map{
		"status": status,
		"checks": fiber.Map{
			"database": dbOK,
			"redis":    redisOK,
		},
		"time": time.Now().UTC().Format(time.RFC3339),
	})
}
