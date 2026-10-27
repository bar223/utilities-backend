// Package redis — клиент Redis для чёрного списка JWT: после logout токен
// хранится в Redis до конца своего срока жизни и больше не принимается.
package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"utilities-backend/internal/app/config"
)

// servicePrefix и jwtPrefix — префиксы ключей, чтобы ключи разных сервисов
// в одном Redis не пересекались: utilities.jwt.<токен>.
const (
	servicePrefix = "utilities."
	jwtPrefix     = "jwt."
)

// Client — подключение к Redis.
type Client struct {
	client *redis.Client
}

// New подключается к Redis и проверяет соединение командой PING.
func New(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:        cfg.Host + ":" + strconv.Itoa(cfg.Port),
		Username:    cfg.User,
		Password:    cfg.Password,
		DB:          0,
		DialTimeout: cfg.DialTimeout,
		ReadTimeout: cfg.ReadTimeout,
	})
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("нет соединения с redis: %w", err)
	}
	return &Client{client: client}, nil
}

func getJWTKey(token string) string {
	return servicePrefix + jwtPrefix + token
}

// WriteJWTToBlacklist добавляет токен в чёрный список на ttl — до момента,
// когда он истёк бы сам.
func (c *Client) WriteJWTToBlacklist(ctx context.Context, token string, ttl time.Duration) error {
	return c.client.Set(ctx, getJWTKey(token), true, ttl).Err()
}

// IsJWTInBlacklist проверяет, разлогинен ли токен.
func (c *Client) IsJWTInBlacklist(ctx context.Context, token string) (bool, error) {
	err := c.client.Get(ctx, getJWTKey(token)).Err()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
