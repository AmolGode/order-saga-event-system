package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

var ErrInsufficientStock = errors.New("insufficient stock")

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

func (db *DB) ReserveStock(ctx context.Context, eventID, productID string, qty int) error {
	timer := prometheus.NewTimer(dbWriteDuration.WithLabelValues("reserve_stock"))
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
		idempotencyHitsTotal.WithLabelValues("reserve_stock").Inc()
		return nil // already handled — safe no-op
	}

	result, err := tx.Exec(ctx,
		"UPDATE inventory_inventory SET quantity_available = quantity_available - $1, quantity_reserved = quantity_reserved + $1 WHERE product_id = $2 AND quantity_available >= $1",
		qty, productID,
	)
	if err != nil {
		// Todo: Retry
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("%w for product %s", ErrInsufficientStock, productID)
	}

	_, err = tx.Exec(ctx,
		"INSERT INTO inventory_processedevent (event_id, topic, processed_at, status) VALUES ($1, $2, NOW(), $3)",
		eventID, "order-requested", "SUCCESS",
	)
	if err != nil {
		// Todo: Retry
		return err
	}

	return tx.Commit(ctx)
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
