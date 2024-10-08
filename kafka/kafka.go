package ekafka

import (
	"errors"
	eventbus "github.com/SyaibanAhmadRamadhan/event-bus"
	"github.com/segmentio/kafka-go"
	"go.uber.org/mock/gomock"
)

type KafkaPubSub = eventbus.PubSub[PubInput, PubOutput, SubInput, SubOutput]

func NewMockKafkaPubSub(mock *gomock.Controller) *eventbus.MockPubSub[PubInput, PubOutput, SubInput, SubOutput] {
	return eventbus.NewMockPubSub[PubInput, PubOutput, SubInput, SubOutput](mock)
}

var ErrProcessShutdownIsRunning = errors.New("process shutdown is running")
var errClosed = errors.New("kafka closed")
var ErrJsonUnmarshal = errors.New("json unmarshal error")

type broker struct {
	kafkaWriter  *kafka.Writer
	pubTracer    TracerPub
	subTracer    TracerSub
	commitTracer TracerCommitMessage
}

func New(opts ...Options) *broker {
	b := &broker{}
	for _, option := range opts {
		option(b)
	}

	return b
}

func (b *broker) Close() {
	b.kafkaWriter.Close()
}
