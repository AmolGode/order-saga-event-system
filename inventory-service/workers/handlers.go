package main

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

func handleOrderRequestedEvent(db *DB, orderEvent OrderRequestedEvent) error {
	err := db.ReserveStock(context.Background(), orderEvent.EventID, orderEvent.ProductID, orderEvent.Qty)

	if errors.Is(err, ErrInsufficientStock) {
		return PublishEvent("inventory-failed", nextEvent(orderEvent))
	}
	if err != nil {
		return err
	}
	if err := PublishEvent("inventory-checked", nextEvent(orderEvent)); err != nil {
		return err
	}
	return PublishEvent("payment-requested", nextEvent(orderEvent))
}

func nextEvent(orderEvent OrderRequestedEvent) OrderRequestedEvent {
	orderEvent.EventID = uuid.New().String()
	return orderEvent
}

func HandleMessage(db *DB, topic string, value []byte) error {
	orderEvent, err := extractOrderEvent(value)
	if err != nil {
		return err
	}

	switch topic {
	case "order-requested":
		return handleOrderRequestedEvent(db, orderEvent)
	case "payment-failed":
		return db.ReleaseStock(context.Background(), orderEvent.EventID, orderEvent.ProductID, orderEvent.Qty)
	}
	return nil
}
