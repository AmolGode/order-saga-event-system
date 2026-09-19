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

func (db *DB) UpdateOrderStatus(orderID string, status string, reason string) error {
	timer := prometheus.NewTimer(dbWriteDuration.WithLabelValues("update_order_status"))
	defer timer.ObserveDuration()

	_, err := db.pool.Exec(context.Background(),
		"UPDATE orders_order SET status = $1, failed_reason = $2 WHERE id = $3",
		status, reason, orderID,
	)
	if err == nil {
		ordersProcessedTotal.WithLabelValues(status).Inc()
	}
	return err
}
