package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/inst-dev/webhook/internal/config"
	"github.com/inst-dev/webhook/internal/database"
	"github.com/inst-dev/webhook/internal/redis"
	"github.com/inst-dev/webhook/internal/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Set log level
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err == nil {
		zerolog.SetGlobalLevel(level)
	}

	if cfg.LogFormat == "json" {
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	}

	log.Info().Str("env", cfg.AppEnv).Msg("Starting Webhook.inst.lk API server")

	// Initialize database
	db, err := database.NewPool(context.Background(), cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	log.Info().Msg("Database connection established")

	// Initialize Redis
	rdb, err := redis.NewClient(cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer rdb.Close()

	log.Info().Msg("Redis connection established")

	// Create and start server
	srv := server.New(cfg, db, rdb)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	log.Info().Int("port", cfg.APIPort).Msg("API server started")

	<-quit
	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited properly")
}
