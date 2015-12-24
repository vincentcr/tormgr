package main

import (
	"log"
	"os"
)

type Config struct {
	PublicURL    string
	BitlyAPIKey  string
	Postgres     PGConfig
	RedisAddr    string
	RabbitMQAddr string
}
type PGConfig struct {
	Addr     string
	Database string
	User     string
	Password string
}

func loadConfig() (Config, error) {
	config, err := loadConfigFromEnv()
	log.Printf("using config: %#v", config)
	return config, err
}

func loadConfigFromEnv() (Config, error) {
	return Config{
		PublicURL:    os.Getenv("API_PUBLIC_URL"),
		BitlyAPIKey:  os.Getenv("BITLY_API_KEY"),
		RedisAddr:    "redis:6379",
		RabbitMQAddr: "rabbitmq",
		Postgres: PGConfig{
			Addr:     "postgres:5432",
			Database: os.Getenv("POSTGRES_ENV_DB_NAME"),
			User:     os.Getenv("POSTGRES_ENV_DB_USER"),
			Password: os.Getenv("POSTGRES_ENV_DB_PASSWD"),
		},
	}, nil
}
