package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

type DB struct {
	pool *pgxpool.Pool
}

func NewDB(connString string) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, err
	}
	return &DB{pool: pool}, nil
}

type AnalyticsRow struct {
	EventID   string
	OrderID   string
	Topic     string
	Status    string
	Timestamp time.Time
	Payload   []byte
}

// FlushBatch inserts all buffered rows in a single bulk COPY — one round trip,
// not one INSERT per row.
func (db *DB) FlushBatch(ctx context.Context, rows []AnalyticsRow) error {
	if len(rows) == 0 {
		return nil
	}

	timer := prometheus.NewTimer(dbWriteDuration.WithLabelValues("flush_batch"))
	defer timer.ObserveDuration()

	source := make([][]interface{}, len(rows))
	for i, r := range rows {
		source[i] = []interface{}{r.EventID, r.OrderID, r.Topic, r.Status, r.Timestamp, r.Payload}
	}

	_, err := db.pool.CopyFrom(
		ctx,
		pgx.Identifier{"analytics_analyticsevent"},
		[]string{"event_id", "order_id", "topic", "status", "timestamp", "payload"},
		pgx.CopyFromRows(source),
	)
	return err
}

// EarliestEventTime returns when the given topic was first recorded for an
// order_id — used to compute saga_completion_duration_seconds when a
// terminal event (payment-success/payment-failed) comes in.
func (db *DB) EarliestEventTime(ctx context.Context, orderID, topic string) (time.Time, bool, error) {
	var ts *time.Time // MIN() over zero matching rows is NULL, not a zero-value time
	err := db.pool.QueryRow(ctx,
		"SELECT MIN(timestamp) FROM analytics_analyticsevent WHERE order_id = $1 AND topic = $2",
		orderID, topic,
	).Scan(&ts)
	if err != nil {
		return time.Time{}, false, err
	}
	if ts == nil {
		return time.Time{}, false, nil
	}
	return *ts, true, nil
}
