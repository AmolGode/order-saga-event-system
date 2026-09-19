package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ordersProcessedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "orders_processed_total",
		Help: "Orders whose status was updated, labeled by the resulting status.",
	}, []string{"status"})

	dbWriteDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_write_duration_seconds",
		Help:    "Time taken for a database write, labeled by operation.",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation"})
)

func StartMetricsServer(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Println("metrics server error:", err)
		}
	}()
}
