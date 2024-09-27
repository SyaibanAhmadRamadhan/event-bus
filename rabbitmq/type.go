package rabbitmq

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

	// if you use channel confirm mode
	Timeout    time.Duration // Timeout for waiting confirmation successfully sending from rabbitmq
	DelayRetry time.Duration
	MaxRetry   int
}

type PubOutput struct {
	MessageID    string
	Confirmation *amqp.DeferredConfirmation
}
