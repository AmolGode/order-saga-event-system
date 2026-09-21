package main

import (
	"context"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	batchSize     = 500
	flushInterval = 500 * time.Millisecond
)

func StartConsumer(cfg Config, db *DB) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  cfg.KafkaBrokers,
		"group.id":           cfg.GroupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	})
	if err != nil {
		log.Fatal(err)
	}

	consumer.SubscribeTopics([]string{
		"order-requested",
		"inventory-checked",
		"payment-requested",
		"inventory-failed",
		"payment-success",
		"payment-failed",
		// dead-letter topics — a message lands here on its first handler failure
		"order-requested-dlq",
		"payment-requested-dlq",
		"payment-failed-dlq",
		"payment-success-dlq",
		"inventory-failed-dlq",
	}, nil)

	var rows []AnalyticsRow
	var messages []*kafka.Message

	flush := func() {
		if len(rows) == 0 {
			return
		}
		if err := db.FlushBatch(context.Background(), rows); err != nil {
			log.Println("flush error:", err)
			rows = nil
			messages = nil
			return
		}
		for _, m := range messages {
			consumer.CommitMessage(m)
		}
		rows = nil
		messages = nil
	}

	for {
		msg, err := consumer.ReadMessage(flushInterval)
		if err != nil {
			if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() == kafka.ErrTimedOut {
				flush() // no new message within the interval — time-based trigger
				continue
			}
			log.Println("read error:", err)
			continue
		}

		row, err := buildAnalyticsRow(*msg.TopicPartition.Topic, msg.Value)
		if err != nil {
			log.Println("parse error:", err)
			continue
		}

		if row.Topic == "payment-success" || row.Topic == "payment-failed" {
			recordSagaCompletion(db, row, rows)
		}

		rows = append(rows, row)
		messages = append(messages, msg)

		if len(rows) >= batchSize {
			flush()
		}
	}
}
