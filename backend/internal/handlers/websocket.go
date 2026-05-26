package handlers

import (
	"github.com/gofiber/fiber/v2"
	fiberws "github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
	"github.com/inst-dev/webhook/internal/service"
	ws "github.com/inst-dev/webhook/internal/websocket"
	"github.com/rs/zerolog/log"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub         *ws.Hub
	authService *service.AuthService
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *ws.Hub, authService *service.AuthService) *WebSocketHandler {
	return &WebSocketHandler{hub: hub, authService: authService}
}

// Upgrade handles WebSocket upgrade requests
func (h *WebSocketHandler) Upgrade(c *fiber.Ctx) error {
	if !fiberws.IsWebSocketUpgrade(c) {
		return c.Status(fiber.StatusUpgradeRequired).JSON(fiber.Map{
			"error":   "upgrade_required",
			"message": "WebSocket upgrade required",
		})
	}

	// Authenticate via query param token
	token := c.Query("token")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "Authentication token required",
		})
	}

	userID, err := h.authService.ValidateToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "Invalid authentication token",
		})
	}

	endpointIDStr := c.Query("endpoint_id")
	var endpointID uuid.UUID
	if endpointIDStr != "" {
		endpointID, _ = uuid.Parse(endpointIDStr)
	}

	return fiberws.New(func(conn *fiberws.Conn) {
		client := &ws.Client{
			ID:         uuid.New(),
			UserID:     userID,
			EndpointID: endpointID,
			Send:       make(chan []byte, 256),
			Hub:        h.hub,
		}

		h.hub.Register(client)
		defer h.hub.Unregister(client)

		// Writer goroutine
		go func() {
			for msg := range client.Send {
				if err := conn.WriteMessage(fiberws.TextMessage, msg); err != nil {
					log.Debug().Err(err).Msg("WebSocket write error")
					return
				}
			}
		}()

		// Reader loop (handle ping/pong and subscription changes)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}

			// Handle subscription messages
			h.handleMessage(client, msg)
		}
	})(c)
}

// handleMessage processes incoming WebSocket messages
func (h *WebSocketHandler) handleMessage(client *ws.Client, msg []byte) {
	// Parse message and handle subscription changes
	// For now, just log
	log.Debug().
		Str("client_id", client.ID.String()).
		Bytes("message", msg).
		Msg("WebSocket message received")
}
