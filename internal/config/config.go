package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env           string `yaml:"env" env-default:"local"`
	Balancer      `yaml:"balancer"`
	HealthChecker `yaml:"healthChecker"`
	Redis         `yaml:"redis"`
	RateLimit     `yaml:"rate-limit"`
}

type Balancer struct {
	Port      int      `yaml:"port" env-required:"true"`
	Backends  []string `yaml:"backends" env-required:"true"`
	Retries   int      `yaml:"retries" env-default:"3"`
	Algorithm string   `yaml:"algorithm" env-required:"true"`
}

type HealthChecker struct {
	Interval time.Duration `yaml:"interval" env-default:"10s"`
	CheckURL string        `yaml:"checkURL" env-required:"true"`
}

type Redis struct {
	Addr         string        `yaml:"addr" env-required:"true"`
	DialTimeout  time.Duration `yaml:"dialTimeout" env-default:"5s"`
	ReadTimeout  time.Duration `yaml:"readTimeout" env-default:"3s"`
	WriteTimeout time.Duration `yaml:"writeTimeout" env-default:"3s"`
	Pool         int           `yaml:"pool" env-default:"100"`
}

type RateLimit struct {
	DefaultCapacity int `yaml:"defaultCapacity" env-default:"20"`
	DefaultRate     int `yaml:"defaultRate" env-default:"2"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
