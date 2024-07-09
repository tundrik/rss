package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"rss/configs"
	"rss/internal/repository"
	"rss/internal/restapi"
	"rss/logger"
)




func main() {
	ctx, cancel := context.WithCancel(context.Background())
	log := logger.CreateLogger()

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("fail read config")
	}
	repo, err := repository.New(ctx, cfg, log) 
	if err != nil {
		log.Fatal().Err(err).Msg("fail new repository")
	}

	rest := restapi.New(repo, log)
	rest.Run()

	log.Info().Msg("starting app")
  
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	sign := <-signals
	log.Info().Str("signal", sign.String()).Msg("shutdown stoping app")
	rest.Shutdown()
	log.Info().Msg("stop app")
	cancel()
}