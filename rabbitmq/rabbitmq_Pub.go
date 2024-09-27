package rabbitmq

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	eventbus "go-event-bus"
	"log"
	"time"
)

func (r *rabbitMQ) Publish(ctx context.Context, input PubInput) (output PubOutput, err error) {
	select {
	case <-r.isShutdown:
		return output, eventbus.Error(ErrProcessShutdownIsRunning)
	default:
	}

	r.wg.Add(1)
	defer r.wg.Done()

	if input.Timeout == 0 {
		input.Timeout = time.Second * 10
	}
	ctx, cancel := context.WithTimeout(ctx, input.Timeout)
	defer cancel()

	if input.Msg.MessageId == "" {
		input.Msg.MessageId = uuid.New().String()
	}

	if input.Msg.CorrelationId == "" {
		input.Msg.CorrelationId = uuid.New().String()
	}

	return r.retryPublish(ctx, input)
}

func (r *rabbitMQ) retryPublish(ctx context.Context, input PubInput) (output PubOutput, err error) {
	for attempts := 0; attempts < input.MaxRetry; attempts++ {
		err = r.publish(ctx, input, &output)
		if err == nil {
			return
		}

		select {
		case <-ctx.Done():
			return output, eventbus.Error(ctx.Err())
		case <-r.isShutdown:
			return output, eventbus.Error(ErrProcessShutdownIsRunning)
		default:
			log.Printf("publish error: %v, attempting reconnection (%d/%d)", err, attempts+1, input.MaxRetry)
			r.signalReconnect()
			time.Sleep(input.DelayRetry)
		}
	}
	if err != nil {
		return output, eventbus.Error(err)
	}
	return
}

func (r *rabbitMQ) publish(_ context.Context, input PubInput, output *PubOutput) error {
	deferredConfirm, err := r.ch.PublishWithDeferredConfirm(input.ExchangeName, input.RoutingKey, input.Mandatory, input.Immediate, input.Msg)
	if err != nil {
		fmt.Println(err)
		return eventbus.Error(err)
	}

	*output = PubOutput{
		Confirmation: deferredConfirm,
		MessageID:    input.Msg.MessageId,
	}
	return nil
}
