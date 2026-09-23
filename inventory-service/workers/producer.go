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

func PublishEvent(topic string, key string, payload interface{}) error {
	value, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return PublishRaw(topic, key, value)
}

// PublishRaw republishes an already-serialized message as-is. key should be
// the order_id — keying by it keeps every event for the same order on the
// same partition, so Kafka's per-partition ordering guarantee actually
// applies across an order's whole saga instead of events landing on
// whichever partition PartitionAny happens to pick.
func PublishRaw(topic string, key string, value []byte) error {
	return producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          value,
	}, nil)
}
