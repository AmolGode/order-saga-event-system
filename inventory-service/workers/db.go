package main

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	StockReserved     = "RESERVED"
	StockInsufficient = "INSUFFICIENT"
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

// ReserveStock returns the outcome ("RESERVED"/"INSUFFICIENT") of this
// event_id, whether decided just now or already recorded from an earlier
// attempt. Callers must use this return value (not re-derive their own)
// to decide what to publish next — a retried delivery has to reproduce the
// exact same decision, not re-roll it.
func (db *DB) ReserveStock(ctx context.Context, eventID, productID string, qty int) (string, error) {
	timer := prometheus.NewTimer(dbWriteDuration.WithLabelValues("reserve_stock"))
	defer timer.ObserveDuration()

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx) // no-op if tx.Commit succeeded

	var existingStatus string
	err = tx.QueryRow(ctx,
		"SELECT status FROM inventory_processedevent WHERE event_id = $1",
		eventID,
	).Scan(&existingStatus)
	if err == nil {
		idempotencyHitsTotal.WithLabelValues("reserve_stock").Inc()
		return existingStatus, nil // already handled — reuse the recorded outcome, don't re-decide
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	status := StockReserved
	result, err := tx.Exec(ctx,
		"UPDATE inventory_inventory SET quantity_available = quantity_available - $1, quantity_reserved = quantity_reserved + $1 WHERE product_id = $2 AND quantity_available >= $1",
		qty, productID,
	)
	if err != nil {
		return "", err
	}
	if result.RowsAffected() == 0 {
		status = StockInsufficient
	}

	_, err = tx.Exec(ctx,
		"INSERT INTO inventory_processedevent (event_id, topic, processed_at, status) VALUES ($1, $2, NOW(), $3)",
		eventID, "order-requested", status,
	)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return status, nil
}

func (db *DB) ReleaseStock(ctx context.Context, eventID, productID string, qty int) error {
	timer := prometheus.NewTimer(dbWriteDuration.WithLabelValues("release_stock"))
	defer timer.ObserveDuration()

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // no-op if tx.Commit succeeded

	var exists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM inventory_processedevent WHERE event_id = $1)",
		eventID,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		idempotencyHitsTotal.WithLabelValues("release_stock").Inc()
		return nil // already handled — safe no-op
	}

	_, err = tx.Exec(ctx,
		"UPDATE inventory_inventory SET quantity_available = quantity_available + $1, quantity_reserved = quantity_reserved - $1 WHERE product_id = $2",
		qty, productID,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		"INSERT INTO inventory_processedevent (event_id, topic, processed_at, status) VALUES ($1, $2, NOW(), $3)",
		eventID, "payment-failed", "SUCCESS",
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
