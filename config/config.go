package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Backends struct {
	Endpoints      []string      `yaml:"endpoints"`
	HealthInterval int `yaml:"health-interval"`
}

type RateLimiter struct {
	DefaultInterval   int `yaml:"default-interval"`
	DefaultCapacity   int           `yaml:"default-capacity"`
	DefaultRefillRate int           `yaml:"default-refill-rate"`
}

type Config struct {
	LoadBalancerPort string      `yaml:"server-port"`
	APIport          string      `yaml:"api-port"`
	Backends         Backends    `yaml:"backends"`
	RateLimiter      RateLimiter `yaml:"rate-limiter"`
	DatabaseURL      string      `yaml:"database"`
}

func NewConfig() *Config {
	c := &Config{}
	return c
}

func ReadConfig(path string) (*Config, error) {
	var config Config
	data, err := os.ReadFile(path)
	if err != nil {
		return &config, err
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return &config, err
	}

	return &config, nil
}
