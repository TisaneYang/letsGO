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

// OrderCreatedEvent represents an order created event
type OrderCreatedEvent struct {
	EventType string    `json:"event_type"`
	EventID   string    `json:"event_id"`
	Timestamp int64     `json:"timestamp"`
	Data      OrderData `json:"data"`
}

// OrderData contains order information
type OrderData struct {
	OrderID     int64       `json:"order_id"`
	OrderNo     string      `json:"order_no"`
	UserID      int64       `json:"user_id"`
	TotalAmount float64     `json:"total_amount"`
	Items       []OrderItem `json:"items"`
}

// OrderItem represents an item in the order
type OrderItem struct {
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
	Price     float64 `json:"price"`
}

// OrderCancelledEvent represents an order cancelled event
type OrderCancelledEvent struct {
	EventType string    `json:"event_type"`
	EventID   string    `json:"event_id"`
	Timestamp int64     `json:"timestamp"`
	Data      struct {
		OrderID int64       `json:"order_id"`
		OrderNo string      `json:"order_no"`
		Items   []OrderItem `json:"items"` // For stock restoration
	} `json:"data"`
}

// OrderStatusChangedEvent represents an order status change event
type OrderStatusChangedEvent struct {
	EventType string `json:"event_type"`
	EventID   string `json:"event_id"`
	Timestamp int64  `json:"timestamp"`
	Data      struct {
		OrderID   int64 `json:"order_id"`
		OrderNo   string `json:"order_no"`
		OldStatus int   `json:"old_status"`
		NewStatus int   `json:"new_status"`
	} `json:"data"`
}
