package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/inst-dev/webhook/internal/config"
	"github.com/inst-dev/webhook/internal/redis"
	"github.com/inst-dev/webhook/internal/service"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	cfg *config.Config
	rdb *redis.Client
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(cfg *config.Config, rdb *redis.Client) *AuthMiddleware {
	return &AuthMiddleware{cfg: cfg, rdb: rdb}
}

// Authenticate validates the JWT token from Authorization header or cookie
func (m *AuthMiddleware) Authenticate(c *fiber.Ctx) error {
	var tokenString string

	// Try Authorization header first
	authHeader := c.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenString = strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Fall back to cookie
	if tokenString == "" {
		tokenString = c.Cookies("access_token")
	}

	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "Missing authentication token",
		})
	}

	// Check if token is blacklisted
	ctx := c.Context()
	blacklisted, err := m.rdb.Exists(ctx, "blacklist:"+tokenString).Result()
	if err == nil && blacklisted > 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "Token has been revoked",
		})
	}

	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(m.cfg.JWT.Secret), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "Invalid or expired token",
		})
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "Invalid token claims",
		})
	}

	// Validate expiry
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "Token expired",
		})
	}

	// Set user ID in context
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "Invalid token subject",
		})
	}

	c.Locals("userID", userID)
	c.Locals("tokenExp", claims.ExpiresAt)

	return c.Next()
}

// APIKeyMiddleware handles API key authentication
type APIKeyMiddleware struct {
	apiKeyService *service.APIKeyService
}

// NewAPIKeyMiddleware creates a new API key middleware
func NewAPIKeyMiddleware(apiKeyService *service.APIKeyService) *APIKeyMiddleware {
	return &APIKeyMiddleware{apiKeyService: apiKeyService}
}

// Authenticate validates the API key from header
func (m *APIKeyMiddleware) Authenticate(c *fiber.Ctx) error {
	apiKey := c.Get("X-API-Key")
	if apiKey == "" {
		apiKey = c.Query("api_key")
	}

	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "Missing API key",
		})
	}

	key, err := m.apiKeyService.Validate(c.Context(), apiKey)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "Invalid API key",
		})
	}

	c.Locals("userID", key.UserID)
	c.Locals("apiKeyID", key.ID)
	c.Locals("apiKeyScopes", key.Scopes)

	return c.Next()
}

// RequireScope checks if the API key has the required scope
func RequireScope(scope string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		scopes, ok := c.Locals("apiKeyScopes").([]string)
		if !ok {
			return c.Next() // JWT auth, no scope check needed
		}

		for _, s := range scopes {
			if s == scope || s == "*" {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   "forbidden",
			"message": "Insufficient permissions",
		})
	}
}
