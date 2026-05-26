package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/inst-dev/webhook/internal/models"
	"github.com/inst-dev/webhook/internal/redis"
	"github.com/inst-dev/webhook/internal/repository"
)

// APIKeyService handles API key business logic
type APIKeyService struct {
	repo *repository.APIKeyRepository
	rdb  *redis.Client
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService(repo *repository.APIKeyRepository, rdb *redis.Client) *APIKeyService {
	return &APIKeyService{repo: repo, rdb: rdb}
}

// CreateAPIKeyInput holds API key creation data
type CreateAPIKeyInput struct {
	Name      string   `json:"name" validate:"required,min=1,max=100"`
	Scopes    []string `json:"scopes" validate:"required,min=1"`
	ExpiresIn *int     `json:"expires_in"` // days, nil = never
}

// CreateAPIKeyResult holds the result of creating an API key
type CreateAPIKeyResult struct {
	Key     string         `json:"key"` // Only shown once
	APIKey  *models.APIKey `json:"api_key"`
}

// Create generates a new API key
func (s *APIKeyService) Create(ctx context.Context, userID uuid.UUID, input CreateAPIKeyInput) (*CreateAPIKeyResult, error) {
	// Generate the raw API key
	rawKey := generateAPIKey()
	keyHash := hashAPIKey(rawKey)
	keyPrefix := rawKey[:8]

	// Calculate expiry
	var expiresAt *time.Time
	if input.ExpiresIn != nil {
		t := time.Now().AddDate(0, 0, *input.ExpiresIn)
		expiresAt = &t
	}

	// Validate scopes
	validScopes := map[string]bool{
		"*":                true,
		"endpoints:read":   true,
		"endpoints:write":  true,
		"requests:read":    true,
		"requests:write":   true,
		"dns:read":         true,
		"email:read":       true,
	}

	for _, scope := range input.Scopes {
		if !validScopes[scope] {
			return nil, fmt.Errorf("invalid scope: %s", scope)
		}
	}

	apiKey := &models.APIKey{
		UserID:    userID,
		Name:      input.Name,
		KeyHash:   keyHash,
		KeyPrefix: keyPrefix,
		Scopes:    input.Scopes,
		ExpiresAt: expiresAt,
	}

	if err := s.repo.Create(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return &CreateAPIKeyResult{
		Key:    rawKey,
		APIKey: apiKey,
	}, nil
}

// Validate validates an API key and returns the key model
func (s *APIKeyService) Validate(ctx context.Context, rawKey string) (*models.APIKey, error) {
	// Check cache first
	cached, err := s.rdb.Get(ctx, "apikey:"+rawKey[:8]).Result()
	if err == nil && cached == "revoked" {
		return nil, fmt.Errorf("API key revoked")
	}

	keyHash := hashAPIKey(rawKey)
	key, err := s.repo.GetByHash(ctx, keyHash)
	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}

	// Check expiry
	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API key expired")
	}

	// Update last used (async)
	go s.repo.UpdateLastUsed(context.Background(), key.ID)

	return key, nil
}

// List lists API keys for a user
func (s *APIKeyService) List(ctx context.Context, userID uuid.UUID) ([]*models.APIKey, error) {
	return s.repo.ListByUser(ctx, userID)
}

// Revoke revokes an API key
func (s *APIKeyService) Revoke(ctx context.Context, userID uuid.UUID, keyID uuid.UUID) error {
	// Mark as revoked in cache
	s.rdb.Set(ctx, "apikey:revoked:"+keyID.String(), "1", 24*time.Hour)
	return s.repo.Revoke(ctx, keyID)
}

// generateAPIKey generates a random API key
func generateAPIKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return "whk_" + hex.EncodeToString(b)
}

// hashAPIKey creates a SHA-256 hash of the API key
func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
