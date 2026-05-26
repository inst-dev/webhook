package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/inst-dev/webhook/internal/config"
	"github.com/inst-dev/webhook/internal/database"
	"github.com/inst-dev/webhook/internal/redis"
	"github.com/inst-dev/webhook/internal/worker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	if cfg.LogFormat == "json" {
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	}

	log.Info().Msg("Starting Webhook.inst.lk Worker")

	db, err := database.NewPool(context.Background(), cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	rdb, err := redis.NewClient(cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer rdb.Close()

	w := worker.New(cfg, db, rdb)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go w.Start(context.Background())

	log.Info().Msg("Worker started")

	<-quit
	log.Info().Msg("Shutting down worker...")
	w.Stop()
	log.Info().Msg("Worker stopped")
}
