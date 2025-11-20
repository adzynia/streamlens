package kafka

import (
	"context"
	"log"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Consumer wraps a franz-go client for consuming messages
type Consumer struct {
	client *kgo.Client
}

// NewConsumer creates a new Kafka consumer with a consumer group
func NewConsumer(brokers []string, group string, topics []string) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topics...),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()), // Start from beginning for new consumers
	)
	if err != nil {
		return nil, err
	}

	return &Consumer{client: client}, nil
}

// Poll fetches records from Kafka
func (c *Consumer) Poll(ctx context.Context) kgo.Fetches {
	return c.client.PollFetches(ctx)
}

// CommitRecords commits the offsets for consumed records
func (c *Consumer) CommitRecords(ctx context.Context, records ...*kgo.Record) error {
	return c.client.CommitRecords(ctx, records...)
}

// Close shuts down the consumer gracefully
func (c *Consumer) Close() {
	c.client.Close()
	log.Println("Kafka consumer closed")
}
