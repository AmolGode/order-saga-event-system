package main

import (
	"context"
	"math/rand"

	"github.com/google/uuid"
)

func handlePaymentRequestedEvent(db *DB, event OrderRequestedEvent) error {
	success := simulatePaymentGatewayCharge(event.TotalAmount)

	if err := db.ProcessPayment(context.Background(), event.EventID, event.OrderID, event.TotalAmount, success); err != nil {
		return err
	}

	next := nextEvent(event)
	if success {
		return PublishEvent("payment-success", next)
	}
	return PublishEvent("payment-failed", next)
}

// simulatePaymentGatewayCharge stands in for a real payment gateway call.
// Occasionally fails (~1 in 100) to simulate a declined card.
func simulatePaymentGatewayCharge(amount float64) bool {
	return rand.Intn(100) != 0
}

func nextEvent(event OrderRequestedEvent) OrderRequestedEvent {
	event.EventID = uuid.New().String()
	return event
}

func HandleMessage(db *DB, topic string, value []byte) error {
	event, err := extractOrderEvent(value)
	if err != nil {
		return err
	}

	switch topic {
	case "payment-requested":
		return handlePaymentRequestedEvent(db, event)
	}
	return nil
}
