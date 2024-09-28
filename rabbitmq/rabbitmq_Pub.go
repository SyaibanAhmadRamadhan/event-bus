package erabbitmq

import (
	"context"
	eventbus "github.com/SyaibanAhmadRamadhan/event-bus"
	"github.com/google/uuid"
	"time"
)

func (r *rabbitMQ) Publish(ctx context.Context, input PubInput) (output PubOutput, err error) {
	if input.Msg.MessageId == "" {
		input.Msg.MessageId = uuid.New().String()
	}

	if input.Msg.CorrelationId == "" {
		input.Msg.CorrelationId = uuid.New().String()
	}

	ctxOtel := r.cfg.pubTracer.TracePubStart(ctx, input)

	if r.isClosed {
		err = eventbus.Error(ErrProcessShutdownIsRunning)
		r.cfg.pubTracer.TracePubEnd(ctxOtel, output, err)
		return
	}

	r.wg.Add(1)
	defer r.wg.Done()

	output, err = r.retryPublish(ctx, ctxOtel, input)
	r.cfg.pubTracer.TracePubEnd(ctxOtel, output, err)
	return
}

func (r *rabbitMQ) retryPublish(ctx context.Context, ctxOtel context.Context, input PubInput) (output PubOutput, err error) {
	for attempts := 0; attempts < input.MaxRetry; attempts++ {
		err = r.publish(ctx, input, &output)
		if err == nil {
			return
		}

		if r.isClosed {
			return output, eventbus.Error(ErrProcessShutdownIsRunning)
		}

		select {
		case <-ctx.Done():
			return output, eventbus.Error(ctx.Err())
		default:
			r.cfg.pubTracer.RecordRetryPub(ctxOtel, attempts+1, err)
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
		return eventbus.Error(err)
	}

	*output = PubOutput{
		Confirmation: deferredConfirm,
		MessageID:    input.Msg.MessageId,
	}
	return nil
}
