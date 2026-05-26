package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/inst-dev/webhook/internal/database"
)

// AdminMiddleware ensures the user has admin privileges
type AdminMiddleware struct {
	db *database.Pool
}

// NewAdminMiddleware creates a new admin middleware
func NewAdminMiddleware(db *database.Pool) *AdminMiddleware {
	return &AdminMiddleware{db: db}
}

// RequireAdmin checks if the authenticated user is an admin
func (m *AdminMiddleware) RequireAdmin(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "Authentication required",
		})
	}

	// Check if user has admin role
	var plan string
	err := m.db.QueryRow(c.Context(), "SELECT plan FROM users WHERE id = $1", userID).Scan(&plan)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   "forbidden",
			"message": "Access denied",
		})
	}

	// Only enterprise users or designated admins can access
	// In production, use a separate admin flag or role table
	if plan != "enterprise" && plan != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   "forbidden",
			"message": "Admin access required",
		})
	}

	return c.Next()
}
