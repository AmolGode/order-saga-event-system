package main

import (
	"context"
	"log"
	"math/rand"

	"github.com/google/uuid"
)

func handlePaymentRequestedEvent(db *DB, event OrderRequestedEvent) error {
	success := simulatePaymentGatewayCharge(event.TotalAmount)

	status, err := db.ProcessPayment(context.Background(), event.EventID, event.OrderID, event.TotalAmount, success)
	if err != nil {
		// Outcome is genuinely unknown here — nothing was recorded, so the
		// order can't be left hanging forever. order-service already knows
		// how to react to "payment-failed" (see
		// order-service/workers/handlers.go), so reuse that instead of a
		// separate failure channel. The message still goes to
		// payment-requested-dlq afterward for forensics/replay.
		if pubErr := PublishEvent("payment-failed", event.OrderID, nextEvent(event, "payment-failed")); pubErr != nil {
			log.Println("failed to notify order-service of processing failure:", pubErr)
		}
		return err
	}

	// status is now known and safely stored. If the publish below fails,
	// don't guess a fallback outcome — let it hit the DLQ. ProcessPayment's
	// idempotency check means a replay of this exact message reproduces the
	// same correct status and republishes it; force-publishing
	// "payment-failed" here could falsely cancel an order whose payment
	// actually succeeded.
	if status == "SUCCESS" {
		return PublishEvent("payment-success", event.OrderID, nextEvent(event, "payment-success"))
	}
	return PublishEvent("payment-failed", event.OrderID, nextEvent(event, "payment-failed"))
}

// simulatePaymentGatewayCharge stands in for a real payment gateway call.
// Occasionally fails (~1 in 100) to simulate a declined card.
func simulatePaymentGatewayCharge(amount float64) bool {
	return rand.Intn(100) != 0
}

// nextEvent derives the downstream event_id deterministically from the
// source event_id + the step name, instead of a fresh random UUID. A retried
// payment-requested delivery (see retry.go) re-runs this handler from
// scratch; without a deterministic ID here, each retry would mint a
// brand-new "payment-success"/"payment-failed" event that order-service's
// and inventory-service's own idempotency checks can't recognize as a
// duplicate — risking the order/inventory state being re-applied.
func nextEvent(event OrderRequestedEvent, step string) OrderRequestedEvent {
	event.EventID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(event.EventID+":"+step)).String()
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
