package restapi

import (
	"errors"
	"net/http"
	"rss/internal/repository"
)

// available возвращает список доступных RSS каналов.
func (e *RestApi) available(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	feeds, err := e.uc.Available(ctx)
	if err != nil {
		e.responseJson(w, "internal server error", 500, nil)
		return
	}
	e.responseJson(w, succes, 200, feeds)
}

// addFeed добавляет новый RSS источник.
func (e *RestApi) addFeed(w http.ResponseWriter, req *http.Request) {
	feedUrl := req.PostFormValue("feed_url")
	if feedUrl == "" {
		e.responseJson(w, "required feed_url", 400, nil)
		return
	}
	ctx := req.Context()

	if err := e.uc.AddFeed(ctx, feedUrl); err != nil {
		if errors.Is(err, repository.ErrFeedExists) {
			e.responseJson(w, msgAlreadyExists, 400, nil)
			return
		}
		e.responseJson(w, "internal server error", 500, nil)
		return
	}
	e.responseJson(w, "created", 201, nil)
}

// subscribe подписывает пользователя на RSS канал.
func (e *RestApi) subscribe(w http.ResponseWriter, req *http.Request) {
	personPk := req.Header.Get("X-Auth-ID")
	feedPk := req.PostFormValue("feed_pk")
	if feedPk == "" {
		e.responseJson(w, "required feed_pk", 400, nil)
		return
	}
	ctx := req.Context()

	if err := e.uc.Subscribe(ctx, personPk, feedPk); err != nil {
		switch {
		case errors.Is(err, repository.ErrReSubscription):
			e.responseJson(w, "no content", 204, nil)
		case errors.Is(err, repository.ErrNoSubscription):
			e.responseJson(w, "feed_pk not found", 400, nil)
		default:
			e.responseJson(w, "internal server error", 500, nil)
		}
		return
	}
	e.responseJson(w, "created", 201, nil)
}

// unsubscribe отписывает пользователя от RSS канала.
func (e *RestApi) unsubscribe(w http.ResponseWriter, req *http.Request) {
	personPk := req.Header.Get("X-Auth-ID")
	feedPk := req.PostFormValue("feed_pk")
	if feedPk == "" {
		e.responseJson(w, "required feed_pk", 400, nil)
		return
	}
	ctx := req.Context()

	if err := e.uc.Unsubscribe(ctx, personPk, feedPk); err != nil {
		e.responseJson(w, "internal server error", 500, nil)
		return
	}
	e.responseJson(w, "no content", 204, nil)
}

// article возвращает список статей для пользователя.
func (e *RestApi) article(w http.ResponseWriter, req *http.Request) {
	personPk := req.Header.Get("X-Auth-ID")
	ctx := req.Context()

	entities, err := e.uc.Article(ctx, personPk)
	if err != nil {
		if errors.Is(err, repository.ErrArticleNotFound) {
			e.responseJson(w, err.Error(), 404, nil)
			return
		}
		e.responseJson(w, "internal server error", 500, nil)
		return
	}
	e.responseJson(w, succes, 200, entities)
}
