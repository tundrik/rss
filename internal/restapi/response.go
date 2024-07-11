package restapi

import (
	"encoding/json"
	"net/http"
)

const (
	succes = "succes"
	msgAlreadyExists = "feed url already exists"
)

var (
	msgInternalError = []byte(`{"code": 500, "message": "Internal Server Error"}`)
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Body    any    `json:"body,omitempty"`
}

// я тут эксперементировал с общим Response на все ручки
func (e *RestApi) responseJson(w http.ResponseWriter, msg string, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	if code != 200 && code != 201 && code != 204 {
		e.log.Error().Str("msg", msg).Msg("response json")
	}
	res := Response{
		Code: code,
		Message: msg,
		Body:   body,
	}
	b, err := json.Marshal(res)
	if err != nil {
		e.log.Err(err).Any("res", res).Msg("json marshal")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(msgInternalError)
		return
	}
	w.WriteHeader(code)
	w.Write(b)
}

