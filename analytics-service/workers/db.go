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
// not one INSERT per row. It COPYs into a temp staging table first, then
// merges with ON CONFLICT DO NOTHING: a plain COPY straight into
// analytics_analyticsevent is all-or-nothing, so one redelivered/duplicate
// event_id in the batch (event_id is the primary key) would fail the whole
// COPY and silently drop every other legitimate row batched alongside it.
func (db *DB) FlushBatch(ctx context.Context, rows []AnalyticsRow) error {
	if len(rows) == 0 {
		return nil
	}

	timer := prometheus.NewTimer(dbWriteDuration.WithLabelValues("flush_batch"))
	defer timer.ObserveDuration()

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // no-op if tx.Commit succeeded

	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE analytics_event_staging
		(LIKE analytics_analyticsevent INCLUDING DEFAULTS)
		ON COMMIT DROP
	`)
	if err != nil {
		return err
	}

	source := make([][]interface{}, len(rows))
	for i, r := range rows {
		source[i] = []interface{}{r.EventID, r.OrderID, r.Topic, r.Status, r.Timestamp, r.Payload}
	}

	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"analytics_event_staging"},
		[]string{"event_id", "order_id", "topic", "status", "timestamp", "payload"},
		pgx.CopyFromRows(source),
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO analytics_analyticsevent (event_id, order_id, topic, status, timestamp, payload)
		SELECT event_id, order_id, topic, status, timestamp, payload FROM analytics_event_staging
		ON CONFLICT (event_id) DO NOTHING
	`)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
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
