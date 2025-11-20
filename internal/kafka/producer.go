package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	TopicLLMRequests  = "llm.requests"
	TopicLLMResponses = "llm.responses"
	TopicLLMMetrics   = "llm.metrics"
)

// Producer wraps a franz-go client for producing messages
type Producer struct {
	client *kgo.Client
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers []string) (*Producer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		return nil, err
	}

	return &Producer{client: client}, nil
}

// ProduceJSON produces a JSON-encoded message to the specified topic with a key
func (p *Producer) ProduceJSON(ctx context.Context, topic, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	record := &kgo.Record{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
	}

	// Synchronous produce with context
	result := p.client.ProduceSync(ctx, record)
	if err := result.FirstErr(); err != nil {
		return err
	}

	return nil
}

// Close shuts down the producer gracefully
func (p *Producer) Close() {
	p.client.Close()
	log.Println("Kafka producer closed")
}
