package main

import "fmt"

func buildConnString(cfg Config) string {
	return fmt.Sprintf(
		"host=%s port=5432 user=%s password=%s dbname=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)
}
