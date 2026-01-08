package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaProducer is a helper for publishing events to Kafka
type KafkaProducer struct {
	brokers []string
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(brokers []string) *KafkaProducer {
	return &KafkaProducer{
		brokers: brokers,
	}
}

// PublishEvent publishes an event to Kafka topic
func (p *KafkaProducer) PublishEvent(ctx context.Context, topic string, key string, value interface{}) error {
	// Marshal value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create writer for this topic
	writer := &kafka.Writer{
		Addr:         kafka.TCP(p.brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false, // Synchronous to ensure message is sent
		MaxAttempts:  3,
		WriteTimeout: 5 * time.Second,
	}
	defer writer.Close()

	// Write message
	err = writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: data,
		Time:  time.Now(),
	})

	if err != nil {
		return fmt.Errorf("failed to publish event to topic %s: %w", topic, err)
	}

	return nil
}

// PaymentSuccessEvent represents a payment success event
type PaymentSuccessEvent struct {
	EventType string      `json:"event_type"`
	EventID   string      `json:"event_id"`
	Timestamp int64       `json:"timestamp"`
	Data      PaymentData `json:"data"`
}

// PaymentFailedEvent represents a payment failed event
type PaymentFailedEvent struct {
	EventType string      `json:"event_type"`
	EventID   string      `json:"event_id"`
	Timestamp int64       `json:"timestamp"`
	Data      PaymentData `json:"data"`
}

// PaymentData contains payment information
type PaymentData struct {
	PaymentID   int64   `json:"payment_id"`
	PaymentNo   string  `json:"payment_no"`
	OrderID     int64   `json:"order_id"`
	UserID      int64   `json:"user_id"`
	Amount      float64 `json:"amount"`
	PaymentType int     `json:"payment_type"`
	Status      int     `json:"status"`
	TradeNo     string  `json:"trade_no,omitempty"`
	Reason      string  `json:"reason,omitempty"` // For failed payments
}
