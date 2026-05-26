package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/inst-dev/webhook/internal/service"
)

// APIKeyHandler handles API key operations
type APIKeyHandler struct {
	apiKeyService *service.APIKeyService
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(apiKeyService *service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{apiKeyService: apiKeyService}
}

// Create creates a new API key
func (h *APIKeyHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	var input service.CreateAPIKeyInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	result, err := h.apiKeyService.Create(c.Context(), userID, input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "creation_failed",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// List lists user API keys
func (h *APIKeyHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	keys, err := h.apiKeyService.List(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "list_failed",
			"message": "Failed to list API keys",
		})
	}

	return c.JSON(fiber.Map{"api_keys": keys})
}

// Revoke revokes an API key
func (h *APIKeyHandler) Revoke(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_id",
			"message": "Invalid API key ID",
		})
	}

	if err := h.apiKeyService.Revoke(c.Context(), userID, id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "revoke_failed",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"message": "API key revoked"})
}
