package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/inst-dev/webhook/internal/database"
	"github.com/inst-dev/webhook/internal/models"
)

// EndpointRepository handles endpoint data access
type EndpointRepository struct {
	db *database.Pool
}

// NewEndpointRepository creates a new endpoint repository
func NewEndpointRepository(db *database.Pool) *EndpointRepository {
	return &EndpointRepository{db: db}
}

// Create creates a new endpoint
func (r *EndpointRepository) Create(ctx context.Context, endpoint *models.Endpoint) error {
	endpoint.ID = uuid.New()
	endpoint.CreatedAt = time.Now()
	endpoint.UpdatedAt = time.Now()

	var customResponseJSON []byte
	if endpoint.CustomResponse != nil {
		var err error
		customResponseJSON, err = json.Marshal(endpoint.CustomResponse)
		if err != nil {
			return err
		}
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO endpoints (id, user_id, token, name, description, is_active, custom_response, created_at, updated_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, endpoint.ID, endpoint.UserID, endpoint.Token, endpoint.Name, endpoint.Description,
		endpoint.IsActive, customResponseJSON, endpoint.CreatedAt, endpoint.UpdatedAt, endpoint.ExpiresAt)

	return err
}

// GetByID gets an endpoint by ID
func (r *EndpointRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Endpoint, error) {
	endpoint := &models.Endpoint{}
	var customResponseJSON []byte

	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, token, name, description, is_active, custom_response, 
			   request_count, created_at, updated_at, expires_at
		FROM endpoints WHERE id = $1
	`, id).Scan(
		&endpoint.ID, &endpoint.UserID, &endpoint.Token, &endpoint.Name,
		&endpoint.Description, &endpoint.IsActive, &customResponseJSON,
		&endpoint.RequestCount, &endpoint.CreatedAt, &endpoint.UpdatedAt, &endpoint.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if customResponseJSON != nil {
		endpoint.CustomResponse = &models.CustomResponse{}
		json.Unmarshal(customResponseJSON, endpoint.CustomResponse)
	}

	return endpoint, nil
}

// GetByToken gets an endpoint by token
func (r *EndpointRepository) GetByToken(ctx context.Context, token string) (*models.Endpoint, error) {
	endpoint := &models.Endpoint{}
	var customResponseJSON []byte

	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, token, name, description, is_active, custom_response,
			   request_count, created_at, updated_at, expires_at
		FROM endpoints WHERE token = $1 AND is_active = true
	`, token).Scan(
		&endpoint.ID, &endpoint.UserID, &endpoint.Token, &endpoint.Name,
		&endpoint.Description, &endpoint.IsActive, &customResponseJSON,
		&endpoint.RequestCount, &endpoint.CreatedAt, &endpoint.UpdatedAt, &endpoint.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if customResponseJSON != nil {
		endpoint.CustomResponse = &models.CustomResponse{}
		json.Unmarshal(customResponseJSON, endpoint.CustomResponse)
	}

	return endpoint, nil
}

// ListByUser lists endpoints for a user with pagination
func (r *EndpointRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Endpoint, int64, error) {
	var total int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM endpoints WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, token, name, description, is_active, custom_response,
			   request_count, created_at, updated_at, expires_at
		FROM endpoints WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var endpoints []*models.Endpoint
	for rows.Next() {
		endpoint := &models.Endpoint{}
		var customResponseJSON []byte

		err := rows.Scan(
			&endpoint.ID, &endpoint.UserID, &endpoint.Token, &endpoint.Name,
			&endpoint.Description, &endpoint.IsActive, &customResponseJSON,
			&endpoint.RequestCount, &endpoint.CreatedAt, &endpoint.UpdatedAt, &endpoint.ExpiresAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if customResponseJSON != nil {
			endpoint.CustomResponse = &models.CustomResponse{}
			json.Unmarshal(customResponseJSON, endpoint.CustomResponse)
		}

		endpoints = append(endpoints, endpoint)
	}

	return endpoints, total, nil
}

// Update updates an endpoint
func (r *EndpointRepository) Update(ctx context.Context, endpoint *models.Endpoint) error {
	endpoint.UpdatedAt = time.Now()

	var customResponseJSON []byte
	if endpoint.CustomResponse != nil {
		var err error
		customResponseJSON, err = json.Marshal(endpoint.CustomResponse)
		if err != nil {
			return err
		}
	}

	_, err := r.db.Exec(ctx, `
		UPDATE endpoints SET name = $2, description = $3, is_active = $4, 
			   custom_response = $5, updated_at = $6, expires_at = $7
		WHERE id = $1
	`, endpoint.ID, endpoint.Name, endpoint.Description, endpoint.IsActive,
		customResponseJSON, endpoint.UpdatedAt, endpoint.ExpiresAt)
	return err
}

// Delete deletes an endpoint
func (r *EndpointRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM endpoints WHERE id = $1`, id)
	return err
}

// IncrementRequestCount increments the request count
func (r *EndpointRepository) IncrementRequestCount(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE endpoints SET request_count = request_count + 1 WHERE id = $1
	`, id)
	return err
}
