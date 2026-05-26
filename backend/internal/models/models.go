package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// User represents a platform user
type User struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	Email           string     `json:"email" db:"email"`
	PasswordHash    string     `json:"-" db:"password_hash"`
	DisplayName     string     `json:"display_name" db:"display_name"`
	Plan            string     `json:"plan" db:"plan"`
	EmailVerified   bool       `json:"email_verified" db:"email_verified"`
	TwoFactorEnabled bool     `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorSecret *string   `json:"-" db:"two_factor_secret"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt     *time.Time `json:"last_login_at" db:"last_login_at"`
}

// Endpoint represents a webhook endpoint
type Endpoint struct {
	ID               uuid.UUID        `json:"id" db:"id"`
	UserID           uuid.UUID        `json:"user_id" db:"user_id"`
	Token            string           `json:"token" db:"token"`
	Name             string           `json:"name" db:"name"`
	Description      string           `json:"description" db:"description"`
	IsActive         bool             `json:"is_active" db:"is_active"`
	CustomResponse   *CustomResponse  `json:"custom_response" db:"custom_response"`
	RequestCount     int64            `json:"request_count" db:"request_count"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at" db:"updated_at"`
	ExpiresAt        *time.Time       `json:"expires_at" db:"expires_at"`
}

// CustomResponse defines the custom response configuration for an endpoint
type CustomResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Delay      int               `json:"delay"` // milliseconds
	Redirect   string            `json:"redirect"`
}

// Request represents a captured webhook request
type Request struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	EndpointID    uuid.UUID              `json:"endpoint_id" db:"endpoint_id"`
	Method        string                 `json:"method" db:"method"`
	Path          string                 `json:"path" db:"path"`
	Headers       json.RawMessage        `json:"headers" db:"headers"`
	QueryParams   json.RawMessage        `json:"query_params" db:"query_params"`
	Body          []byte                 `json:"body" db:"body"`
	ContentType   string                 `json:"content_type" db:"content_type"`
	ContentLength int64                  `json:"content_length" db:"content_length"`
	SourceIP      string                 `json:"source_ip" db:"source_ip"`
	UserAgent     string                 `json:"user_agent" db:"user_agent"`
	Country       string                 `json:"country" db:"country"`
	ASN           string                 `json:"asn" db:"asn"`
	ResponseCode  int                    `json:"response_code" db:"response_code"`
	ResponseTime  int64                  `json:"response_time" db:"response_time"` // microseconds
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
}

// DNSLog represents a DNS query log
type DNSLog struct {
	ID          uuid.UUID `json:"id" db:"id"`
	EndpointID  uuid.UUID `json:"endpoint_id" db:"endpoint_id"`
	QueryName   string    `json:"query_name" db:"query_name"`
	QueryType   string    `json:"query_type" db:"query_type"`
	SourceIP    string    `json:"source_ip" db:"source_ip"`
	SourcePort  int       `json:"source_port" db:"source_port"`
	RawQuery    []byte    `json:"raw_query" db:"raw_query"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// EmailLog represents a captured email
type EmailLog struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	EndpointID  uuid.UUID       `json:"endpoint_id" db:"endpoint_id"`
	From        string          `json:"from" db:"from_addr"`
	To          string          `json:"to" db:"to_addr"`
	Subject     string          `json:"subject" db:"subject"`
	Body        string          `json:"body" db:"body"`
	HTMLBody    string          `json:"html_body" db:"html_body"`
	RawMessage  []byte          `json:"-" db:"raw_message"`
	Headers     json.RawMessage `json:"headers" db:"headers"`
	Attachments json.RawMessage `json:"attachments" db:"attachments"`
	SourceIP    string          `json:"source_ip" db:"source_ip"`
	Size        int64           `json:"size" db:"size"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// Session represents a user session
type Session struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	RefreshToken string     `json:"-" db:"refresh_token"`
	UserAgent    string     `json:"user_agent" db:"user_agent"`
	IP           string     `json:"ip" db:"ip"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	RevokedAt    *time.Time `json:"revoked_at" db:"revoked_at"`
}

// APIKey represents an API key
type APIKey struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	Name        string     `json:"name" db:"name"`
	KeyHash     string     `json:"-" db:"key_hash"`
	KeyPrefix   string     `json:"key_prefix" db:"key_prefix"`
	Scopes      []string   `json:"scopes" db:"scopes"`
	LastUsedAt  *time.Time `json:"last_used_at" db:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	RevokedAt   *time.Time `json:"revoked_at" db:"revoked_at"`
}

// Subscription represents a billing subscription
type Subscription struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	UserID          uuid.UUID  `json:"user_id" db:"user_id"`
	Plan            string     `json:"plan" db:"plan"`
	Status          string     `json:"status" db:"status"`
	Provider        string     `json:"provider" db:"provider"`
	ProviderSubID   string     `json:"provider_sub_id" db:"provider_sub_id"`
	CurrentPeriodStart time.Time `json:"current_period_start" db:"current_period_start"`
	CurrentPeriodEnd   time.Time `json:"current_period_end" db:"current_period_end"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	CancelledAt     *time.Time `json:"cancelled_at" db:"cancelled_at"`
}

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	UserID    *uuid.UUID      `json:"user_id" db:"user_id"`
	Action    string          `json:"action" db:"action"`
	Resource  string          `json:"resource" db:"resource"`
	Details   json.RawMessage `json:"details" db:"details"`
	IP        string          `json:"ip" db:"ip"`
	UserAgent string          `json:"user_agent" db:"user_agent"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
}
