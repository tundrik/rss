package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"rss/configs"
	"rss/internal/crawly"
	"rss/internal/repository"
	"rss/logger"
)

func main() {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	log := logger.CreateLogger()

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("fail read config")
	}

	repo, err := repository.New(ctx, cfg, log)
	if err != nil {
		log.Fatal().Err(err).Msg("fail new repository")
	}

	crawl := crawly.New(repo, cfg, log)
	crawl.Run()
	log.Info().Msg("starting crawly")

	signals := make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	sign := <-signals
	log.Info().Str("signal", sign.String()).Msg("stoping crawly")
}
