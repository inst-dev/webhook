package server

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/inst-dev/webhook/internal/config"
	"github.com/inst-dev/webhook/internal/database"
	"github.com/inst-dev/webhook/internal/handlers"
	"github.com/inst-dev/webhook/internal/middleware"
	"github.com/inst-dev/webhook/internal/redis"
	"github.com/inst-dev/webhook/internal/repository"
	"github.com/inst-dev/webhook/internal/service"
	"github.com/inst-dev/webhook/internal/websocket"
)

// Server represents the HTTP server
type Server struct {
	app *fiber.App
	cfg *config.Config
	db  *database.Pool
	rdb *redis.Client
	ws  *websocket.Hub
}

// New creates a new server instance
func New(cfg *config.Config, db *database.Pool, rdb *redis.Client) *Server {
	app := fiber.New(fiber.Config{
		AppName:               cfg.AppName,
		ServerHeader:          "",
		DisableStartupMessage: cfg.AppEnv == "production",
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           120 * time.Second,
		BodyLimit:             int(cfg.PayloadMaxPremium),
		EnablePrintRoutes:     cfg.AppDebug,
	})

	s := &Server{
		app: app,
		cfg: cfg,
		db:  db,
		rdb: rdb,
		ws:  websocket.NewHub(),
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware configures global middleware
func (s *Server) setupMiddleware() {
	// Recovery from panics
	s.app.Use(recover.New())

	// Request ID
	s.app.Use(requestid.New())

	// Logger
	if s.cfg.AppEnv != "test" {
		s.app.Use(logger.New(logger.Config{
			Format:     "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path}\n",
			TimeFormat: "2006-01-02 15:04:05",
		}))
	}

	// Compression
	s.app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	// Security headers
	s.app.Use(helmet.New())

	// CORS
	s.app.Use(cors.New(cors.Config{
		AllowOrigins:     s.cfg.CORSOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-CSRF-Token,X-Request-ID",
		AllowCredentials: true,
		ExposeHeaders:    "X-Request-ID,X-RateLimit-Limit,X-RateLimit-Remaining",
		MaxAge:           86400,
	}))

	// Global rate limiter
	s.app.Use(limiter.New(limiter.Config{
		Max:        s.cfg.RateLimitRequests,
		Expiration: time.Duration(s.cfg.RateLimitWindow) * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests, please try again later",
			})
		},
	}))
}

// setupRoutes configures all routes
func (s *Server) setupRoutes() {
	// Initialize repositories
	userRepo := repository.NewUserRepository(s.db)
	endpointRepo := repository.NewEndpointRepository(s.db)
	requestRepo := repository.NewRequestRepository(s.db)
	apiKeyRepo := repository.NewAPIKeyRepository(s.db)

	// Initialize services
	authService := service.NewAuthService(s.cfg, userRepo, s.rdb)
	endpointService := service.NewEndpointService(s.cfg, endpointRepo, s.rdb)
	requestService := service.NewRequestService(s.cfg, requestRepo, endpointRepo, s.rdb, s.ws)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, s.rdb)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(s.db, s.rdb)
	authHandler := handlers.NewAuthHandler(authService, s.cfg)
	endpointHandler := handlers.NewEndpointHandler(endpointService)
	requestHandler := handlers.NewRequestHandler(requestService)
	webhookHandler := handlers.NewWebhookHandler(requestService, endpointService)
	wsHandler := handlers.NewWebSocketHandler(s.ws, authService)

	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(s.cfg, s.rdb)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(apiKeyService)

	// Health check
	s.app.Get("/health", healthHandler.Health)
	s.app.Get("/ready", healthHandler.Ready)

	// API v1 routes
	v1 := s.app.Group("/api/v1")

	// Auth routes
	auth := v1.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)
	auth.Post("/logout", authHandler.Logout)
	auth.Post("/forgot-password", authHandler.ForgotPassword)
	auth.Post("/reset-password", authHandler.ResetPassword)
	auth.Post("/verify-email", authHandler.VerifyEmail)

	// Protected routes
	protected := v1.Group("", authMiddleware.Authenticate)

	// User routes
	protected.Get("/me", authHandler.Me)
	protected.Put("/me", authHandler.UpdateProfile)
	protected.Put("/me/password", authHandler.ChangePassword)
	protected.Get("/me/sessions", authHandler.ListSessions)
	protected.Delete("/me/sessions/:id", authHandler.RevokeSession)

	// Endpoint routes
	endpoints := protected.Group("/endpoints")
	endpoints.Post("/", endpointHandler.Create)
	endpoints.Get("/", endpointHandler.List)
	endpoints.Get("/:id", endpointHandler.Get)
	endpoints.Put("/:id", endpointHandler.Update)
	endpoints.Delete("/:id", endpointHandler.Delete)
	endpoints.Put("/:id/response", endpointHandler.SetCustomResponse)

	// Request routes
	requests := protected.Group("/endpoints/:endpointId/requests")
	requests.Get("/", requestHandler.List)
	requests.Get("/:id", requestHandler.Get)
	requests.Delete("/:id", requestHandler.Delete)
	requests.Post("/:id/replay", requestHandler.Replay)

	// API Key routes
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyService)
	apiKeys := protected.Group("/api-keys")
	apiKeys.Post("/", apiKeyHandler.Create)
	apiKeys.Get("/", apiKeyHandler.List)
	apiKeys.Delete("/:id", apiKeyHandler.Revoke)

	// Public API with API key auth
	publicAPI := v1.Group("/public", apiKeyMiddleware.Authenticate)
	publicAPI.Get("/endpoints", endpointHandler.List)
	publicAPI.Get("/endpoints/:id/requests", requestHandler.List)

	// Admin routes (requires admin role)
	adminHandler := handlers.NewAdminHandler(s.cfg, s.db, s.rdb, s.ws)
	adminMiddleware := middleware.NewAdminMiddleware(s.db)
	admin := v1.Group("/admin", authMiddleware.Authenticate, adminMiddleware.RequireAdmin)
	admin.Get("/dashboard", adminHandler.Dashboard)
	admin.Get("/metrics/realtime", adminHandler.RealtimeMetrics)
	admin.Get("/users", adminHandler.UsersList)
	admin.Get("/top-ips", adminHandler.TopIPs)
	admin.Get("/security-logs", adminHandler.SecurityLogs)
	admin.Get("/health", adminHandler.SystemHealth)
	admin.Get("/storage", adminHandler.StorageStats)
	admin.Get("/billing", adminHandler.BillingStats)
	admin.Get("/queues", adminHandler.QueueHealth)

	// Billing routes
	billingHandler := handlers.NewBillingHandler(s.cfg, s.db, s.rdb)
	billing := protected.Group("/billing")
	billing.Get("/plans", billingHandler.ListPlans)
	billing.Get("/subscription", billingHandler.GetSubscription)
	billing.Post("/subscribe", billingHandler.CreateSubscription)
	billing.Post("/cancel", billingHandler.CancelSubscription)
	billing.Get("/invoices", billingHandler.ListInvoices)
	// Payment webhooks (public - no auth)
	v1.Post("/webhooks/paypal", billingHandler.PayPalWebhook)
	v1.Post("/webhooks/payhere", billingHandler.PayHereWebhook)

	// Webhook capture endpoint (public - no auth required)
	s.app.All("/:token", webhookHandler.Capture)
	s.app.All("/:token/*", webhookHandler.Capture)

	// WebSocket upgrade
	s.app.Get("/ws", wsHandler.Upgrade)
}

// Start starts the server
func (s *Server) Start() error {
	// Start WebSocket hub
	go s.ws.Run()

	addr := fmt.Sprintf(":%d", s.cfg.APIPort)
	return s.app.Listen(addr)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.ws.Shutdown()
	return s.app.ShutdownWithContext(ctx)
}
