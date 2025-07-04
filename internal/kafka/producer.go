package kafka

import (
    "context"
    "log"
    "time"

    "github.com/segmentio/kafka-go"
)

// Producer represents a Kafka producer.
type Producer struct {
    writer *kafka.Writer
}

// NewProducer creates a new Kafka producer.
func NewProducer(brokers []string, topic string) *Producer {
    writer := &kafka.Writer{
        Addr:     kafka.TCP(brokers...),
        Topic:    topic,
        Balancer: &kafka.LeastBytes{}, // Distribute messages among partitions
        // Required for local development to ensure topics are created automatically
        // This is generally not recommended for production.
        AllowAutoTopicCreation: true,
        // Async writes for performance
        Async: true,
        Completion: func(messages []kafka.Message, err error) {
            if err != nil {
                // For debugging, you might log errors, but in production, handle retries or dead-letter queues.
                log.Printf("ERROR Kafka message completion: %v", err)
            }
        },
    }
    log.Printf("Kafka producer initialized for topic '%s' on brokers %v", topic, brokers)
    return &Producer{writer: writer}
}

// PublishMessage sends a message to Kafka.
func (p *Producer) PublishMessage(ctx context.Context, key, value []byte) error {
    msg := kafka.Message{
        Key:   key,
        Value: value,
        Time:  time.Now(),
    }
    return p.writer.WriteMessages(ctx, msg)
}

// Close closes the Kafka producer connection.
func (p *Producer) Close() error {
    log.Println("Closing Kafka producer...")
    return p.writer.Close()
}