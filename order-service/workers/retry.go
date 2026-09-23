package main

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// handleFailure sends a message straight to its topic's dead-letter queue
// on any handler failure — no in-place retry loop. Kafka already redelivers
// on its own if the consumer never gets to commit (e.g. a crash mid-handler),
// so a retry-count header on top of that was extra bookkeeping, not extra
// safety.
func handleFailure(msg *kafka.Message) error {
	topic := *msg.TopicPartition.Topic
	return PublishRaw(topic+"-dlq", string(msg.Key), msg.Value)
}
