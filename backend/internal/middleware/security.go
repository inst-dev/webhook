package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/inst-dev/webhook/internal/config"
	"github.com/inst-dev/webhook/internal/redis"
)

// CSRFMiddleware handles CSRF protection
type CSRFMiddleware struct {
	cfg *config.Config
}

// NewCSRFMiddleware creates a new CSRF middleware
func NewCSRFMiddleware(cfg *config.Config) *CSRFMiddleware {
	return &CSRFMiddleware{cfg: cfg}
}

// Protect validates CSRF tokens for state-changing requests
func (m *CSRFMiddleware) Protect(c *fiber.Ctx) error {
	method := c.Method()

	// Skip CSRF check for safe methods
	if method == "GET" || method == "HEAD" || method == "OPTIONS" {
		return c.Next()
	}

	// Get CSRF token from header
	token := c.Get("X-CSRF-Token")
	if token == "" {
		token = c.FormValue("_csrf")
	}

	if token == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   "csrf_error",
			"message": "Missing CSRF token",
		})
	}

	// Validate CSRF token
	cookieToken := c.Cookies("csrf_token")
	if cookieToken == "" || !validateCSRFToken(token, cookieToken, m.cfg.CSRFSecret) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   "csrf_error",
			"message": "Invalid CSRF token",
		})
	}

	return c.Next()
}

// validateCSRFToken validates the CSRF token against the cookie
func validateCSRFToken(token, cookieToken, secret string) bool {
	expected := generateCSRFSignature(cookieToken, secret)
	return hmac.Equal([]byte(token), []byte(expected))
}

// generateCSRFSignature creates HMAC signature
func generateCSRFSignature(token, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}

// BruteForceMiddleware protects against brute force attacks
type BruteForceMiddleware struct {
	rdb       *redis.Client
	maxAttempts int
	window      time.Duration
	blockDuration time.Duration
}

// NewBruteForceMiddleware creates a new brute force middleware
func NewBruteForceMiddleware(rdb *redis.Client) *BruteForceMiddleware {
	return &BruteForceMiddleware{
		rdb:           rdb,
		maxAttempts:   5,
		window:        15 * time.Minute,
		blockDuration: 30 * time.Minute,
	}
}

// Protect checks if the IP is blocked
func (m *BruteForceMiddleware) Protect(c *fiber.Ctx) error {
	ip := c.IP()
	key := "bruteforce:" + ip

	// Check if IP is blocked
	blocked, err := m.rdb.Exists(c.Context(), "blocked:"+ip).Result()
	if err == nil && blocked > 0 {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error":   "blocked",
			"message": "Too many failed attempts. Please try again later.",
		})
	}

	// Increment attempt counter
	count, err := m.rdb.IncrCounter(c.Context(), key, m.window)
	if err == nil && count > int64(m.maxAttempts) {
		// Block the IP
		m.rdb.Set(c.Context(), "blocked:"+ip, "1", m.blockDuration)
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error":   "blocked",
			"message": "Too many failed attempts. Please try again later.",
		})
	}

	return c.Next()
}

// SSRFProtection validates URLs to prevent SSRF attacks
func SSRFProtection() fiber.Handler {
	// Blocked IP ranges
	blockedPrefixes := []string{
		"127.", "10.", "172.16.", "172.17.", "172.18.", "172.19.",
		"172.20.", "172.21.", "172.22.", "172.23.", "172.24.",
		"172.25.", "172.26.", "172.27.", "172.28.", "172.29.",
		"172.30.", "172.31.", "192.168.", "169.254.", "0.",
		"::1", "fc00:", "fe80:", "fd",
	}

	blockedHosts := []string{
		"localhost", "metadata.google.internal",
		"169.254.169.254", "metadata.internal",
	}

	return func(c *fiber.Ctx) error {
		targetURL := c.Query("url", "")
		if targetURL == "" {
			return c.Next()
		}

		lower := strings.ToLower(targetURL)

		for _, prefix := range blockedPrefixes {
			if strings.Contains(lower, prefix) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error":   "ssrf_blocked",
					"message": "Access to internal resources is blocked",
				})
			}
		}

		for _, host := range blockedHosts {
			if strings.Contains(lower, host) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error":   "ssrf_blocked",
					"message": "Access to internal resources is blocked",
				})
			}
		}

		return c.Next()
	}
}
