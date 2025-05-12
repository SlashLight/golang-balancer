package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

func RespondError(w http.ResponseWriter, code int, message string) error {
	w.Header().Set("Content-Type", "application/json:charset=UTF-8")
	w.WriteHeader(code)

	resp := &Response{
		Code:    code,
		Message: message,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		return err
	}

	return nil
}
