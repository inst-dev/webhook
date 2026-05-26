package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/inst-dev/webhook/internal/config"
	"github.com/inst-dev/webhook/internal/database"
	"github.com/inst-dev/webhook/internal/redis"
	"github.com/inst-dev/webhook/internal/smtp"
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

	log.Info().Msg("Starting Webhook.inst.lk SMTP Service")

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

	smtpServer := smtp.NewServer(cfg, db, rdb)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := smtpServer.Start(context.Background()); err != nil {
			log.Fatal().Err(err).Msg("SMTP server failed to start")
		}
	}()

	log.Info().Int("port", cfg.SMTPPort).Msg("SMTP server started")

	<-quit
	log.Info().Msg("Shutting down SMTP server...")
	smtpServer.Stop()
}
