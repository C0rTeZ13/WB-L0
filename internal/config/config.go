package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env        string     `yaml:"env" env-default:"local"`
	DBHost     string     `yaml:"db_host" env-required:"true"`
	DBPort     int        `yaml:"db_port" env-required:"true"`
	DBUser     string     `yaml:"db_user" env-required:"true"`
	DBPassword string     `yaml:"db_password" env-required:"true"`
	DBName     string     `yaml:"db_name" env-required:"true"`
	HTTPServer HTTPServer `yaml:"http_server"`
	Kafka      Kafka      `yaml:"kafka"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type Kafka struct {
	Brokers []string `yaml:"brokers" env-required:"true"`
	Topic   string   `yaml:"topic" env-required:"true"`
	GroupID string   `yaml:"group_id" env-required:"true"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("CONFIG_PATH does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Error reading config: %s", err)
	}

	return &cfg
}
