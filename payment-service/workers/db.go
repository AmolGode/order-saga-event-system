package main

import (
	"context"
	"errors"

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

// ProcessPayment returns the outcome ("SUCCESS"/"FAILED") of this event_id,
// whether decided just now or already recorded from an earlier attempt.
// Callers must use this return value (not the `success` they passed in) to
// decide what to publish next — a retried delivery has to reproduce the
// exact same decision, not re-roll the simulated gateway charge.
func (db *DB) ProcessPayment(ctx context.Context, eventID, orderID string, amount float64, success bool) (string, error) {
	timer := prometheus.NewTimer(dbWriteDuration.WithLabelValues("process_payment"))
	defer timer.ObserveDuration()

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx) // no-op if tx.Commit succeeded

	var existingStatus string
	err = tx.QueryRow(ctx,
		"SELECT status FROM payments_processedevent WHERE event_id = $1",
		eventID,
	).Scan(&existingStatus)
	if err == nil {
		idempotencyHitsTotal.WithLabelValues("process_payment").Inc()
		return existingStatus, nil // already handled — reuse the recorded outcome, don't re-decide
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
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
		return "", err
	}

	_, err = tx.Exec(ctx,
		"INSERT INTO payments_processedevent (event_id, topic, processed_at, status) VALUES ($1, $2, NOW(), $3)",
		eventID, "payment-requested", status,
	)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return status, nil
}
