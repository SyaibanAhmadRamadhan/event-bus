package erabbitmq

import (
	"github.com/guregu/null/v5"
	"time"
)

type Config struct {
	reconnect          chan struct{}
	reconnectDelay     time.Duration
	maxRetryConnection null.Int

	confirmMode bool
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

//
//func WithConfirmMode() Options {
//	return func(cfg *Config) {
//		cfg.confirmMode = true
//		cfg.confirmsCh = make(chan struct {
//			deliveredConfirmation *amqp.DeferredConfirmation
//			message               amqp.Publishing
//		})
//		cfg.exitCh = make(chan struct{})
//		cfg.confirmsDoneCh = make(chan struct{})
//		cfg.publishOkCh = make(chan struct{}, 1)
//		cfg.maxShelterConfirmation = 8
//	}
//}
