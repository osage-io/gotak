// Package kafkabus is a tiny, best-effort Kafka producer for the Communications
// tab. Chat messages are published write-only to a per-channel topic; failures
// never block chat (the broker is a side-channel for the demo). When no brokers
// are configured the producer is a no-op.
package kafkabus

import (
	"context"
	"strings"
	"time"

	"github.com/dfedick/gotak/pkg/logger"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
	logger *logger.Logger
}

// New returns a Producer for the comma-separated broker list, or nil if empty.
func New(brokers string, log *logger.Logger) *Producer {
	brokers = strings.TrimSpace(brokers)
	if brokers == "" {
		return nil
	}
	w := &kafka.Writer{
		Addr:                   kafka.TCP(strings.Split(brokers, ",")...),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true, // open the topic on first write
		BatchTimeout:           50 * time.Millisecond,
		RequiredAcks:           kafka.RequireOne,
	}
	log.Info().Str("brokers", brokers).Msg("Kafka producer enabled")
	return &Producer{writer: w, logger: log}
}

// PublishAsync fires a message at the topic without blocking the caller.
func (p *Producer) PublishAsync(topic, key string, value []byte) {
	if p == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := p.writer.WriteMessages(ctx, kafka.Message{
			Topic: topic,
			Key:   []byte(key),
			Value: value,
		})
		if err != nil {
			p.logger.Warn().Err(err).Str("topic", topic).Msg("Kafka publish failed (chat unaffected)")
		}
	}()
}

func (p *Producer) Close() error {
	if p == nil {
		return nil
	}
	return p.writer.Close()
}
