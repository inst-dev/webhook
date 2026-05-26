package handlers

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/inst-dev/webhook/internal/models"
	"github.com/inst-dev/webhook/internal/service"
)

// WebhookHandler handles incoming webhook requests
type WebhookHandler struct {
	requestService  *service.RequestService
	endpointService *service.EndpointService
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(requestService *service.RequestService, endpointService *service.EndpointService) *WebhookHandler {
	return &WebhookHandler{
		requestService:  requestService,
		endpointService: endpointService,
	}
}

// Capture captures an incoming webhook request
func (h *WebhookHandler) Capture(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "not_found",
			"message": "Endpoint not found",
		})
	}

	// Look up endpoint by token
	endpoint, err := h.endpointService.GetByToken(c.Context(), token)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "not_found",
			"message": "Endpoint not found",
		})
	}

	// Parse headers
	headers := make(map[string]string)
	c.Request().Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})
	headersJSON, _ := json.Marshal(headers)

	// Parse query params
	queryParams := make(map[string]string)
	c.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
		queryParams[string(key)] = string(value)
	})
	queryParamsJSON, _ := json.Marshal(queryParams)

	// Build request model
	req := &models.Request{
		EndpointID:    endpoint.ID,
		Method:        c.Method(),
		Path:          c.Path(),
		Headers:       headersJSON,
		QueryParams:   queryParamsJSON,
		Body:          c.Body(),
		ContentType:   c.Get("Content-Type"),
		ContentLength: int64(len(c.Body())),
		SourceIP:      c.IP(),
		UserAgent:     c.Get("User-Agent"),
	}

	// Store the request
	if err := h.requestService.CaptureRequest(c.Context(), endpoint, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "capture_failed",
			"message": "Failed to capture request",
		})
	}

	// Return custom response if configured
	if endpoint.CustomResponse != nil {
		cr := endpoint.CustomResponse

		// Apply delay
		if cr.Delay > 0 {
			time.Sleep(time.Duration(cr.Delay) * time.Millisecond)
		}

		// Handle redirect
		if cr.Redirect != "" {
			return c.Redirect(cr.Redirect, cr.StatusCode)
		}

		// Set custom headers
		for key, value := range cr.Headers {
			c.Set(key, value)
		}

		// Return custom body
		statusCode := cr.StatusCode
		if statusCode == 0 {
			statusCode = 200
		}

		if cr.Body != "" {
			return c.Status(statusCode).SendString(cr.Body)
		}

		return c.SendStatus(statusCode)
	}

	// Default response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":    true,
		"message":    "Request captured",
		"request_id": req.ID,
		"timestamp":  req.CreatedAt,
	})
}
