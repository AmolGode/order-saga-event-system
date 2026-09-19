package main

import (
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
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
		"payment-failed", // compensation: release stock reserved earlier for this order
	}, nil)

	for {
		msg, err := consumer.ReadMessage(-1)
		if err != nil {
			log.Println("read error:", err)
			continue
		}

		err = HandleMessage(db, *msg.TopicPartition.Topic, msg.Value)

		if err != nil {
			log.Println("inventory service handle error:", err)
			if pubErr := handleFailure(msg); pubErr != nil {
				log.Println("retry/dlq publish error:", pubErr)
				continue // don't commit — will be redelivered on next restart
			}
			consumer.CommitMessage(msg) // original done — retry/DLQ copy takes over
			continue
		}

		consumer.CommitMessage(msg)
	}
}
