package render

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Message string `json:"message"`
	Time    string `json:"time"`
	Data    any    `json:"data"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	Time  string `json:"time"`
}

func SendResponseJSON(w http.ResponseWriter, code int, data *Response) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(data)
}

func SendError(w http.ResponseWriter, code int, data *ErrorResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(data)
}
