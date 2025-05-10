package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env       string   `yaml:"env" env-default:"local"`
	Algorithm string   `yaml: "algorithm" env-required:"true"`
	Backends  []string `yaml: "backends" env-required:"true"`
	Port      int      `yaml:"port" env-required: "true"`
	Retries   int      `yaml:"retries" env-default:"3"`
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
