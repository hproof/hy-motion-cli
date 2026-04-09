package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const configFileName = ".hy-motion-cli.toml"

type Config struct {
	API  APIConfig  `mapstructure:"api"`
	Auth AuthConfig `mapstructure:"auth"`
}

type APIConfig struct {
	URL     string `mapstructure:"url"`
	Timeout int    `mapstructure:"timeout"`
}

type AuthConfig struct {
	UserID string `mapstructure:"user_id"`
	Token  string `mapstructure:"token"`
}

var cfg *Config

// Load 加载配置，优先从当前目录读取，找不到则从 home 目录读取
func Load() (*Config, error) {
	var configPath string

	// 优先查找当前目录
	cwd, err := os.Getwd()
	if err == nil {
		candidate := filepath.Join(cwd, configFileName)
		if _, err := os.Stat(candidate); err == nil {
			configPath = candidate
		}
	}

	// 当前目录找不到，查找 home 目录
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			candidate := filepath.Join(home, configFileName)
			if _, err := os.Stat(candidate); err == nil {
				configPath = candidate
			}
		}
	}

	if configPath == "" {
		return nil, fmt.Errorf("未找到配置文件 %s（当前目录和 home 目录下均不存在）", configFileName)
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("toml")

	viper.SetDefault("api.url", "http://localhost:8000")
	viper.SetDefault("api.timeout", 30)
	viper.SetDefault("auth.user_id", "")
	viper.SetDefault("auth.token", "")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}

	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 修复缺少协议前缀的 URL
	if !strings.HasPrefix(cfg.API.URL, "http://") && !strings.HasPrefix(cfg.API.URL, "https://") {
		cfg.API.URL = "http://" + cfg.API.URL
	}

	return cfg, nil
}
