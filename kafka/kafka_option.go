package ekafka

import (
	"github.com/segmentio/kafka-go"
	"strings"
)

type Options func(cfg *broker)

func KafkaWriter(url, topic string) Options {
	return func(cfg *broker) {
		cfg.kafkaWriter = &kafka.Writer{
			Addr:     kafka.TCP(url),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		}
	}
}

func KafkaCustomWriter(k *kafka.Writer) Options {
	return func(cfg *broker) {
		cfg.kafkaWriter = k
	}
}

func KafkaReader(url, groupID, topic string) Options {
	return func(cfg *broker) {
		brokers := strings.Split(url, ",")
		cfg.kafkaReader = kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			GroupID:  groupID,
			Topic:    topic,
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		})
	}
}

func KafkaReaderCustom(c kafka.ReaderConfig) Options {
	return func(cfg *broker) {
		cfg.kafkaReader = kafka.NewReader(c)
	}
}

func WithOtel() Options {
	return func(cfg *broker) {
		o := NewOtel()
		cfg.pubTracer = o
		cfg.subTracer = o
		cfg.commitTracer = o
	}
}
