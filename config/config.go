package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Broker   string
	ClientID string
	Username string
	Password string
}

func LoadConfig() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	config := Config{
		Broker:   os.Getenv("MQTT_BROKER"),
		ClientID: os.Getenv("MQTT_CLIENT_ID"),
		Username: os.Getenv("MQTT_USERNAME"),
		Password: os.Getenv("MQTT_PASSWORD"),
	}

	missing := make([]string, 0, 4)
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
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing environment variables: %s", strings.Join(missing, ", "))
	}

	return config, nil
}
