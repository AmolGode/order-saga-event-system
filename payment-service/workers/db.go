package main

import (
	"context"

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

func (db *DB) ProcessPayment(ctx context.Context, eventID, orderID string, amount float64, success bool) error {
	timer := prometheus.NewTimer(dbWriteDuration.WithLabelValues("process_payment"))
	defer timer.ObserveDuration()

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // no-op if tx.Commit succeeded

	var exists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM payments_processedevent WHERE event_id = $1)",
		eventID,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		idempotencyHitsTotal.WithLabelValues("process_payment").Inc()
		return nil // already handled — safe no-op
	}

	status := "FAILED"
	if success {
		status = "SUCCESS"
	}

	_, err = tx.Exec(ctx,
		"INSERT INTO payments_payment (order_id, amount, status, created_at) VALUES ($1, $2, $3, NOW())",
		orderID, amount, status,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		"INSERT INTO payments_processedevent (event_id, topic, processed_at, status) VALUES ($1, $2, NOW(), $3)",
		eventID, "payment-requested", status,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
