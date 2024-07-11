package restapi

import (
	"net/http"
)

// globalMiddleware ловит и логгирует панику
func (e *RestApi) globalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if rcv := recover(); rcv != nil {
				e.log.Error().Any("panic", rcv).Msg("panic recover")
				e.responseJson(w, "internal server error", 500, nil)
			}
		}()

		next.ServeHTTP(w, req)
	})
}

// authAdminMiddleware допускает только админа
func (e *RestApi) authAdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		token := req.Header.Get("X-Auth-ID")
		if token != "35be0a7c-8570-4987-be59-efeac5906d74" {
			e.responseJson(w, "Forbidden", 403, nil)
			return
		}
		next(w, req)
	}
}

// authUserMiddleware допускает только авторизованных
func (e *RestApi) authUserMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("X-Auth-ID") == "" {
			e.responseJson(w, "Unauthorized", 401, nil)
			return
		}
		next(w, req)
	}
}
