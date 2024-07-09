package restapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"rss/internal/repository"

	"github.com/rs/zerolog"
)

var (
	msgCreated       = []byte(`{"message": "201 Created"}`)
	msgNoContent     = []byte(`{"message": "204 No Content"}`)
	msgBadRequest    = []byte(`{"message": "400 Bad Request"}`)
	msgUnauthorized  = []byte(`{"message": "401 Unauthorized"}`)
	msgForbidden     = []byte(`{"message": "403 Forbidden"}`)
	msgConflict      = []byte(`{"message": "409 Conflict"}`)
	msgInternalError = []byte(`{"message": "500 Internal Server Error"}`)
)

type RestApi struct {
	repo *repository.Repo
	srv  *http.Server
	log  zerolog.Logger
}

func New(repo *repository.Repo, log zerolog.Logger) *RestApi {
	e := &RestApi{
		repo: repo,
		log:  log,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /add", e.authUser(e.authAdmin(e.addFeed)))
	mux.HandleFunc("GET /{$}", e.authUser(e.available))
	mux.HandleFunc("PUT /subscribe", e.authUser(e.subscribe))
	mux.HandleFunc("PUT /unsubscribe", e.authUser(e.unsubscribe))
	mux.HandleFunc("GET /article", e.authUser(e.article))

	e.srv = &http.Server{
		Addr:         ":8000",
		Handler:      e.middleware(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
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
		e.srv.Close()
		e.log.Fatal().Err(err).Msg("fail shutdown")
	}
}

func forbidden(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden)
	w.Write(msgForbidden)
}

func unauthorized(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write(msgUnauthorized)
}

func created(w http.ResponseWriter) {
	w.WriteHeader(http.StatusCreated)
	w.Write(msgCreated)
}

func noContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
	w.Write(msgNoContent)
}

func badRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write(msgBadRequest)
}

func conflict(w http.ResponseWriter) {
	w.WriteHeader(http.StatusConflict)
	w.Write(msgConflict)
}

func serverError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(msgInternalError)
}
