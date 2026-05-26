-- Webhook.inst.lk Database Schema
-- Migration 001: Initial schema

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    plan VARCHAR(20) NOT NULL DEFAULT 'free',
    email_verified BOOLEAN NOT NULL DEFAULT false,
    two_factor_enabled BOOLEAN NOT NULL DEFAULT false,
    two_factor_secret VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_plan ON users(plan);

-- Endpoints table
CREATE TABLE endpoints (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(100),
    description VARCHAR(500),
    is_active BOOLEAN NOT NULL DEFAULT true,
    custom_response JSONB,
    request_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_endpoints_user_id ON endpoints(user_id);
CREATE INDEX idx_endpoints_token ON endpoints(token);
CREATE INDEX idx_endpoints_active ON endpoints(is_active) WHERE is_active = true;

-- Requests table (partitioned by month for performance)
CREATE TABLE requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    endpoint_id UUID NOT NULL REFERENCES endpoints(id) ON DELETE CASCADE,
    method VARCHAR(10) NOT NULL,
    path TEXT NOT NULL DEFAULT '/',
    headers JSONB,
    query_params JSONB,
    body BYTEA,
    content_type VARCHAR(255),
    content_length BIGINT DEFAULT 0,
    source_ip VARCHAR(45) NOT NULL,
    user_agent TEXT,
    country VARCHAR(2),
    asn VARCHAR(50),
    response_code INTEGER NOT NULL DEFAULT 200,
    response_time BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_requests_endpoint_id ON requests(endpoint_id);
CREATE INDEX idx_requests_created_at ON requests(created_at DESC);
CREATE INDEX idx_requests_method ON requests(method);
CREATE INDEX idx_requests_source_ip ON requests(source_ip);
CREATE INDEX idx_requests_endpoint_created ON requests(endpoint_id, created_at DESC);

-- DNS Logs table
CREATE TABLE dns_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    endpoint_id UUID REFERENCES endpoints(id) ON DELETE SET NULL,
    query_name VARCHAR(255) NOT NULL,
    query_type VARCHAR(10) NOT NULL,
    source_ip VARCHAR(45) NOT NULL,
    source_port INTEGER,
    raw_query BYTEA,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_dns_logs_endpoint_id ON dns_logs(endpoint_id);
CREATE INDEX idx_dns_logs_created_at ON dns_logs(created_at DESC);
CREATE INDEX idx_dns_logs_source_ip ON dns_logs(source_ip);

-- Email Logs table
CREATE TABLE email_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    endpoint_id UUID REFERENCES endpoints(id) ON DELETE SET NULL,
    from_addr VARCHAR(255) NOT NULL,
    to_addr VARCHAR(255) NOT NULL,
    subject TEXT,
    body TEXT,
    html_body TEXT,
    raw_message BYTEA,
    headers JSONB,
    attachments JSONB,
    source_ip VARCHAR(45),
    size BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_logs_endpoint_id ON email_logs(endpoint_id);
CREATE INDEX idx_email_logs_created_at ON email_logs(created_at DESC);
CREATE INDEX idx_email_logs_to_addr ON email_logs(to_addr);

-- Sessions table
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(255) NOT NULL,
    user_agent TEXT,
    ip VARCHAR(45),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- API Keys table
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    key_hash VARCHAR(64) NOT NULL UNIQUE,
    key_prefix VARCHAR(12) NOT NULL,
    scopes TEXT[] NOT NULL DEFAULT '{}',
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);

-- Subscriptions table
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    provider VARCHAR(20) NOT NULL,
    provider_sub_id VARCHAR(255),
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    cancelled_at TIMESTAMPTZ
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);

-- Audit Logs table
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    details JSONB,
    ip VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- Metrics hourly aggregation table
CREATE TABLE metrics_hourly (
    id SERIAL PRIMARY KEY,
    hour TIMESTAMPTZ NOT NULL UNIQUE,
    request_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_metrics_hourly_hour ON metrics_hourly(hour DESC);

-- Custom domains table
CREATE TABLE custom_domains (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    domain VARCHAR(255) NOT NULL UNIQUE,
    verified BOOLEAN NOT NULL DEFAULT false,
    verification_token VARCHAR(64),
    ssl_certificate TEXT,
    ssl_key TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMPTZ
);

CREATE INDEX idx_custom_domains_user_id ON custom_domains(user_id);
CREATE INDEX idx_custom_domains_domain ON custom_domains(domain);
