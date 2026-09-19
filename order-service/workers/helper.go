package main

import (
	"encoding/json"
	"fmt"
)

type OrderEvent struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

func extractOrderID(value []byte) string {
	var event OrderEvent
	json.Unmarshal(value, &event)
	return event.OrderID
}

func buildConnString(cfg Config) string {
	return fmt.Sprintf(
		"host=%s port=5432 user=%s password=%s dbname=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)
}
