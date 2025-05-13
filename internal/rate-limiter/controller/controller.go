package controller

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	resp "github.com/SlashLight/golang-balancer/internal/api/response"
	"github.com/SlashLight/golang-balancer/internal/logger"
	rate_limiter "github.com/SlashLight/golang-balancer/internal/rate-limiter"
)

type RateLimitRepo interface {
	CreateClient(context.Context, *rate_limiter.Client) error
	ReadClient(context.Context, string) (*rate_limiter.Client, error)
	UpdateClient(context.Context, *rate_limiter.Client) error
	DeleteClient(context.Context, string) error
}

type RateLimitController struct {
	Repo RateLimitRepo
	Log  *slog.Logger
}

func NewRateLimitController(repo RateLimitRepo, log *slog.Logger) *RateLimitController {
	return &RateLimitController{Repo: repo, Log: log}
}

func (c *RateLimitController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.HandleClients(w, r)
}

// TODO: [] добавить errors.Is на проверку того, что пользователь существует
// TODO: [] уменьшить дублирование кода
func (c *RateLimitController) HandleClients(w http.ResponseWriter, r *http.Request) {
	var client *rate_limiter.Client
	var err error

	switch r.Method {
	case http.MethodGet:
		clientID := r.URL.Query().Get("client_id")
		if clientID == "" {
			c.Log.Error("empty client ID")
			err = resp.RespondError(w, http.StatusBadRequest, "no client ID")
			return
		}

		client, err = c.Repo.ReadClient(r.Context(), clientID)
		if err != nil {
			c.Log.Error("error at getting client", logger.Err(err))
			err = resp.RespondError(w, http.StatusInternalServerError, "error at getting client")
			if err != nil {
				c.Log.Error("Error at sending message", logger.Err(err))
			}
			return
		}

		if err := json.NewEncoder(w).Encode(client); err != nil {
			c.Log.Error("error at sending JSON", logger.Err(err))
		}
	case http.MethodPost:
		if err := json.NewDecoder(r.Body).Decode(&client); err != nil {
			c.Log.Error("error at getting client from request body", logger.Err(err))
			return
		}

		if client.ClientIP == "" {
			c.Log.Error("empty client IP")
			resp.RespondError(w, http.StatusBadRequest, "Client IP is empty")
			return
		}
		if err = c.Repo.CreateClient(r.Context(), client); err != nil {
			c.Log.Error("error at creating client", logger.Err(err))
			err = resp.RespondError(w, http.StatusInternalServerError, "error at creating client")
			if err != nil {
				c.Log.Error("Error at sending message", logger.Err(err))
			}
			return
		}

		err = resp.RespondOK(w, http.StatusOK)
		if err != nil {
			c.Log.Error("Error at sending message", logger.Err(err))
		}
	case http.MethodPut:
		if err := json.NewDecoder(r.Body).Decode(&client); err != nil {
			c.Log.Error("error at getting client from request body", logger.Err(err))
			return
		}

		if err = c.Repo.UpdateClient(r.Context(), client); err != nil {
			c.Log.Error("error at updating client", logger.Err(err))
			err = resp.RespondError(w, http.StatusInternalServerError, "error at updating client")
			if err != nil {
				c.Log.Error("Error at sending message", logger.Err(err))
			}
			return
		}

		err = resp.RespondOK(w, http.StatusOK)
		if err != nil {
			c.Log.Error("Error at sending message", logger.Err(err))
		}
	case http.MethodDelete:
		clientID := r.URL.Query().Get("client_id")
		if clientID == "" {
			c.Log.Error("empty client ID")
			err = resp.RespondError(w, http.StatusBadRequest, "no client ID")
			if err != nil {
				c.Log.Error("Error at sending message", logger.Err(err))
			}
			return
		}

		err = c.Repo.DeleteClient(r.Context(), clientID)
		if err != nil {
			c.Log.Error("error at getting client", logger.Err(err))
			err = resp.RespondError(w, http.StatusInternalServerError, "error at getting client")
			if err != nil {
				c.Log.Error("Error at sending message", logger.Err(err))
			}
			return
		}

		err = resp.RespondOK(w, http.StatusOK)
		if err != nil {
			c.Log.Error("Error at sending message", logger.Err(err))
		}

	default:
		err = resp.RespondError(w, http.StatusMethodNotAllowed, "method not allowed")
		if err != nil {
			c.Log.Error("Error at sending message", logger.Err(err))
		}
	}
}
