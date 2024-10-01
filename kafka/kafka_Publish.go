package ekafka

import (
	"context"
	eventbus "github.com/SyaibanAhmadRamadhan/event-bus"
)

func (b *broker) Publish(ctx context.Context, input PubInput) (output PubOutput, err error) {
	if input.Messages == nil || len(input.Messages) <= 0 {
		return
	}

	var ctxTracer context.Context
	if b.pubTracer != nil {
		ctxTracer = b.pubTracer.TracePubStart(ctx, &input.Messages[0])
	}

	err = b.kafkaWriter.WriteMessages(ctx, input.Messages...)
	if err != nil {
		err = eventbus.Error(err)
	}

	if b.pubTracer != nil {
		b.pubTracer.TracePubEnd(ctxTracer, output, err)
	}
	return
}
