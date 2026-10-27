// Package config читает конфигурацию приложения из config/config.toml и
// секреты из .env.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config — параметры HTTP-сервера, JWT и Redis.
type Config struct {
	ServiceHost string
	ServicePort int

	JWT   JWTConfig
	Redis RedisConfig
}

// JWTConfig — подпись и срок жизни токена. Секрет берётся из .env
// (JWT_SECRET), срок жизни — из config.toml.
type JWTConfig struct {
	Secret    string `mapstructure:"-"`
	ExpiresIn time.Duration
}

// RedisConfig — подключение к Redis, где хранится чёрный список токенов.
// Адрес и пароль — из .env, таймауты — из config.toml.
type RedisConfig struct {
	Host        string `mapstructure:"-"`
	Port        int    `mapstructure:"-"`
	User        string `mapstructure:"-"`
	Password    string `mapstructure:"-"`
	DialTimeout time.Duration
	ReadTimeout time.Duration
}

// NewConfig читает config/config.toml и .env и возвращает конфигурацию.
func NewConfig() (*Config, error) {
	configName := "config"
	_ = godotenv.Load()
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	cfg.JWT.Secret = os.Getenv("JWT_SECRET")
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET не задан в .env")
	}

	cfg.Redis.Host = os.Getenv("REDIS_HOST")
	port, err := strconv.Atoi(os.Getenv("REDIS_PORT"))
	if err != nil {
		return nil, fmt.Errorf("REDIS_PORT должен быть числом: %w", err)
	}
	cfg.Redis.Port = port
	cfg.Redis.User = os.Getenv("REDIS_USER")
	cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")

	log.Info("config parsed")
	return cfg, nil
}
