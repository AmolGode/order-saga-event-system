package main

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"
)

type incomingEvent struct {
	EventID string `json:"event_id"`
	OrderID string `json:"order_id"`
}

func buildAnalyticsRow(topic string, value []byte) (AnalyticsRow, error) {
	var e incomingEvent
	if err := json.Unmarshal(value, &e); err != nil {
		return AnalyticsRow{}, err
	}

	return AnalyticsRow{
		EventID:   e.EventID,
		OrderID:   e.OrderID,
		Topic:     topic,
		Status:    deriveStatus(topic),
		Timestamp: time.Now(),
		Payload:   value,
	}, nil
}

// recordSagaCompletion looks up when this order's order-requested event was
// ingested and observes the elapsed time as saga_completion_duration_seconds.
// Note: this measures ingestion-to-ingestion time (when analytics saw each
// event), not true production-to-production time — the events themselves
// carry no timestamp of their own, so this is the closest proxy available.
//
// pendingBatch is checked first — a fast order can have its order-requested
// still sitting unflushed in the current in-memory batch when the terminal
// event arrives moments later, so a DB-only lookup would miss exactly the
// fastest orders.
func recordSagaCompletion(db *DB, row AnalyticsRow, pendingBatch []AnalyticsRow) {
	for _, r := range pendingBatch {
		if r.OrderID == row.OrderID && r.Topic == "order-requested" {
			observeSagaCompletion(row, r.Timestamp)
			return
		}
	}

	startedAt, found, err := db.EarliestEventTime(context.Background(), row.OrderID, "order-requested")
	if err != nil {
		log.Println("saga completion lookup error:", err)
		return
	}
	if !found {
		return // order-requested for this order_id hasn't been seen yet
	}
	observeSagaCompletion(row, startedAt)
}

func observeSagaCompletion(row AnalyticsRow, startedAt time.Time) {
	outcome := "success"
	if row.Topic == "payment-failed" {
		outcome = "failed"
	}
	sagaCompletionDuration.WithLabelValues(outcome).Observe(row.Timestamp.Sub(startedAt).Seconds())
}

// deriveStatus buckets a topic into a coarse outcome — the events themselves
// carry no explicit status field, so this is inferred from the topic name.
func deriveStatus(topic string) string {
	switch {
	case strings.HasSuffix(topic, "-dlq"):
		return "DEAD_LETTER"
	case topic == "order-requested" || topic == "payment-requested":
		return "REQUESTED"
	case topic == "payment-success":
		return "SUCCESS"
	case topic == "inventory-failed" || topic == "payment-failed":
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}
