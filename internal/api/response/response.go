package response

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/SlashLight/golang-balancer/internal/logger"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

const messageOK = "OK"

func RespondError(w http.ResponseWriter, code int, message string, log *slog.Logger) {
	w.Header().Set("Content-Type", "application/json:charset=UTF-8")
	w.WriteHeader(code)

	resp := &Response{
		Code:    code,
		Message: message,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error("error at sending message", logger.Err(err))
	}
}

func RespondOK(w http.ResponseWriter, code int, log *slog.Logger) {
	w.Header().Set("Content-Type", "application/json:charset=UTF-8")
	w.WriteHeader(code)

	resp := &Response{
		Code:    code,
		Message: messageOK,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error("error at sending message", logger.Err(err))
	}
}
