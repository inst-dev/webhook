package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/inst-dev/webhook/internal/models"
	"github.com/inst-dev/webhook/internal/service"
)

// EndpointHandler handles endpoint CRUD operations
type EndpointHandler struct {
	endpointService *service.EndpointService
}

// NewEndpointHandler creates a new endpoint handler
func NewEndpointHandler(endpointService *service.EndpointService) *EndpointHandler {
	return &EndpointHandler{endpointService: endpointService}
}

// Create creates a new webhook endpoint
func (h *EndpointHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	var input service.CreateEndpointInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	endpoint, err := h.endpointService.Create(c.Context(), userID, input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "creation_failed",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(endpoint)
}

// List lists user endpoints
func (h *EndpointHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	endpoints, total, err := h.endpointService.List(c.Context(), userID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "list_failed",
			"message": "Failed to list endpoints",
		})
	}

	return c.JSON(fiber.Map{
		"endpoints": endpoints,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// Get gets a single endpoint
func (h *EndpointHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_id",
			"message": "Invalid endpoint ID",
		})
	}

	endpoint, err := h.endpointService.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "not_found",
			"message": "Endpoint not found",
		})
	}

	return c.JSON(endpoint)
}

// Update updates an endpoint
func (h *EndpointHandler) Update(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_id",
			"message": "Invalid endpoint ID",
		})
	}

	var input service.CreateEndpointInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	endpoint, err := h.endpointService.Update(c.Context(), userID, id, input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "update_failed",
			"message": err.Error(),
		})
	}

	return c.JSON(endpoint)
}

// Delete deletes an endpoint
func (h *EndpointHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_id",
			"message": "Invalid endpoint ID",
		})
	}

	if err := h.endpointService.Delete(c.Context(), userID, id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "delete_failed",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"message": "Endpoint deleted"})
}

// SetCustomResponse sets the custom response for an endpoint
func (h *EndpointHandler) SetCustomResponse(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_id",
			"message": "Invalid endpoint ID",
		})
	}

	var response models.CustomResponse
	if err := c.BodyParser(&response); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	if err := h.endpointService.SetCustomResponse(c.Context(), userID, id, &response); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "update_failed",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"message": "Custom response set"})
}
