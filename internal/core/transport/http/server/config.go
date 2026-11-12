package core_http_server

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr            string        `envconfig:"ADDR" default:"127.0.0.1:8080"`
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"10s"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("HTTP", &config); err != nil {
		return Config{}, fmt.Errorf("load HTTP server config: %w", err)
	}

	_, portText, err := net.SplitHostPort(config.Addr)
	if err != nil {
		return Config{}, fmt.Errorf("HTTP_ADDR must have host:port format: %w", err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("HTTP_ADDR port must be a number from 1 to 65535")
	}
	if config.ShutdownTimeout <= 0 {
		return Config{}, fmt.Errorf("HTTP_SHUTDOWN_TIMEOUT must be positive")
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get HTTP server config: %w", err))
	}
	return config
}
