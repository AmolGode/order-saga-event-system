package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	dbWriteDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_write_duration_seconds",
		Help:    "Time taken for a database write, labeled by operation.",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation"})

	// sagaCompletionDuration is the end-to-end time from order-requested being
	// produced to a terminal event (payment-success/payment-failed) landing
	// here, for the same order_id. This is the only service that sees every
	// hop of an order's journey, so it's the natural place to compute it.
	sagaCompletionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "saga_completion_duration_seconds",
		Help:    "Time from order-requested to a terminal event for the same order_id, labeled by outcome.",
		Buckets: prometheus.DefBuckets,
	}, []string{"outcome"})
)

func StartMetricsServer(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Println("metrics server error:", err)
		}
	}()
}
