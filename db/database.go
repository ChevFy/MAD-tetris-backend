package db

import (
	"database/sql"
	"fmt"

	"mqtt-tetris-backend/config"

	_ "github.com/lib/pq"
)

func Open(appConfig config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		appConfig.PostgresHost,
		appConfig.PostgresPort,
		appConfig.PostgresUser,
		appConfig.PostgresPass,
		appConfig.PostgresDB,
	)

	database, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := database.Ping(); err != nil {
		database.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return database, nil
}
