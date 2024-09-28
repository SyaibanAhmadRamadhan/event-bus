package erabbitmq

import (
	"github.com/guregu/null/v5"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type Config struct {
	reconnect          chan struct{}
	reconnectDelay     time.Duration
	maxRetryConnection null.Int

	confirmMode bool

	pubTracer   TracerPub
	reconTracer TracerReconnection
}

type Options func(cfg *Config)

func WithReconnect(reconnectDelay time.Duration, maxRetry null.Int) Options {
	return func(cfg *Config) {
		cfg.reconnect = make(chan struct{})
		cfg.reconnectDelay = reconnectDelay
		cfg.maxRetryConnection = maxRetry
	}
}

func WithConfirmMode(b bool) Options {
	return func(cfg *Config) {
		cfg.confirmMode = b
	}
}

func WithOtel(url string) Options {
	return func(cfg *Config) {
		uri, err := amqp.ParseURI(url)
		if err != nil {
			panic(err)
		}
		o := NewOtel(uri)
		cfg.pubTracer = o
		cfg.reconTracer = o
	}
}
