package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/inst-dev/webhook/internal/repository"
	"github.com/inst-dev/webhook/internal/service"
)

// RequestHandler handles request inspection
type RequestHandler struct {
	requestService *service.RequestService
}

// NewRequestHandler creates a new request handler
func NewRequestHandler(requestService *service.RequestService) *RequestHandler {
	return &RequestHandler{requestService: requestService}
}

// List lists captured requests for an endpoint
func (h *RequestHandler) List(c *fiber.Ctx) error {
	endpointID, err := uuid.Parse(c.Params("endpointId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_id",
			"message": "Invalid endpoint ID",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	search := c.Query("search")

	opts := repository.ListOptions{
		Limit:  limit,
		Offset: offset,
	}

	var requests []*service.RequestService
	var total int64

	if search != "" {
		reqs, t, err := h.requestService.Search(c.Context(), endpointID, search, opts)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "search_failed",
				"message": "Failed to search requests",
			})
		}
		_ = reqs
		_ = t
		return c.JSON(fiber.Map{
			"requests": reqs,
			"total":    t,
			"limit":    limit,
			"offset":   offset,
		})
	}

	reqs, total, err := h.requestService.List(c.Context(), endpointID, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "list_failed",
			"message": "Failed to list requests",
		})
	}

	_ = requests
	_ = total

	return c.JSON(fiber.Map{
		"requests": reqs,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// Get gets a single captured request
func (h *RequestHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_id",
			"message": "Invalid request ID",
		})
	}

	req, err := h.requestService.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "not_found",
			"message": "Request not found",
		})
	}

	return c.JSON(req)
}

// Delete deletes a captured request
func (h *RequestHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_id",
			"message": "Invalid request ID",
		})
	}

	if err := h.requestService.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "delete_failed",
			"message": "Failed to delete request",
		})
	}

	return c.JSON(fiber.Map{"message": "Request deleted"})
}

// Replay replays a captured request
func (h *RequestHandler) Replay(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_id",
			"message": "Invalid request ID",
		})
	}

	var input service.ReplayInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	result, err := h.requestService.Replay(c.Context(), id, input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "replay_failed",
			"message": err.Error(),
		})
	}

	return c.JSON(result)
}
