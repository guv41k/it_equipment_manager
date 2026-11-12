package core_logger

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Level  string `envconfig:"LEVEL" required:"true"`
	Folder string `envconfig:"FOLDER" required:"true"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("LOGGER", &config); err != nil {
		return Config{}, fmt.Errorf("load logger config: %w", err)
	}
	if strings.TrimSpace(config.Level) == "" || strings.TrimSpace(config.Folder) == "" {
		return Config{}, fmt.Errorf("LOGGER_LEVEL and LOGGER_FOLDER must not be empty")
	}
	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get logger config: %w", err))
	}
	return config
}
