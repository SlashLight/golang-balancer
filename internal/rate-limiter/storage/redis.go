package storage

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/SlashLight/golang-balancer/internal/config"
	"github.com/SlashLight/golang-balancer/internal/rate-limiter"
	"github.com/SlashLight/golang-balancer/pkg/my_err"
)

type RedisRateLimiter struct {
	Client          *redis.Client
	defaultCapacity int
	defaultRate     int
}

func NewRedisRateLimiter(cfg *config.Config) *RedisRateLimiter {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
		PoolSize:     cfg.Redis.Pool,
	})

	return &RedisRateLimiter{
		Client:          client,
		defaultCapacity: cfg.DefaultCapacity,
		defaultRate:     cfg.DefaultRate,
	}
}

func (rl *RedisRateLimiter) Allow(ctx context.Context, userIP string) (bool, error) {
	key := "user:" + userIP + ":tokens"

	var allowed bool
	err := rl.Client.Watch(ctx, func(tx *redis.Tx) error {

		data, err := tx.Get(ctx, key).Bytes()
		var tb rate_limiter.TokenBucket

		if errors.Is(err, redis.Nil) {
			tb = rate_limiter.TokenBucket{
				Tokens:     rl.defaultCapacity - 1,
				LastUpdate: time.Now().Unix(),
				Capacity:   rl.defaultCapacity,
				Rate:       rl.defaultRate,
			}
			allowed = true
		} else if err != nil {
			return err
		} else {
			err := json.Unmarshal(data, &tb)
			if err != nil {
				return err
			}

			now := time.Now().Unix()
			since := now - tb.LastUpdate
			newTokens := tb.Tokens + int(since)*tb.Rate
			newTokens = min(newTokens, tb.Capacity)

			if newTokens < 1 {
				allowed = false
				return nil
			}

			tb.Tokens = newTokens - 1
			tb.LastUpdate = now
			allowed = true

		}

		_, err = tx.TxPipelined(ctx, func(pipeliner redis.Pipeliner) error {
			_, err := pipeliner.HSet(ctx, key,
				"tokens", tb.Tokens,
				"last_update", time.Now(),
				"capacity", tb.Capacity,
				"rate", tb.Rate,
			).Result()

			return err
		})

		return err
	}, key)

	return allowed, err
}

func (rl *RedisRateLimiter) CreateUser(ctx context.Context, user rate_limiter.Client) error {
	key := "user:" + user.ClientIP + ":tokens"

	_, err := rl.Client.HSet(ctx, key,
		"tokens", user.Capacity,
		"last_update", time.Now(),
		"capacity", user.Capacity,
		"rate", user.Rate,
	).Result()

	return err
}

func (rl *RedisRateLimiter) ReadUser(ctx context.Context, userIP string) (*rate_limiter.Client, error) {
	key := "user:" + userIP + ":tokens"

	result, err := rl.Client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, my_err.ErrUserNotFound
	}

	rateLimit, err := strconv.Atoi(result["rate_limit"])
	if err != nil {
		return nil, err
	}
	capacity, err := strconv.Atoi(result["capacity"])
	if err != nil {
		return nil, err
	}
	lastUpdate, err := strconv.Atoi(result["last_update"])
	if err != nil {
		return nil, err
	}
	tokens, err := strconv.Atoi(result["tokens"])
	if err != nil {
		return nil, err
	}

	return &rate_limiter.Client{
		ClientIP: userIP,
		TokenBucket: rate_limiter.TokenBucket{
			Tokens:     tokens,
			LastUpdate: int64(lastUpdate),
			Capacity:   capacity,
			Rate:       rateLimit,
		},
	}, nil
}

func (rl *RedisRateLimiter) UpdateClient(ctx context.Context, newClient rate_limiter.Client) error {
	key := "user:" + newClient.ClientIP + ":tokens"
	_, err := rl.Client.HSet(ctx, key,
		"tokens", newClient.Tokens,
		"last_update", newClient.LastUpdate,
		"capacity", newClient.Capacity,
		"rate", newClient.Rate,
	).Result()

	return err
}

func (rl *RedisRateLimiter) DeleteClient(ctx context.Context, clientIP string) error {
	key := "user:" + clientIP + ":tokens"
	_, err := rl.Client.Del(ctx, key).Result()

	return err
}
