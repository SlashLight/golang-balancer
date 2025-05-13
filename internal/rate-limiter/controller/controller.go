package controller

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	resp "github.com/SlashLight/golang-balancer/internal/api/response"
	"github.com/SlashLight/golang-balancer/internal/logger"
	rate_limiter "github.com/SlashLight/golang-balancer/internal/rate-limiter"
	"github.com/SlashLight/golang-balancer/pkg/my_err"
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

func (c *RateLimitController) HandleClients(w http.ResponseWriter, r *http.Request) {
	var client *rate_limiter.Client
	var err error

	switch r.Method {
	case http.MethodGet:
		clientID := r.URL.Query().Get("client_id")
		if clientID == "" {
			c.Log.Error("empty client ID")
			resp.RespondError(w, http.StatusBadRequest, "no client ID", c.Log)
			return
		}

		client, err = c.Repo.ReadClient(r.Context(), clientID)
		if err != nil {
			c.Log.Error("error at getting client", logger.Err(err))
			if errors.Is(err, my_err.ErrUserNotFound) {
				resp.RespondError(w, http.StatusBadRequest, "client not found", c.Log)
				return
			}

			resp.RespondError(w, http.StatusInternalServerError, "error at getting client", c.Log)
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
			resp.RespondError(w, http.StatusBadRequest, "Client IP is empty", c.Log)
			return
		}

		if err = c.Repo.CreateClient(r.Context(), client); err != nil {
			c.Log.Error("error at creating client", logger.Err(err))
			if errors.Is(err, my_err.ErrUserAlreadyExists) {
				resp.RespondError(w, http.StatusBadRequest, "client already exists", c.Log)
				return
			}

			resp.RespondError(w, http.StatusInternalServerError, "error at creating client", c.Log)
			return
		}

		resp.RespondOK(w, http.StatusOK, c.Log)
	case http.MethodPut:
		if err := json.NewDecoder(r.Body).Decode(&client); err != nil {
			c.Log.Error("error at getting client from request body", logger.Err(err))
			return
		}

		if err = c.Repo.UpdateClient(r.Context(), client); err != nil {
			c.Log.Error("error at updating client", logger.Err(err))
			if errors.Is(err, my_err.ErrUserNotFound) {
				resp.RespondError(w, http.StatusBadRequest, "client not exists", c.Log)
				return
			}
			resp.RespondError(w, http.StatusInternalServerError, "error at updating client", c.Log)
			return
		}

		resp.RespondOK(w, http.StatusOK, c.Log)
	case http.MethodDelete:
		clientID := r.URL.Query().Get("client_id")
		if clientID == "" {
			c.Log.Error("empty client ID")
			resp.RespondError(w, http.StatusBadRequest, "no client ID", c.Log)
			return
		}

		err = c.Repo.DeleteClient(r.Context(), clientID)
		if err != nil {
			c.Log.Error("error at getting client", logger.Err(err))
			if errors.Is(err, my_err.ErrUserNotFound) {
				resp.RespondError(w, http.StatusBadRequest, "client not exists", c.Log)
				return
			}

			resp.RespondError(w, http.StatusInternalServerError, "error at getting client", c.Log)
			return
		}

		resp.RespondOK(w, http.StatusOK, c.Log)

	default:
		resp.RespondError(w, http.StatusMethodNotAllowed, "method not allowed", c.Log)
	}
}
