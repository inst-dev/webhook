package websocket

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Client represents a WebSocket client connection
type Client struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	EndpointID uuid.UUID
	Send       chan []byte
	Hub        *Hub
	mu         sync.Mutex
	closed     bool
}

// Close safely closes the client
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		c.closed = true
		close(c.Send)
	}
}

// Hub manages WebSocket connections
type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
	mu         sync.RWMutex
	done       chan struct{}
}

// BroadcastMessage is a message to broadcast to specific endpoint subscribers
type BroadcastMessage struct {
	EndpointID uuid.UUID
	UserID     uuid.UUID
	Data       []byte
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		broadcast:  make(chan *BroadcastMessage, 1024),
		done:       make(chan struct{}),
	}
}

// Run starts the hub event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Debug().
				Str("client_id", client.ID.String()).
				Str("user_id", client.UserID.String()).
				Msg("WebSocket client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()
			log.Debug().
				Str("client_id", client.ID.String()).
				Msg("WebSocket client disconnected")

		case msg := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				if client.EndpointID == msg.EndpointID || client.UserID == msg.UserID {
					select {
					case client.Send <- msg.Data:
					default:
						// Client buffer full, disconnect
						go func(c *Client) {
							h.unregister <- c
						}(client)
					}
				}
			}
			h.mu.RUnlock()

		case <-h.done:
			h.mu.Lock()
			for client := range h.clients {
				client.Close()
				delete(h.clients, client)
			}
			h.mu.Unlock()
			return
		}
	}
}

// Register registers a new client
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// BroadcastToEndpoint sends a message to all clients subscribed to an endpoint
func (h *Hub) BroadcastToEndpoint(endpointID uuid.UUID, msgType string, payload interface{}) {
	msg := Message{
		Type:    msgType,
		Payload: payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal WebSocket message")
		return
	}

	h.broadcast <- &BroadcastMessage{
		EndpointID: endpointID,
		Data:       data,
	}
}

// BroadcastToUser sends a message to all clients of a user
func (h *Hub) BroadcastToUser(userID uuid.UUID, msgType string, payload interface{}) {
	msg := Message{
		Type:    msgType,
		Payload: payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal WebSocket message")
		return
	}

	h.broadcast <- &BroadcastMessage{
		UserID: userID,
		Data:   data,
	}
}

// ClientCount returns the number of connected clients
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Shutdown gracefully shuts down the hub
func (h *Hub) Shutdown() {
	close(h.done)
}
