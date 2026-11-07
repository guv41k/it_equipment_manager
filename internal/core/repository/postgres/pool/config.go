package core_postgres_pool

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host     string        `envconfig:"HOST" default:"127.0.0.1"`
	Port     string        `envconfig:"PORT" default:"5432"`
	User     string        `envconfig:"USER" required:"true"`
	Password string        `envconfig:"PASSWORD" required:"true"`
	Database string        `envconfig:"DB" required:"true"`
	Timeout  time.Duration `envconfig:"TIMEOUT" default:"5s"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("POSTGRES", &config); err != nil {
		return Config{}, fmt.Errorf("load postgres config: %w", err)
	}

	if strings.TrimSpace(config.Host) == "" ||
		strings.TrimSpace(config.User) == "" ||
		strings.TrimSpace(config.Password) == "" ||
		strings.TrimSpace(config.Database) == "" {
		return Config{}, fmt.Errorf("postgres host, user, password and database must not be empty")
	}

	port, err := strconv.Atoi(config.Port)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("POSTGRES_PORT must be a number from 1 to 65535")
	}
	if config.Timeout <= 0 {
		return Config{}, fmt.Errorf("POSTGRES_TIMEOUT must be positive")
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get postgres connection pool config: %w", err))
	}
	return config
}
