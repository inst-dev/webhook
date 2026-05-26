package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/inst-dev/webhook/internal/config"
	"github.com/inst-dev/webhook/internal/models"
	"github.com/inst-dev/webhook/internal/redis"
	"github.com/inst-dev/webhook/internal/repository"
	"github.com/inst-dev/webhook/internal/websocket"
	"github.com/rs/zerolog/log"
)

// RequestService handles request business logic
type RequestService struct {
	cfg          *config.Config
	requestRepo  *repository.RequestRepository
	endpointRepo *repository.EndpointRepository
	rdb          *redis.Client
	ws           *websocket.Hub
}

// NewRequestService creates a new request service
func NewRequestService(
	cfg *config.Config,
	requestRepo *repository.RequestRepository,
	endpointRepo *repository.EndpointRepository,
	rdb *redis.Client,
	ws *websocket.Hub,
) *RequestService {
	return &RequestService{
		cfg:          cfg,
		requestRepo:  requestRepo,
		endpointRepo: endpointRepo,
		rdb:          rdb,
		ws:           ws,
	}
}

// CaptureRequest captures and stores an incoming webhook request
func (s *RequestService) CaptureRequest(ctx context.Context, endpoint *models.Endpoint, req *models.Request) error {
	startTime := time.Now()

	// Store the request
	if err := s.requestRepo.Create(ctx, req); err != nil {
		return fmt.Errorf("failed to store request: %w", err)
	}

	// Increment endpoint request count
	s.endpointRepo.IncrementRequestCount(ctx, endpoint.ID)

	// Update response time
	req.ResponseTime = time.Since(startTime).Microseconds()

	// Update analytics counters
	s.incrementCounters(ctx, endpoint.ID.String())

	// Broadcast to WebSocket clients
	s.ws.BroadcastToEndpoint(endpoint.ID, "new_request", req)

	log.Debug().
		Str("endpoint_id", endpoint.ID.String()).
		Str("method", req.Method).
		Str("source_ip", req.SourceIP).
		Msg("Request captured")

	return nil
}

// List lists requests for an endpoint
func (s *RequestService) List(ctx context.Context, endpointID uuid.UUID, opts repository.ListOptions) ([]*models.Request, int64, error) {
	if opts.Limit <= 0 || opts.Limit > 100 {
		opts.Limit = 20
	}
	if opts.Offset < 0 {
		opts.Offset = 0
	}
	return s.requestRepo.ListByEndpoint(ctx, endpointID, opts)
}

// Get gets a single request by ID
func (s *RequestService) Get(ctx context.Context, id uuid.UUID) (*models.Request, error) {
	return s.requestRepo.GetByID(ctx, id)
}

// Delete deletes a request
func (s *RequestService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.requestRepo.Delete(ctx, id)
}

// Search searches requests
func (s *RequestService) Search(ctx context.Context, endpointID uuid.UUID, query string, opts repository.ListOptions) ([]*models.Request, int64, error) {
	return s.requestRepo.Search(ctx, endpointID, query, opts)
}

// ReplayInput holds replay request data
type ReplayInput struct {
	Method  string            `json:"method"`
	URL     string            `json:"url" validate:"required,url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// Replay replays a captured request
func (s *RequestService) Replay(ctx context.Context, requestID uuid.UUID, input ReplayInput) (*ReplayResult, error) {
	// Get original request if no input provided
	if input.URL == "" {
		return nil, fmt.Errorf("target URL is required")
	}

	// Build the request
	var body io.Reader
	if input.Body != "" {
		body = bytes.NewBufferString(input.Body)
	}

	method := input.Method
	if method == "" {
		method = "POST"
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, input.URL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range input.Headers {
		httpReq.Header.Set(key, value)
	}

	// Execute the request
	client := &http.Client{Timeout: 30 * time.Second}
	startTime := time.Now()
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("replay request failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)

	// Read response body
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Build response headers
	respHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			respHeaders[key] = values[0]
		}
	}

	return &ReplayResult{
		StatusCode: resp.StatusCode,
		Headers:    respHeaders,
		Body:       string(respBody),
		Duration:   duration.Milliseconds(),
	}, nil
}

// ReplayResult holds the result of a replayed request
type ReplayResult struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Duration   int64             `json:"duration_ms"`
}

// incrementCounters updates Redis analytics counters
func (s *RequestService) incrementCounters(ctx context.Context, endpointID string) {
	now := time.Now()

	keys := []string{
		fmt.Sprintf("stats:req:sec:%s:%d", endpointID, now.Unix()),
		fmt.Sprintf("stats:req:min:%s:%s", endpointID, now.Format("200601021504")),
		fmt.Sprintf("stats:req:hour:%s:%s", endpointID, now.Format("2006010215")),
		fmt.Sprintf("stats:req:day:%s:%s", endpointID, now.Format("20060102")),
		"stats:req:global:sec:" + fmt.Sprintf("%d", now.Unix()),
		"stats:req:global:min:" + now.Format("200601021504"),
		"stats:req:global:hour:" + now.Format("2006010215"),
		"stats:req:global:day:" + now.Format("20060102"),
	}

	expiries := []time.Duration{
		2 * time.Minute,
		2 * time.Hour,
		2 * 24 * time.Hour,
		2 * 30 * 24 * time.Hour,
		2 * time.Minute,
		2 * time.Hour,
		2 * 24 * time.Hour,
		2 * 30 * 24 * time.Hour,
	}

	for i, key := range keys {
		s.rdb.IncrCounter(ctx, key, expiries[i])
	}

	// Also publish event for realtime subscribers
	event, _ := json.Marshal(map[string]string{
		"type":        "request_received",
		"endpoint_id": endpointID,
		"timestamp":   now.Format(time.RFC3339),
	})
	s.rdb.PublishEvent(ctx, "events:requests", string(event))
}
