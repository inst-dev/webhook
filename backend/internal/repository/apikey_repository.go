package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/inst-dev/webhook/internal/database"
	"github.com/inst-dev/webhook/internal/models"
)

// APIKeyRepository handles API key data access
type APIKeyRepository struct {
	db *database.Pool
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *database.Pool) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// Create creates a new API key
func (r *APIKeyRepository) Create(ctx context.Context, key *models.APIKey) error {
	key.ID = uuid.New()
	key.CreatedAt = time.Now()

	_, err := r.db.Exec(ctx, `
		INSERT INTO api_keys (id, user_id, name, key_hash, key_prefix, scopes, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, key.ID, key.UserID, key.Name, key.KeyHash, key.KeyPrefix, key.Scopes, key.ExpiresAt, key.CreatedAt)

	return err
}

// GetByHash gets an API key by hash
func (r *APIKeyRepository) GetByHash(ctx context.Context, keyHash string) (*models.APIKey, error) {
	key := &models.APIKey{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, name, key_hash, key_prefix, scopes, last_used_at, expires_at, created_at, revoked_at
		FROM api_keys WHERE key_hash = $1 AND revoked_at IS NULL
	`, keyHash).Scan(
		&key.ID, &key.UserID, &key.Name, &key.KeyHash, &key.KeyPrefix,
		&key.Scopes, &key.LastUsedAt, &key.ExpiresAt, &key.CreatedAt, &key.RevokedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return key, nil
}

// ListByUser lists API keys for a user
func (r *APIKeyRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.APIKey, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, name, key_prefix, scopes, last_used_at, expires_at, created_at, revoked_at
		FROM api_keys WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*models.APIKey
	for rows.Next() {
		key := &models.APIKey{}
		err := rows.Scan(
			&key.ID, &key.UserID, &key.Name, &key.KeyPrefix,
			&key.Scopes, &key.LastUsedAt, &key.ExpiresAt, &key.CreatedAt, &key.RevokedAt,
		)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	return keys, nil
}

// Revoke revokes an API key
func (r *APIKeyRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := r.db.Exec(ctx, `UPDATE api_keys SET revoked_at = $2 WHERE id = $1`, id, now)
	return err
}

// UpdateLastUsed updates the last_used_at timestamp
func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := r.db.Exec(ctx, `UPDATE api_keys SET last_used_at = $2 WHERE id = $1`, id, now)
	return err
}
