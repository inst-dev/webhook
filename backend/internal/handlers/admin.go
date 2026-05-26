package handlers

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/inst-dev/webhook/internal/config"
	"github.com/inst-dev/webhook/internal/database"
	"github.com/inst-dev/webhook/internal/redis"
	"github.com/inst-dev/webhook/internal/websocket"
)

// AdminHandler handles admin console endpoints
type AdminHandler struct {
	cfg *config.Config
	db  *database.Pool
	rdb *redis.Client
	ws  *websocket.Hub
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(cfg *config.Config, db *database.Pool, rdb *redis.Client, ws *websocket.Hub) *AdminHandler {
	return &AdminHandler{cfg: cfg, db: db, rdb: rdb, ws: ws}
}

// Dashboard returns the main admin dashboard metrics
func (h *AdminHandler) Dashboard(c *fiber.Ctx) error {
	ctx := c.Context()

	stats, err := h.getSystemStats(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "stats_failed",
			"message": "Failed to retrieve system stats",
		})
	}

	return c.JSON(stats)
}

// RealtimeMetrics returns current realtime counters
func (h *AdminHandler) RealtimeMetrics(c *fiber.Ctx) error {
	ctx := c.Context()
	now := time.Now()

	// Get current counters from Redis
	secKey := fmt.Sprintf("stats:req:global:sec:%d", now.Unix())
	minKey := "stats:req:global:min:" + now.Format("200601021504")
	hourKey := "stats:req:global:hour:" + now.Format("2006010215")
	dayKey := "stats:req:global:day:" + now.Format("20060102")
	monthKey := "stats:req:global:month:" + now.Format("200601")

	reqSec, _ := h.rdb.Get(ctx, secKey).Int64()
	reqMin, _ := h.rdb.Get(ctx, minKey).Int64()
	reqHour, _ := h.rdb.Get(ctx, hourKey).Int64()
	reqDay, _ := h.rdb.Get(ctx, dayKey).Int64()
	reqMonth, _ := h.rdb.Get(ctx, monthKey).Int64()

	return c.JSON(fiber.Map{
		"requests_per_second": reqSec,
		"requests_per_minute": reqMin,
		"requests_per_hour":   reqHour,
		"requests_per_day":    reqDay,
		"requests_per_month":  reqMonth,
		"active_websockets":   h.ws.ClientCount(),
		"timestamp":           now.UTC().Format(time.RFC3339),
	})
}

// UsersList returns paginated user list
func (h *AdminHandler) UsersList(c *fiber.Ctx) error {
	ctx := c.Context()
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)
	search := c.Query("search")

	var query string
	var args []interface{}

	if search != "" {
		query = `
			SELECT id, email, display_name, plan, email_verified, created_at, last_login_at
			FROM users
			WHERE email ILIKE $1 OR display_name ILIKE $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{"%" + search + "%", limit, offset}
	} else {
		query = `
			SELECT id, email, display_name, plan, email_verified, created_at, last_login_at
			FROM users
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to query users",
		})
	}
	defer rows.Close()

	var users []fiber.Map
	for rows.Next() {
		var id uuid.UUID
		var email, displayName, plan string
		var emailVerified bool
		var createdAt time.Time
		var lastLoginAt *time.Time

		if err := rows.Scan(&id, &email, &displayName, &plan, &emailVerified, &createdAt, &lastLoginAt); err != nil {
			continue
		}

		users = append(users, fiber.Map{
			"id":             id,
			"email":          email,
			"display_name":   displayName,
			"plan":           plan,
			"email_verified": emailVerified,
			"created_at":     createdAt,
			"last_login_at":  lastLoginAt,
		})
	}

	var total int64
	h.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total)

	return c.JSON(fiber.Map{
		"users":  users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// TopIPs returns the most active source IPs
func (h *AdminHandler) TopIPs(c *fiber.Ctx) error {
	ctx := c.Context()
	limit := c.QueryInt("limit", 20)

	rows, err := h.db.Query(ctx, `
		SELECT source_ip, COUNT(*) as request_count
		FROM requests
		WHERE created_at > NOW() - INTERVAL '24 hours'
		GROUP BY source_ip
		ORDER BY request_count DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to query top IPs",
		})
	}
	defer rows.Close()

	var ips []fiber.Map
	for rows.Next() {
		var ip string
		var count int64
		if err := rows.Scan(&ip, &count); err != nil {
			continue
		}
		ips = append(ips, fiber.Map{
			"ip":            ip,
			"request_count": count,
		})
	}

	return c.JSON(fiber.Map{"top_ips": ips})
}

// SecurityLogs returns recent security-related events
func (h *AdminHandler) SecurityLogs(c *fiber.Ctx) error {
	ctx := c.Context()
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	rows, err := h.db.Query(ctx, `
		SELECT id, user_id, action, resource, details, ip, user_agent, created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to query audit logs",
		})
	}
	defer rows.Close()

	var logs []fiber.Map
	for rows.Next() {
		var id uuid.UUID
		var userID *uuid.UUID
		var action, resource, ip, userAgent string
		var details []byte
		var createdAt time.Time

		if err := rows.Scan(&id, &userID, &action, &resource, &details, &ip, &userAgent, &createdAt); err != nil {
			continue
		}

		logs = append(logs, fiber.Map{
			"id":         id,
			"user_id":    userID,
			"action":     action,
			"resource":   resource,
			"details":    string(details),
			"ip":         ip,
			"user_agent": userAgent,
			"created_at": createdAt,
		})
	}

	return c.JSON(fiber.Map{
		"logs":   logs,
		"limit":  limit,
		"offset": offset,
	})
}

// SystemHealth returns detailed system health info
func (h *AdminHandler) SystemHealth(c *fiber.Ctx) error {
	ctx := c.Context()

	// Database check
	dbOK := true
	var dbLatency time.Duration
	start := time.Now()
	if err := h.db.Ping(ctx); err != nil {
		dbOK = false
	}
	dbLatency = time.Since(start)

	// Redis check
	redisOK := true
	var redisLatency time.Duration
	start = time.Now()
	if err := h.rdb.Ping(ctx).Err(); err != nil {
		redisOK = false
	}
	redisLatency = time.Since(start)

	// Redis info
	redisInfo, _ := h.rdb.Info(ctx, "memory", "clients", "stats").Result()

	// Runtime stats
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	return c.JSON(fiber.Map{
		"status": func() string {
			if dbOK && redisOK {
				return "healthy"
			}
			return "degraded"
		}(),
		"services": fiber.Map{
			"database": fiber.Map{
				"status":     dbOK,
				"latency_ms": dbLatency.Milliseconds(),
			},
			"redis": fiber.Map{
				"status":     redisOK,
				"latency_ms": redisLatency.Milliseconds(),
				"info":       redisInfo,
			},
			"websocket": fiber.Map{
				"active_connections": h.ws.ClientCount(),
			},
		},
		"runtime": fiber.Map{
			"goroutines":    runtime.NumGoroutine(),
			"memory_alloc":  mem.Alloc,
			"memory_sys":    mem.Sys,
			"gc_runs":       mem.NumGC,
			"go_version":    runtime.Version(),
			"num_cpu":       runtime.NumCPU(),
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// StorageStats returns storage usage information
func (h *AdminHandler) StorageStats(c *fiber.Ctx) error {
	ctx := c.Context()

	var requestsSize, dnsLogsSize, emailLogsSize string
	var requestsCount, dnsLogsCount, emailLogsCount int64

	h.db.QueryRow(ctx, "SELECT COUNT(*), COALESCE(pg_total_relation_size('requests')::text, '0') FROM requests").Scan(&requestsCount, &requestsSize)
	h.db.QueryRow(ctx, "SELECT COUNT(*), COALESCE(pg_total_relation_size('dns_logs')::text, '0') FROM dns_logs").Scan(&dnsLogsCount, &dnsLogsSize)
	h.db.QueryRow(ctx, "SELECT COUNT(*), COALESCE(pg_total_relation_size('email_logs')::text, '0') FROM email_logs").Scan(&emailLogsCount, &emailLogsSize)

	return c.JSON(fiber.Map{
		"requests": fiber.Map{
			"count":      requestsCount,
			"table_size": requestsSize,
		},
		"dns_logs": fiber.Map{
			"count":      dnsLogsCount,
			"table_size": dnsLogsSize,
		},
		"email_logs": fiber.Map{
			"count":      emailLogsCount,
			"table_size": emailLogsSize,
		},
	})
}

// BillingStats returns billing analytics
func (h *AdminHandler) BillingStats(c *fiber.Ctx) error {
	ctx := c.Context()

	var freeCount, proCount, teamCount, enterpriseCount int64
	h.db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE plan = 'free'").Scan(&freeCount)
	h.db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE plan = 'pro'").Scan(&proCount)
	h.db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE plan = 'team'").Scan(&teamCount)
	h.db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE plan = 'enterprise'").Scan(&enterpriseCount)

	var activeSubCount int64
	h.db.QueryRow(ctx, "SELECT COUNT(*) FROM subscriptions WHERE status = 'active'").Scan(&activeSubCount)

	return c.JSON(fiber.Map{
		"plans": fiber.Map{
			"free":       freeCount,
			"pro":        proCount,
			"team":       teamCount,
			"enterprise": enterpriseCount,
		},
		"active_subscriptions": activeSubCount,
	})
}

// QueueHealth returns queue/worker health
func (h *AdminHandler) QueueHealth(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get Redis queue lengths
	cleanupLen, _ := h.rdb.LLen(ctx, "queue:cleanup").Result()
	emailLen, _ := h.rdb.LLen(ctx, "queue:email").Result()
	dnsLen, _ := h.rdb.LLen(ctx, "queue:dns").Result()

	return c.JSON(fiber.Map{
		"queues": fiber.Map{
			"cleanup": fiber.Map{"length": cleanupLen},
			"email":   fiber.Map{"length": emailLen},
			"dns":     fiber.Map{"length": dnsLen},
		},
	})
}

// getSystemStats retrieves comprehensive system statistics
func (h *AdminHandler) getSystemStats(ctx context.Context) (fiber.Map, error) {
	now := time.Now()

	// User counts
	var totalUsers, activeToday int64
	h.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	h.db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE last_login_at > $1", now.Add(-24*time.Hour)).Scan(&activeToday)

	// Endpoint count
	var totalEndpoints, activeEndpoints int64
	h.db.QueryRow(ctx, "SELECT COUNT(*) FROM endpoints").Scan(&totalEndpoints)
	h.db.QueryRow(ctx, "SELECT COUNT(*) FROM endpoints WHERE is_active = true").Scan(&activeEndpoints)

	// Request counts
	var totalRequests, requestsToday int64
	h.db.QueryRow(ctx, "SELECT COUNT(*) FROM requests").Scan(&totalRequests)
	h.db.QueryRow(ctx, "SELECT COUNT(*) FROM requests WHERE created_at > $1", now.Truncate(24*time.Hour)).Scan(&requestsToday)

	// DNS count
	var dnsToday int64
	h.db.QueryRow(ctx, "SELECT COUNT(*) FROM dns_logs WHERE created_at > $1", now.Truncate(24*time.Hour)).Scan(&dnsToday)

	// Email count
	var emailsToday int64
	h.db.QueryRow(ctx, "SELECT COUNT(*) FROM email_logs WHERE created_at > $1", now.Truncate(24*time.Hour)).Scan(&emailsToday)

	return fiber.Map{
		"users": fiber.Map{
			"total":        totalUsers,
			"active_today": activeToday,
		},
		"endpoints": fiber.Map{
			"total":  totalEndpoints,
			"active": activeEndpoints,
		},
		"requests": fiber.Map{
			"total": totalRequests,
			"today": requestsToday,
		},
		"dns": fiber.Map{
			"today": dnsToday,
		},
		"emails": fiber.Map{
			"today": emailsToday,
		},
		"websockets": fiber.Map{
			"active": h.ws.ClientCount(),
		},
		"timestamp": now.UTC().Format(time.RFC3339),
	}, nil
}
