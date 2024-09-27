package erabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type PubInput struct {
	ExchangeName string
	RoutingKey   string
	Mandatory    bool
	Immediate    bool
	Msg          amqp.Publishing

	DelayRetry time.Duration
	MaxRetry   int
}

type PubOutput struct {
	MessageID    string
	Confirmation *amqp.DeferredConfirmation
}
