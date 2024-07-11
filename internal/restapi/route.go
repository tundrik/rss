package restapi

import "net/http"

func (e *RestApi) registerRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /add", e.authUserMiddleware(e.authAdminMiddleware(e.addFeed)))
	mux.HandleFunc("GET /{$}", e.authUserMiddleware(e.available))
	mux.HandleFunc("PUT /subscribe", e.authUserMiddleware(e.subscribe))
	mux.HandleFunc("PUT /unsubscribe", e.authUserMiddleware(e.unsubscribe))
	mux.HandleFunc("GET /article", e.authUserMiddleware(e.article))

	return e.globalMiddleware(mux)
}