package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/inst-dev/webhook/internal/config"
	"github.com/inst-dev/webhook/internal/models"
	"github.com/inst-dev/webhook/internal/redis"
	"github.com/inst-dev/webhook/internal/repository"
)

var tokenRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// EndpointService handles endpoint business logic
type EndpointService struct {
	cfg          *config.Config
	endpointRepo *repository.EndpointRepository
	rdb          *redis.Client
}

// NewEndpointService creates a new endpoint service
func NewEndpointService(cfg *config.Config, endpointRepo *repository.EndpointRepository, rdb *redis.Client) *EndpointService {
	return &EndpointService{
		cfg:          cfg,
		endpointRepo: endpointRepo,
		rdb:          rdb,
	}
}

// CreateInput holds endpoint creation data
type CreateEndpointInput struct {
	Name        string `json:"name" validate:"max=100"`
	Description string `json:"description" validate:"max=500"`
	CustomToken string `json:"custom_token" validate:"max=64"`
}

// Create creates a new webhook endpoint
func (s *EndpointService) Create(ctx context.Context, userID uuid.UUID, input CreateEndpointInput) (*models.Endpoint, error) {
	token := input.CustomToken
	if token == "" {
		token = generateToken()
	} else {
		// Validate custom token format
		if !tokenRegex.MatchString(token) {
			return nil, fmt.Errorf("invalid token format: must be alphanumeric with hyphens and underscores")
		}
		// Check if token already exists
		_, err := s.endpointRepo.GetByToken(ctx, token)
		if err == nil {
			return nil, fmt.Errorf("token already in use")
		}
	}

	endpoint := &models.Endpoint{
		UserID:      userID,
		Token:       token,
		Name:        input.Name,
		Description: input.Description,
		IsActive:    true,
	}

	if err := s.endpointRepo.Create(ctx, endpoint); err != nil {
		return nil, fmt.Errorf("failed to create endpoint: %w", err)
	}

	// Cache the endpoint token mapping
	s.rdb.Set(ctx, "endpoint:token:"+token, endpoint.ID.String(), 0)

	return endpoint, nil
}

// Get gets an endpoint by ID
func (s *EndpointService) Get(ctx context.Context, id uuid.UUID) (*models.Endpoint, error) {
	return s.endpointRepo.GetByID(ctx, id)
}

// GetByToken gets an endpoint by token
func (s *EndpointService) GetByToken(ctx context.Context, token string) (*models.Endpoint, error) {
	// Try cache first
	cached, err := s.rdb.Get(ctx, "endpoint:token:"+token).Result()
	if err == nil {
		id, err := uuid.Parse(cached)
		if err == nil {
			return s.endpointRepo.GetByID(ctx, id)
		}
	}

	// Fall back to database
	endpoint, err := s.endpointRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Cache it
	s.rdb.Set(ctx, "endpoint:token:"+token, endpoint.ID.String(), 0)

	return endpoint, nil
}

// List lists endpoints for a user
func (s *EndpointService) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Endpoint, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.endpointRepo.ListByUser(ctx, userID, limit, offset)
}

// Update updates an endpoint
func (s *EndpointService) Update(ctx context.Context, userID, endpointID uuid.UUID, input CreateEndpointInput) (*models.Endpoint, error) {
	endpoint, err := s.endpointRepo.GetByID(ctx, endpointID)
	if err != nil {
		return nil, err
	}

	if endpoint.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	if input.Name != "" {
		endpoint.Name = input.Name
	}
	if input.Description != "" {
		endpoint.Description = input.Description
	}

	if err := s.endpointRepo.Update(ctx, endpoint); err != nil {
		return nil, err
	}

	return endpoint, nil
}

// Delete deletes an endpoint
func (s *EndpointService) Delete(ctx context.Context, userID, endpointID uuid.UUID) error {
	endpoint, err := s.endpointRepo.GetByID(ctx, endpointID)
	if err != nil {
		return err
	}

	if endpoint.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	// Remove from cache
	s.rdb.Del(ctx, "endpoint:token:"+endpoint.Token)

	return s.endpointRepo.Delete(ctx, endpointID)
}

// SetCustomResponse sets the custom response for an endpoint
func (s *EndpointService) SetCustomResponse(ctx context.Context, userID, endpointID uuid.UUID, response *models.CustomResponse) error {
	endpoint, err := s.endpointRepo.GetByID(ctx, endpointID)
	if err != nil {
		return err
	}

	if endpoint.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	endpoint.CustomResponse = response
	return s.endpointRepo.Update(ctx, endpoint)
}

// generateToken generates a random URL-safe token
func generateToken() string {
	b := make([]byte, 6)
	rand.Read(b)
	return hex.EncodeToString(b)
}
