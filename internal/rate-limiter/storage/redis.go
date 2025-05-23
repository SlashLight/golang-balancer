package storage

import (
	"context"
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

const TimeFormat = time.RFC3339

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

		data, err := rl.ReadClient(ctx, userIP)
		var tb rate_limiter.TokenBucket

		if errors.Is(err, my_err.ErrUserNotFound) {
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
			tb = data.TokenBucket
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
				"last_update", time.Now().Format(TimeFormat),
				"capacity", tb.Capacity,
				"rate", tb.Rate,
			).Result()

			return err
		})

		return err
	}, key)

	return allowed, err
}

func (rl *RedisRateLimiter) CreateClient(ctx context.Context, user *rate_limiter.Client) error {
	key := "user:" + user.ClientIP + ":tokens"
	if user.Capacity == 0 {
		user.Capacity = rl.defaultCapacity
	}
	if user.Rate == 0 {
		user.Rate = rl.defaultRate
	}

	return rl.Client.Watch(ctx, func(tx *redis.Tx) error {
		exists, err := tx.Exists(ctx, key).Result()
		if err != nil {
			return err
		}
		if exists != 0 {
			return my_err.ErrUserAlreadyExists
		}

		_, err = rl.Client.HSet(ctx, key,
			"tokens", user.Capacity,
			"last_update", time.Now().Format(TimeFormat),
			"capacity", user.Capacity,
			"rate", user.Rate,
		).Result()

		return err
	})
}

func (rl *RedisRateLimiter) ReadClient(ctx context.Context, userIP string) (*rate_limiter.Client, error) {
	key := "user:" + userIP + ":tokens"
	var client *rate_limiter.Client

	err := rl.Client.Watch(ctx, func(tx *redis.Tx) error {
		exists, err := rl.Client.Exists(ctx, key).Result()
		if err != nil {
			return err
		}
		if exists == 0 {
			return my_err.ErrUserNotFound
		}

		result, err := rl.Client.HGetAll(ctx, key).Result()
		if err != nil {
			return err
		}
		if len(result) == 0 {
			return my_err.ErrUserNotFound
		}

		rateLimit, err := strconv.Atoi(result["rate"])
		if err != nil {
			return err
		}
		capacity, err := strconv.Atoi(result["capacity"])
		if err != nil {
			return err
		}
		lastUpdate, err := time.Parse(TimeFormat, result["last_update"])
		if err != nil {
			return err
		}

		tokens, err := strconv.Atoi(result["tokens"])
		if err != nil {
			return err
		}

		client = &rate_limiter.Client{
			ClientIP: userIP,
			TokenBucket: rate_limiter.TokenBucket{
				Tokens:     tokens,
				LastUpdate: lastUpdate.Unix(),
				Capacity:   capacity,
				Rate:       rateLimit,
			},
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return client, nil
}

func (rl *RedisRateLimiter) UpdateClient(ctx context.Context, newClient *rate_limiter.Client) error {
	key := "user:" + newClient.ClientIP + ":tokens"

	return rl.Client.Watch(ctx, func(tx *redis.Tx) error {
		exists, err := tx.Exists(ctx, key).Result()
		if err != nil {
			return err
		}
		if exists == 0 {
			return my_err.ErrUserNotFound
		}

		if newClient.Capacity != 0 {
			_, err = rl.Client.HSet(ctx, key,
				"capacity", newClient.Capacity,
			).Result()
		}
		if err != nil {
			return err
		}

		if newClient.Rate != 0 {
			_, err = rl.Client.HSet(ctx, key,
				"rate", newClient.Rate,
			).Result()
		}

		return err
	})
}

func (rl *RedisRateLimiter) DeleteClient(ctx context.Context, clientIP string) error {
	key := "user:" + clientIP + ":tokens"

	return rl.Client.Watch(ctx, func(tx *redis.Tx) error {
		exists, err := tx.Exists(ctx, key).Result()
		if err != nil {
			return err
		}
		if exists == 0 {
			return my_err.ErrUserNotFound
		}

		_, err = rl.Client.Del(ctx, key).Result()

		return err
	})
}
