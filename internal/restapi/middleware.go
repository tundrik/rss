package restapi

import "net/http"


func (e *RestApi) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		defer func() {
			if rcv := recover(); rcv != nil {
				e.log.Error().Any("panic", rcv).Msg("panic recover")
				
				serverError(w)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (e *RestApi) authAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		
		if r.Header.Get("X-Auth-ID") != "35be0a7c-8570-4987-be59-efeac5906d74" {
			forbidden(w)
			return
		}

		next(w, r)
	}
}

func (e *RestApi) authUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		
		if r.Header.Get("X-Auth-ID") == "" {
			unauthorized(w)
			return
		}

		next(w, r)
	}
}
