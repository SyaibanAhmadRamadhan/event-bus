package erabbitmq

import (
	"context"
	"errors"
	"fmt"
	eventbus "github.com/SyaibanAhmadRamadhan/event-bus"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/mock/gomock"
	"log"
	"sync"
	"time"
)

type RabbitMQPubSub = eventbus.PubSub[PubInput, PubOutput, SubInput, SubOutput]

func NewMockRabbitMQPubSub(mock *gomock.Controller) *eventbus.MockPubSub[PubInput, PubOutput, SubInput, SubOutput] {
	return eventbus.NewMockPubSub[PubInput, PubOutput, SubInput, SubOutput](mock)
}

var ErrProcessShutdownIsRunning = errors.New("process shutdown is running")
var errClosedRabbitmq = errors.New("rabbitmq closed")

type rabbitMQ struct {
	ch   *amqp.Channel
	conn *amqp.Connection

	isClosed bool
	mu       sync.Mutex
	wg       sync.WaitGroup
	cfg      *Config
}

func New(url string, opt ...Options) *rabbitMQ {
	r := &rabbitMQ{
		cfg: &Config{},
	}

	for _, o := range opt {
		o(r.cfg)
	}

	if r.cfg.pubTracer == nil || r.cfg.reconTracer == nil {
		o := WithOtel(url)
		o(r.cfg)
	}

	err := r.connect(url)
	if err != nil {
		panic(err)
	}

	if r.cfg.reconnect != nil {
		go r.reconnectUrl(url)
	}
	return r
}

// Close shuts down the RabbitMQ connection gracefully
func (r *rabbitMQ) Close() {
	if r.isClosed {
		return
	}
	r.mu.Lock()
	r.isClosed = true
	r.mu.Unlock()

	done := make(chan struct{})
	go r.waitForShutdown(done)

	<-done
	close(done)
	close(r.cfg.reconnect)

	if err := r.closeChannel(); err != nil {
		log.Printf("Error closing channel: %v", err)
	}

	if err := r.closeConnection(); err != nil {
		log.Printf("Error closing connection: %v", err)
	}
}

func (r *rabbitMQ) waitForShutdown(done chan struct{}) {
	r.wg.Wait()
	done <- struct{}{}
}

func (r *rabbitMQ) closeChannel() error {
	if r.ch != nil {
		return r.ch.Close()
	}
	return nil
}

func (r *rabbitMQ) closeConnection() error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// connect establishes a new RabbitMQ connection
func (r *rabbitMQ) connect(url string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.isClosed {
		return eventbus.Error(errClosedRabbitmq)
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("cannot (re)dial: %v: %q", err, url)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("cannot create channel: %v", err)
	}

	if r.cfg.confirmMode {
		if err = ch.Confirm(false); err != nil {
			conn.Close()
			ch.Close()
			return fmt.Errorf("cannot support confirm mode: %v", err)
		}
	}

	r.ch = ch
	r.conn = conn

	return nil
}

func (r *rabbitMQ) signalReconnect() {
	if r.cfg.reconnect != nil {
		select {
		case r.cfg.reconnect <- struct{}{}:
		default:
			break
		}
	}
}

func (r *rabbitMQ) reconnectUrl(url string) {
	for {
		select {
		case _, ok := <-r.cfg.reconnect:
			if !ok {
				return
			}
			ctxOtel := r.cfg.reconTracer.TraceReConnStart(context.Background(), r.cfg.reconnectDelay,
				r.cfg.maxRetryConnection.Ptr(),
			)
			r.tryReconnect(ctxOtel, url)
		}
	}
}

func (r *rabbitMQ) tryReconnect(ctx context.Context, url string) {
	r.closeConnection()
	r.closeChannel()
	if r.cfg.maxRetryConnection.Valid {
		for i := int64(1); i <= r.cfg.maxRetryConnection.Int64; i++ {
			if err := r.connect(url); err != nil {
				if errors.Is(err, errClosedRabbitmq) {
					r.cfg.reconTracer.TraceReConnEnd(ctx, err)
					return
				}
				r.cfg.reconTracer.RecordRetryReConn(ctx, i, err)
				time.Sleep(r.cfg.reconnectDelay)
			} else {
				r.cfg.reconTracer.TraceReConnEnd(ctx, nil)
				return
			}
		}
		r.cfg.reconTracer.TraceReConnEnd(ctx, errors.New("max reconnection attempts reached, giving up"))
	} else {
		for attempt := int64(1); ; attempt++ {
			if err := r.connect(url); err != nil {
				if errors.Is(err, errClosedRabbitmq) {
					r.cfg.reconTracer.TraceReConnEnd(ctx, err)
					return
				}
				r.cfg.reconTracer.RecordRetryReConn(ctx, attempt, err)
				time.Sleep(r.cfg.reconnectDelay)
			} else {
				r.cfg.reconTracer.TraceReConnEnd(ctx, nil)
				return
			}
		}
	}
}
