package main

import (
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const maxRetries = 3
const retryCountHeader = "retry-count"

func retryCount(msg *kafka.Message) int {
	for _, h := range msg.Headers {
		if h.Key == retryCountHeader {
			n, _ := strconv.Atoi(string(h.Value))
			return n
		}
	}
	return 0
}

// handleFailure decides whether to retry (republish the same message with an
// incremented retry-count header) or give up and route it to the topic's
// dead-letter queue after maxRetries attempts.
func handleFailure(msg *kafka.Message) error {
	attempt := retryCount(msg) + 1
	topic := *msg.TopicPartition.Topic

	if attempt >= maxRetries {
		return PublishRaw(topic+"-dlq", msg.Value, nil)
	}

	return PublishRaw(topic, msg.Value, []kafka.Header{
		{Key: retryCountHeader, Value: []byte(strconv.Itoa(attempt))},
	})
}
