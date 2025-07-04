package kafka

import (
    "context"
    "log"
    "time"

    "github.com/segmentio/kafka-go"
)

// Consumer represents a Kafka consumer.
type Consumer struct {
    reader *kafka.Reader
}

// NewConsumer creates a new Kafka consumer.
func NewConsumer(brokers []string, topic, groupID string) *Consumer {
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers:        brokers,
        Topic:          topic,
        GroupID:        groupID,
        MinBytes:       10e3, // 10KB
        MaxBytes:       10e6, // 10MB
        MaxAttempts:    3,    // Retry reading a message up to 3 times
        Dialer: &kafka.Dialer{
            Timeout: 10 * time.Second,
        },
        // Start reading from the beginning if no offset is found for the group
        StartOffset: kafka.FirstOffset,
    })
    log.Printf("Kafka consumer initialized for topic '%s', group '%s' on brokers %v", topic, groupID, brokers)
    return &Consumer{reader: reader}
}

// ConsumeMessages continuously reads messages from Kafka and processes them using the provided handler func.
func (c *Consumer) ConsumeMessages(ctx context.Context, handler func(message kafka.Message) error) {
    for {
        select {
        case <-ctx.Done():
            log.Println("Consumer context cancelled, stopping message consumption.")
            return
        default:
            m, err := c.reader.FetchMessage(ctx) // Fetch one message
            if err != nil {
                // This error indicates a problem with fetching, not processing a message
                if err == context.Canceled {
                    log.Println("Context cancelled during FetchMessage, exiting consumer loop.")
                    return // Exit gracefully
                }
                log.Printf("ERROR Consumer: Failed to fetch message: %v", err)
                time.Sleep(time.Second) // Wait a bit before retrying fetch
                continue
            }

            // Process the message
            processingErr := handler(m)
            if processingErr != nil {
                log.Printf("ERROR Consumer: Failed to process message (will not commit offset): %v", processingErr)
                // If handler returns an error, we don't commit the offset, so Kafka will re-deliver.
                // Depending on retry logic, this might go to a dead-letter queue eventually.
            } else {
                // Commit the offset if the message was processed successfully
                // This tells Kafka we've successfully handled this message.
                commitCtx, commitCancel := context.WithTimeout(context.Background(), 3*time.Second)
                err = c.reader.CommitMessages(commitCtx, m)
                commitCancel()
                if err != nil {
                    log.Printf("ERROR Consumer: Failed to commit message offset for topic %s partition %d offset %d: %v", m.Topic, m.Partition, m.Offset, err)
                    // If commit fails, the message will be re-delivered on restart, which is safe.
                }
            }
        }
    }
}

// Close closes the Kafka consumer connection.
func (c *Consumer) Close() error {
    log.Println("Closing Kafka consumer...")
    return c.reader.Close()
}