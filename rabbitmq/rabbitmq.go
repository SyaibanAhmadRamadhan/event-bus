package erabbitmq

import (
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	eventbus "go-event-bus"
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
	config   *Config
}

func New(url string, opt ...Options) *rabbitMQ {
	r := &rabbitMQ{
		ch:       nil,
		conn:     nil,
		config:   &Config{},
		isClosed: false,
		wg:       sync.WaitGroup{},
	}

	for _, o := range opt {
		o(r.config)
	}

	err := r.connect(url)
	if err != nil {
		panic(err)
	}

	if r.config.reconnect != nil {
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
	close(r.config.reconnect)

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
		return errClosedRabbitmq
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

	if r.config.confirmMode {
		if err = ch.Confirm(false); err != nil {
			conn.Close()
			ch.Close()
			return fmt.Errorf("cannot support confirm mode: %v", err)
		}
	}

	r.closeConnection()
	r.closeChannel()

	r.ch = ch
	r.conn = conn

	return nil
}

func (r *rabbitMQ) signalReconnect() {
	if r.config.reconnect != nil {
		select {
		case r.config.reconnect <- struct{}{}:
		default:
			break
		}
	}
}

func (r *rabbitMQ) reconnectUrl(url string) {
	for {
		select {
		case _, ok := <-r.config.reconnect:
			if !ok {
				return
			}
			r.tryReconnect(url)
		}
	}
}

func (r *rabbitMQ) tryReconnect(url string) {
	if r.config.maxRetryConnection.Valid {
		for i := int64(1); i <= r.config.maxRetryConnection.Int64; i++ {
			log.Printf("Attempt %d to reconnect...", i)
			if err := r.connect(url); err != nil {
				if errors.Is(err, errClosedRabbitmq) {
					return
				}
				log.Printf("Reconnection attempt %d failed: %v", i, err)
				time.Sleep(r.config.reconnectDelay)
			} else {
				log.Println("Reconnection successful")
				return
			}
		}
		log.Println("Max reconnection attempts reached, giving up")
	} else {
		for attempt := int64(1); ; attempt++ {
			log.Printf("Attempt %d to reconnect...", attempt)
			if err := r.connect(url); err != nil {
				if errors.Is(err, errClosedRabbitmq) {
					return
				}
				log.Printf("Reconnection attempt %d failed: %v", attempt, err)
				time.Sleep(r.config.reconnectDelay)
			} else {
				log.Println("Reconnection successful")
				return
			}
		}
	}
}
