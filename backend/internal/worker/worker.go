package worker

import (
	"context"
	"time"

	"github.com/inst-dev/webhook/internal/config"
	"github.com/inst-dev/webhook/internal/database"
	"github.com/inst-dev/webhook/internal/redis"
	"github.com/rs/zerolog/log"
)

// Worker handles background cleanup and maintenance tasks
type Worker struct {
	cfg  *config.Config
	db   *database.Pool
	rdb  *redis.Client
	done chan struct{}
}

// New creates a new worker
func New(cfg *config.Config, db *database.Pool, rdb *redis.Client) *Worker {
	return &Worker{
		cfg:  cfg,
		db:   db,
		rdb:  rdb,
		done: make(chan struct{}),
	}
}

// Start starts all worker routines
func (w *Worker) Start(ctx context.Context) {
	log.Info().Msg("Worker started - running cleanup tasks")

	// Run cleanup every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Run immediately on start
	w.runCleanup(ctx)

	for {
		select {
		case <-ticker.C:
			w.runCleanup(ctx)
		case <-w.done:
			return
		case <-ctx.Done():
			return
		}
	}
}

// Stop stops the worker
func (w *Worker) Stop() {
	close(w.done)
}

// runCleanup runs all cleanup tasks
func (w *Worker) runCleanup(ctx context.Context) {
	log.Debug().Msg("Running cleanup tasks")

	w.cleanExpiredRequests(ctx)
	w.cleanExpiredSessions(ctx)
	w.cleanExpiredDNSLogs(ctx)
	w.cleanExpiredEmails(ctx)
	w.aggregateMetrics(ctx)
}

// cleanExpiredRequests removes requests past their retention period
func (w *Worker) cleanExpiredRequests(ctx context.Context) {
	// Free tier: 72 hours
	freeExpiry := time.Now().Add(-time.Duration(w.cfg.RetentionFree) * time.Hour)

	result, err := w.db.Exec(ctx, `
		DELETE FROM requests 
		WHERE created_at < $1 
		AND endpoint_id IN (
			SELECT e.id FROM endpoints e 
			JOIN users u ON e.user_id = u.id 
			WHERE u.plan = 'free'
		)
	`, freeExpiry)

	if err != nil {
		log.Error().Err(err).Msg("Failed to clean expired requests (free)")
		return
	}

	if result.RowsAffected() > 0 {
		log.Info().Int64("deleted", result.RowsAffected()).Msg("Cleaned expired free requests")
	}

	// Pro tier: 720 hours
	proExpiry := time.Now().Add(-time.Duration(w.cfg.RetentionPro) * time.Hour)

	result, err = w.db.Exec(ctx, `
		DELETE FROM requests 
		WHERE created_at < $1 
		AND endpoint_id IN (
			SELECT e.id FROM endpoints e 
			JOIN users u ON e.user_id = u.id 
			WHERE u.plan = 'pro'
		)
	`, proExpiry)

	if err != nil {
		log.Error().Err(err).Msg("Failed to clean expired requests (pro)")
		return
	}

	if result.RowsAffected() > 0 {
		log.Info().Int64("deleted", result.RowsAffected()).Msg("Cleaned expired pro requests")
	}
}

// cleanExpiredSessions removes expired sessions
func (w *Worker) cleanExpiredSessions(ctx context.Context) {
	result, err := w.db.Exec(ctx, `
		DELETE FROM sessions WHERE expires_at < $1
	`, time.Now())

	if err != nil {
		log.Error().Err(err).Msg("Failed to clean expired sessions")
		return
	}

	if result.RowsAffected() > 0 {
		log.Info().Int64("deleted", result.RowsAffected()).Msg("Cleaned expired sessions")
	}
}

// cleanExpiredDNSLogs removes old DNS logs
func (w *Worker) cleanExpiredDNSLogs(ctx context.Context) {
	expiry := time.Now().Add(-72 * time.Hour) // 3 days for free

	result, err := w.db.Exec(ctx, `
		DELETE FROM dns_logs WHERE created_at < $1
		AND endpoint_id IN (
			SELECT e.id FROM endpoints e 
			JOIN users u ON e.user_id = u.id 
			WHERE u.plan = 'free'
		)
	`, expiry)

	if err != nil {
		log.Error().Err(err).Msg("Failed to clean expired DNS logs")
		return
	}

	if result.RowsAffected() > 0 {
		log.Info().Int64("deleted", result.RowsAffected()).Msg("Cleaned expired DNS logs")
	}
}

// cleanExpiredEmails removes old email logs
func (w *Worker) cleanExpiredEmails(ctx context.Context) {
	expiry := time.Now().Add(-72 * time.Hour)

	result, err := w.db.Exec(ctx, `
		DELETE FROM email_logs WHERE created_at < $1
		AND endpoint_id IN (
			SELECT e.id FROM endpoints e 
			JOIN users u ON e.user_id = u.id 
			WHERE u.plan = 'free'
		)
	`, expiry)

	if err != nil {
		log.Error().Err(err).Msg("Failed to clean expired email logs")
		return
	}

	if result.RowsAffected() > 0 {
		log.Info().Int64("deleted", result.RowsAffected()).Msg("Cleaned expired email logs")
	}
}

// aggregateMetrics aggregates realtime metrics into persistent storage
func (w *Worker) aggregateMetrics(ctx context.Context) {
	// Aggregate hourly stats from Redis into PostgreSQL
	now := time.Now()
	hourKey := "stats:req:global:hour:" + now.Add(-1*time.Hour).Format("2006010215")

	count, err := w.rdb.Get(ctx, hourKey).Int64()
	if err != nil {
		return // No data for this hour
	}

	_, err = w.db.Exec(ctx, `
		INSERT INTO metrics_hourly (hour, request_count, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (hour) DO UPDATE SET request_count = $2
	`, now.Add(-1*time.Hour).Truncate(time.Hour), count, now)

	if err != nil {
		log.Error().Err(err).Msg("Failed to aggregate hourly metrics")
	}
}
