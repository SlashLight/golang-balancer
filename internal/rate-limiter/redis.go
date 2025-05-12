package rate_limiter

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/SlashLight/golang-balancer/internal/config"
)

type RedisRateLimiter struct {
	client          *redis.Client
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
		client:          client,
		defaultCapacity: cfg.DefaultCapacity,
		defaultRate:     cfg.DefaultRate,
	}
}

func (rl *RedisRateLimiter) Allow(ctx context.Context, userIP string) (bool, error) {
	key := "user:" + userIP + ":tokens"

	var allowed bool
	err := rl.client.Watch(ctx, func(tx *redis.Tx) error {

		data, err := tx.Get(ctx, key).Bytes()
		var tb TokenBucket

		if errors.Is(err, redis.Nil) {
			tb = TokenBucket{
				Tokens:     rl.defaultCapacity - 1,
				LastUpdate: time.Now().Unix(),
				Capacity:   rl.defaultCapacity,
				Rate:       rl.defaultRate,
			}
			allowed = true
		} else if err != nil {
			return err
		} else {
			json.Unmarshal(data, &tb)

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

		newData, err := json.Marshal(tb)
		if err != nil {
			return err
		}

		_, err = tx.TxPipelined(ctx, func(pipeliner redis.Pipeliner) error {
			pipeliner.Set(ctx, key, newData, 0)
			return nil
		})

		return err
	}, key)

	return allowed, err
}
