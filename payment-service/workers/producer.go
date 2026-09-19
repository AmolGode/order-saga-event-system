package main

import (
	"encoding/json"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var producer *kafka.Producer

func InitProducer(brokers string) error {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": brokers})
	if err != nil {
		return err
	}
	producer = p
	return nil
}

func PublishEvent(topic string, payload interface{}) error {
	value, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return PublishRaw(topic, value, nil)
}

// PublishRaw republishes an already-serialized message, optionally carrying
// headers (used for the retry-count header on retry/DLQ republishing).
func PublishRaw(topic string, value []byte, headers []kafka.Header) error {
	return producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          value,
		Headers:        headers,
	}, nil)
}
