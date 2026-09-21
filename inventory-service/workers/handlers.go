package main

import (
	"context"
	"log"

	"github.com/google/uuid"
)

func handleOrderRequestedEvent(db *DB, orderEvent OrderRequestedEvent) error {
	status, err := db.ReserveStock(context.Background(), orderEvent.EventID, orderEvent.ProductID, orderEvent.Qty)
	if err != nil {
		// Outcome is genuinely unknown here — nothing was recorded, so the
		// order can't be left hanging forever. order-service already knows
		// how to react to "inventory-failed" (see
		// order-service/workers/handlers.go), so reuse that instead of a
		// separate failure channel. The message still goes to
		// order-requested-dlq afterward for forensics/replay.
		if pubErr := PublishEvent("inventory-failed", nextEvent(orderEvent, "inventory-failed")); pubErr != nil {
			log.Println("failed to notify order-service of processing failure:", pubErr)
		}
		return err
	}

	if status == StockInsufficient {
		return PublishEvent("inventory-failed", nextEvent(orderEvent, "inventory-failed"))
	}

	// Stock is already reserved and safely recorded at this point. If either
	// publish below fails, don't force "inventory-failed" — that would tell
	// order-service to cancel while stock is genuinely held, with nothing to
	// release it (ReleaseStock only runs on payment-failed). Let it hit the
	// DLQ instead; ReserveStock's idempotency check means a replay picks up
	// exactly where this left off.
	if err := PublishEvent("inventory-checked", nextEvent(orderEvent, "inventory-checked")); err != nil {
		return err
	}
	return PublishEvent("payment-requested", nextEvent(orderEvent, "payment-requested"))
}

// nextEvent derives the downstream event_id deterministically from the
// source event_id + the step name, instead of a fresh random UUID. A retried
// order-requested delivery (see retry.go) re-runs this handler from scratch;
// without a deterministic ID here, each retry would mint a brand-new
// "payment-requested" event that payment-service's own idempotency check
// can't recognize as a duplicate — risking a second real charge.
func nextEvent(orderEvent OrderRequestedEvent, step string) OrderRequestedEvent {
	orderEvent.EventID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(orderEvent.EventID+":"+step)).String()
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
