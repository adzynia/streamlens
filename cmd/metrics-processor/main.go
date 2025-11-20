package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"streamlens/internal/config"
	"streamlens/internal/kafka"
	"streamlens/internal/processor"
	"streamlens/internal/store"
	"syscall"
)

func main() {
	log.Println("Starting Metrics Processor...")

	// Load configuration
	cfg := config.Load()

	// Create Kafka consumer
	topics := []string{kafka.TopicLLMRequests, kafka.TopicLLMResponses}
	consumer, err := kafka.NewConsumer(cfg.KafkaBrokers, cfg.ConsumerGroup, topics)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()

	// Create Kafka producer for metrics output
	producer, err := kafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// Connect to Postgres
	metricsStore, err := store.NewMetricsStore(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer metricsStore.Close()

	// Create processor
	proc := processor.NewMetricsProcessor(consumer, producer, metricsStore)
	defer proc.Close()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Run processor in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := proc.Run(ctx); err != nil && err != context.Canceled {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-quit:
		log.Println("Received shutdown signal...")
		cancel()
	case err := <-errChan:
		log.Printf("Processor error: %v", err)
		cancel()
	}

	log.Println("Metrics processor exited")
}
