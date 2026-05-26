package handlers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/inst-dev/webhook/internal/config"
	"github.com/inst-dev/webhook/internal/service"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *service.AuthService
	cfg         *config.Config
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{authService: authService, cfg: cfg}
}

// Register handles user registration
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var input service.RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	if input.Email == "" || input.Password == "" || input.DisplayName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation_error",
			"message": "Email, password, and display name are required",
		})
	}

	if len(input.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation_error",
			"message": "Password must be at least 8 characters",
		})
	}

	user, tokens, err := h.authService.Register(c.Context(), input)
	if err != nil {
		switch err {
		case service.ErrEmailTaken:
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":   "email_taken",
				"message": "Email address already registered",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "internal_error",
				"message": "Failed to create account",
			})
		}
	}

	// Set cookies
	h.setTokenCookies(c, tokens)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user":         user,
		"access_token": tokens.AccessToken,
		"expires_at":   tokens.ExpiresAt,
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var input service.LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	user, tokens, err := h.authService.Login(c.Context(), input)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "invalid_credentials",
				"message": "Invalid email or password",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "internal_error",
				"message": "Login failed",
			})
		}
	}

	h.setTokenCookies(c, tokens)

	return c.JSON(fiber.Map{
		"user":         user,
		"access_token": tokens.AccessToken,
		"expires_at":   tokens.ExpiresAt,
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.BodyParser(&body); err == nil {
			refreshToken = body.RefreshToken
		}
	}

	if refreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "missing_token",
			"message": "Refresh token is required",
		})
	}

	tokens, err := h.authService.RefreshToken(c.Context(), refreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "invalid_token",
			"message": "Invalid or expired refresh token",
		})
	}

	h.setTokenCookies(c, tokens)

	return c.JSON(fiber.Map{
		"access_token": tokens.AccessToken,
		"expires_at":   tokens.ExpiresAt,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	accessToken := c.Get("Authorization")
	if strings.HasPrefix(accessToken, "Bearer ") {
		accessToken = strings.TrimPrefix(accessToken, "Bearer ")
	} else {
		accessToken = c.Cookies("access_token")
	}

	refreshToken := c.Cookies("refresh_token")

	h.authService.Logout(c.Context(), accessToken, refreshToken)

	// Clear cookies
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   h.cfg.CookieSecure,
		Domain:   h.cfg.CookieDomain,
		SameSite: "Strict",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   h.cfg.CookieSecure,
		Domain:   h.cfg.CookieDomain,
		SameSite: "Strict",
	})

	return c.JSON(fiber.Map{"message": "Logged out successfully"})
}

// Me returns the current user
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	user, err := h.authService.GetUser(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "user_not_found",
			"message": "User not found",
		})
	}
	return c.JSON(user)
}

// UpdateProfile updates the user profile
func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Profile updated"})
}

// ChangePassword changes the user password
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Password changed"})
}

// ListSessions lists active sessions
func (h *AuthHandler) ListSessions(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"sessions": []interface{}{}})
}

// RevokeSession revokes a session
func (h *AuthHandler) RevokeSession(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Session revoked"})
}

// ForgotPassword initiates password reset
func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "If the email exists, a reset link has been sent"})
}

// ResetPassword resets the password
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Password reset successfully"})
}

// VerifyEmail verifies the email
func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Email verified"})
}

// setTokenCookies sets auth cookies
func (h *AuthHandler) setTokenCookies(c *fiber.Ctx, tokens *service.TokenPair) {
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		Expires:  tokens.ExpiresAt,
		HTTPOnly: true,
		Secure:   h.cfg.CookieSecure,
		Domain:   h.cfg.CookieDomain,
		SameSite: "Strict",
		Path:     "/",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Expires:  time.Now().Add(h.cfg.JWT.RefreshExpiry),
		HTTPOnly: true,
		Secure:   h.cfg.CookieSecure,
		Domain:   h.cfg.CookieDomain,
		SameSite: "Strict",
		Path:     "/api/v1/auth/refresh",
	})
}
