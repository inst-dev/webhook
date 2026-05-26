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

// RequestRepository handles request data access
type RequestRepository struct {
	db *database.Pool
}

// NewRequestRepository creates a new request repository
func NewRequestRepository(db *database.Pool) *RequestRepository {
	return &RequestRepository{db: db}
}

// Create stores a new captured request
func (r *RequestRepository) Create(ctx context.Context, req *models.Request) error {
	req.ID = uuid.New()
	req.CreatedAt = time.Now()

	_, err := r.db.Exec(ctx, `
		INSERT INTO requests (id, endpoint_id, method, path, headers, query_params, body, 
			content_type, content_length, source_ip, user_agent, country, asn,
			response_code, response_time, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, req.ID, req.EndpointID, req.Method, req.Path, req.Headers, req.QueryParams,
		req.Body, req.ContentType, req.ContentLength, req.SourceIP, req.UserAgent,
		req.Country, req.ASN, req.ResponseCode, req.ResponseTime, req.CreatedAt)

	return err
}

// GetByID gets a request by ID
func (r *RequestRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Request, error) {
	req := &models.Request{}
	err := r.db.QueryRow(ctx, `
		SELECT id, endpoint_id, method, path, headers, query_params, body,
			   content_type, content_length, source_ip, user_agent, country, asn,
			   response_code, response_time, created_at
		FROM requests WHERE id = $1
	`, id).Scan(
		&req.ID, &req.EndpointID, &req.Method, &req.Path, &req.Headers,
		&req.QueryParams, &req.Body, &req.ContentType, &req.ContentLength,
		&req.SourceIP, &req.UserAgent, &req.Country, &req.ASN,
		&req.ResponseCode, &req.ResponseTime, &req.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return req, nil
}

// ListByEndpoint lists requests for an endpoint with pagination and filtering
func (r *RequestRepository) ListByEndpoint(ctx context.Context, endpointID uuid.UUID, opts ListOptions) ([]*models.Request, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM requests WHERE endpoint_id = $1`
	err := r.db.QueryRow(ctx, countQuery, endpointID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, endpoint_id, method, path, headers, query_params, body,
			   content_type, content_length, source_ip, user_agent, country, asn,
			   response_code, response_time, created_at
		FROM requests WHERE endpoint_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, endpointID, opts.Limit, opts.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var requests []*models.Request
	for rows.Next() {
		req := &models.Request{}
		err := rows.Scan(
			&req.ID, &req.EndpointID, &req.Method, &req.Path, &req.Headers,
			&req.QueryParams, &req.Body, &req.ContentType, &req.ContentLength,
			&req.SourceIP, &req.UserAgent, &req.Country, &req.ASN,
			&req.ResponseCode, &req.ResponseTime, &req.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		requests = append(requests, req)
	}

	return requests, total, nil
}

// Search searches requests by various criteria
func (r *RequestRepository) Search(ctx context.Context, endpointID uuid.UUID, query string, opts ListOptions) ([]*models.Request, int64, error) {
	searchQuery := `
		SELECT id, endpoint_id, method, path, headers, query_params, body,
			   content_type, content_length, source_ip, user_agent, country, asn,
			   response_code, response_time, created_at
		FROM requests 
		WHERE endpoint_id = $1 
		  AND (
			  source_ip ILIKE $4 OR
			  user_agent ILIKE $4 OR
			  method ILIKE $4 OR
			  headers::text ILIKE $4 OR
			  body::text ILIKE $4
		  )
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	searchPattern := "%" + query + "%"

	var total int64
	countQuery := `
		SELECT COUNT(*) FROM requests 
		WHERE endpoint_id = $1 
		  AND (source_ip ILIKE $2 OR user_agent ILIKE $2 OR method ILIKE $2 OR headers::text ILIKE $2 OR body::text ILIKE $2)
	`
	err := r.db.QueryRow(ctx, countQuery, endpointID, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, searchQuery, endpointID, opts.Limit, opts.Offset, searchPattern)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var requests []*models.Request
	for rows.Next() {
		req := &models.Request{}
		err := rows.Scan(
			&req.ID, &req.EndpointID, &req.Method, &req.Path, &req.Headers,
			&req.QueryParams, &req.Body, &req.ContentType, &req.ContentLength,
			&req.SourceIP, &req.UserAgent, &req.Country, &req.ASN,
			&req.ResponseCode, &req.ResponseTime, &req.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		requests = append(requests, req)
	}

	return requests, total, nil
}

// Delete deletes a request by ID
func (r *RequestRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM requests WHERE id = $1`, id)
	return err
}

// DeleteExpired deletes requests older than the retention period
func (r *RequestRepository) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.db.Exec(ctx, `DELETE FROM requests WHERE created_at < $1`, before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// ListOptions holds pagination options
type ListOptions struct {
	Limit  int
	Offset int
	Sort   string
	Order  string
}
