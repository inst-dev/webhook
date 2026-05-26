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

var (
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
)

// UserRepository handles user data access
type UserRepository struct {
	db *database.Pool
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *database.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, display_name, plan, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, user.ID, user.Email, user.PasswordHash, user.DisplayName, user.Plan, user.EmailVerified, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		if isDuplicateError(err) {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

// GetByID gets a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, display_name, plan, email_verified, 
			   two_factor_enabled, two_factor_secret, created_at, updated_at, last_login_at
		FROM users WHERE id = $1
	`, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.Plan,
		&user.EmailVerified, &user.TwoFactorEnabled, &user.TwoFactorSecret,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

// GetByEmail gets a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, display_name, plan, email_verified,
			   two_factor_enabled, two_factor_secret, created_at, updated_at, last_login_at
		FROM users WHERE email = $1
	`, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.Plan,
		&user.EmailVerified, &user.TwoFactorEnabled, &user.TwoFactorSecret,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx, `
		UPDATE users SET display_name = $2, plan = $3, email_verified = $4,
			   two_factor_enabled = $5, two_factor_secret = $6, updated_at = $7, last_login_at = $8
		WHERE id = $1
	`, user.ID, user.DisplayName, user.Plan, user.EmailVerified,
		user.TwoFactorEnabled, user.TwoFactorSecret, user.UpdatedAt, user.LastLoginAt)
	return err
}

// UpdatePassword updates user password
func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET password_hash = $2, updated_at = $3 WHERE id = $1
	`, id, passwordHash, time.Now())
	return err
}

// Count returns total user count
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

// isDuplicateError checks for PostgreSQL unique violation
func isDuplicateError(err error) bool {
	return err != nil && (err.Error() == "ERROR: duplicate key value violates unique constraint" ||
		contains(err.Error(), "duplicate key"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
