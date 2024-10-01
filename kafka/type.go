package ekafka

import (
	"github.com/segmentio/kafka-go"
)

type PubInput struct {
	Messages []kafka.Message
}

type PubOutput struct{}

type SubInput struct {
	Config kafka.ReaderConfig
}

type SubOutput struct {
	Reader Reader
}
