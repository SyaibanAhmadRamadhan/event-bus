package ekafka

import (
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/propagation"
)

type textMapCarrier struct {
	msg *kafka.Message
}

var _ propagation.TextMapCarrier = (*textMapCarrier)(nil)

func NewMsgCarrier(msg *kafka.Message) *textMapCarrier {
	return &textMapCarrier{msg}
}

// Get returns the value associated with the passed key.
func (c *textMapCarrier) Get(key string) string {
	for _, h := range c.msg.Headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

// Set stores the key-value pair.
func (c *textMapCarrier) Set(key, value string) {
	// Ensure the uniqueness of the key.
	for i := len(c.msg.Headers) - 1; i >= 0; i-- {
		if c.msg.Headers[i].Key == key {
			c.msg.Headers = append(c.msg.Headers[:i], c.msg.Headers[i+1:]...)
		}
	}
	c.msg.Headers = append(c.msg.Headers, kafka.Header{
		Key:   key,
		Value: []byte(value),
	})
}

// Keys lists the keys stored in this carrier.
func (c *textMapCarrier) Keys() []string {
	out := make([]string, len(c.msg.Headers))
	for i, h := range c.msg.Headers {
		out[i] = h.Key
	}
	return out
}
