package restapi

import (
	"encoding/json"
	"net/http"
)

// addFeed добавляет новый RSS источник.
func (e *RestApi) addFeed(w http.ResponseWriter, req *http.Request) {
	feedUrl := req.PostFormValue("feed_url")

	if err := e.repo.AddFeed(feedUrl); err != nil {
		e.log.Err(err).Msg("add feed")
		conflict(w)
		return
	}
	created(w)
}

// available возвращает список доступных RSS каналов.
func (e *RestApi) available(w http.ResponseWriter, req *http.Request) {
	feeds, err := e.repo.Feed()
	if err != nil {
		e.log.Err(err).Msg("available")
		serverError(w)
		return
	}
	b, err := json.Marshal(feeds)
	if err != nil {
		e.log.Err(err).Msg("available")
		serverError(w)
		return
	}
	w.Write(b)
}

// subscribe подписывает пользователя на RSS канал.
func (e *RestApi) subscribe(w http.ResponseWriter, req *http.Request) {
	personPk := req.Header.Get("X-Auth-ID")
	feedPk := req.PostFormValue("feed_pk")

	if err := e.repo.Subscribe(personPk, feedPk); err != nil {
		e.log.Err(err).Msg("subscribe")
		badRequest(w)
		return
	}
	noContent(w)
}

// unsubscribe отписывает пользователя на RSS канал.
func (e *RestApi) unsubscribe(w http.ResponseWriter, req *http.Request) {
	personPk := req.Header.Get("X-Auth-ID")
	feedPk := req.PostFormValue("feed_pk")

	if err := e.repo.Unsubscribe(personPk, feedPk); err != nil {
		e.log.Err(err).Msg("unsubscribe")
		badRequest(w)
		return
	}
	noContent(w)
}

// article возвращает список статей для пользователя.
func (e *RestApi) article(w http.ResponseWriter, req *http.Request) {
	userUuid := req.Header.Get("X-Auth-ID")
	entities, err := e.repo.Article(userUuid)
	if err != nil {
		e.log.Err(err).Msg("article")
		serverError(w)
		return
	}
	b, err := json.Marshal(entities)
	if err != nil {
		e.log.Err(err).Msg("article")
		serverError(w)
		return
	}
	w.Write(b)
}
