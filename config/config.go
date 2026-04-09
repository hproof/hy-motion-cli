package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	API APIConfig `mapstructure:"api"`
	Auth AuthConfig `mapstructure:"auth"`
}

type APIConfig struct {
	URL      string `mapstructure:"url"`
	Timeout  int    `mapstructure:"timeout"`
}

type AuthConfig struct {
	UserID string `mapstructure:"user_id"`
	Token  string `mapstructure:"token"`
}

var cfg *Config

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("toml")

	viper.SetDefault("api.url", "http://localhost:8000")
	viper.SetDefault("api.timeout", 30)
	viper.SetDefault("auth.user_id", "")
	viper.SetDefault("auth.token", "")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

func Get() *Config {
	return cfg
}
