package main

import (
	"encoding/json"
	"fmt"
)

type OrderRequestedEvent struct {
	EventID     string  `json:"event_id"`
	OrderID     string  `json:"order_id"`
	CustomerID  string  `json:"customer_id"`
	ProductID   string  `json:"product_id"`
	Qty         int     `json:"qty"`
	UnitPrice   float64 `json:"unit_price"`
	TotalAmount float64 `json:"total_amount"`
}

func extractOrderEvent(value []byte) (OrderRequestedEvent, error) {
	var event OrderRequestedEvent
	err := json.Unmarshal(value, &event)
	return event, err
}

func buildConnString(cfg Config) string {
	return fmt.Sprintf(
		"host=%s port=5432 user=%s password=%s dbname=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)
}
