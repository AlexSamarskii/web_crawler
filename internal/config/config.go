package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Crawler  CrawlerConfig  `yaml:"crawler"`
	Database DatebaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
}

type CrawlerConfig struct {
	UserAgent  string `yaml:"user_agent"`
	MaxWorkers int    `yaml:"max_workers"`
	RateLimit  int    `yaml:"rate_limit"`
}

type DatebaseConfig struct {
	ConnStr string `yaml:"conn_str"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"-"`
	DB       int    `yaml:"db"`
}

func LoadConfig() (*Config, error) {
	f, err := os.Open("config.yml")
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, fmt.Errorf("error parsing YAML: %w", err)
	}
	return &cfg, nil
}
