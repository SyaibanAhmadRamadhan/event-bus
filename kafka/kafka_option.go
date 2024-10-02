package ekafka

import (
	"github.com/segmentio/kafka-go"
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

func WithOtel() Options {
	return func(cfg *broker) {
		o := NewOtel()
		cfg.pubTracer = o
		cfg.subTracer = o
		cfg.commitTracer = o
	}
}
