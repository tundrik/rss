package restapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"rss/configs"
	"rss/internal/usecase"

	"github.com/rs/zerolog"
)


type RestApi struct {
	uc  *usecase.UseCase
	srv *http.Server
	log zerolog.Logger
}

func New(uc *usecase.UseCase, cfg config.HttpConfig, log zerolog.Logger) *RestApi {
	e := &RestApi{
		uc:  uc,
		log: log,
	}
	e.srv = &http.Server{
		Addr:         cfg.Port,
		Handler:      e.registerRoutes(),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
	return e
}

func (e *RestApi) Run() {
	go func() {
		if err := e.srv.ListenAndServe(); err != nil {
			// исключаем shutdown
			if !errors.Is(err, http.ErrServerClosed) {
				e.log.Fatal().Err(err).Msg("fail listen and serve")
			}
		}
	}()
}

func (e *RestApi) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := e.srv.Shutdown(ctx); err != nil {
		e.log.Fatal().Err(err).Msg("fail shutdown")
		e.srv.Close()
	}
}
