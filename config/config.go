package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Broker       string
	ClientID     string
	Username     string
	Password     string
	PostgresDB   string
	PostgresUser string
	PostgresPass string
	PostgresPort string
	PostgresHost string
}

func LoadConfig() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	config := Config{
		Broker:       os.Getenv("MQTT_BROKER"),
		ClientID:     os.Getenv("MQTT_CLIENT_ID"),
		Username:     os.Getenv("MQTT_USERNAME"),
		Password:     os.Getenv("MQTT_PASSWORD"),
		PostgresDB:   os.Getenv("POSTGRES_DB"),
		PostgresUser: os.Getenv("POSTGRES_USER"),
		PostgresPass: os.Getenv("POSTGRES_PASSWORD"),
		PostgresPort: os.Getenv("POSTGRES_PORT"),
		PostgresHost: os.Getenv("POSTGRES_HOST"),
	}
	if config.PostgresHost == "" {
		config.PostgresHost = "localhost"
	}

	missing := make([]string, 0, 8)
	if config.Broker == "" {
		missing = append(missing, "MQTT_BROKER")
	}
	if config.ClientID == "" {
		missing = append(missing, "MQTT_CLIENT_ID")
	}
	if config.Username == "" {
		missing = append(missing, "MQTT_USERNAME")
	}
	if config.Password == "" {
		missing = append(missing, "MQTT_PASSWORD")
	}
	if config.PostgresDB == "" {
		missing = append(missing, "POSTGRES_DB")
	}
	if config.PostgresUser == "" {
		missing = append(missing, "POSTGRES_USER")
	}
	if config.PostgresPass == "" {
		missing = append(missing, "POSTGRES_PASSWORD")
	}
	if config.PostgresPort == "" {
		missing = append(missing, "POSTGRES_PORT")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing environment variables: %s", strings.Join(missing, ", "))
	}

	return config, nil
}
