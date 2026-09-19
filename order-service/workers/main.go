package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := LoadConfig()

	db, err := NewDB(buildConnString(cfg))
	if err != nil {
		log.Fatal("db connection failed:", err)
	}

	if err := InitProducer(cfg.KafkaBrokers); err != nil {
		log.Fatal("producer init failed:", err)
	}

	StartMetricsServer(":9100")
	go StartConsumer(cfg, db)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("shutting down gracefully")
}
