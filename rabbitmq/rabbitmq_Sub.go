package erabbitmq

import (
	"context"
	eventbus "github.com/SyaibanAhmadRamadhan/event-bus"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (r *rabbitMQ) Subscribe(ctx context.Context, input SubInput) (output SubOutput, err error) {
	deliveries, err := r.ch.ConsumeWithContext(ctx, input.QueueName, input.ConsumerName, input.AutoAck, input.Exclusive, input.NoLocal, input.NoWait, input.Args)
	if err != nil {
		return output, eventbus.Error(err)
	}

	newDeliveries := make(chan amqp.Delivery)

	consume := func(delivery amqp.Delivery) {
		if r.cfg.subTracer != nil {
			r.cfg.subTracer.TraceSubStart(ctx, input, &delivery, r.ch)
		}

		newDeliveries <- delivery
		if input.AutoAck {
			defer func() {
				delivery.Ack(false)
			}()
		}
	}

	go func() {
		for delivery := range deliveries {
			consume(delivery)
		}
	}()

	output = SubOutput{
		Deliveries: newDeliveries,
	}

	return output, nil
}

type SubInput struct {
	QueueName    string
	ConsumerName string
	AutoAck      bool
	Exclusive    bool
	NoLocal      bool
	NoWait       bool
	Args         amqp.Table
}

type SubOutput struct {
	Deliveries chan amqp.Delivery
	D          <-chan func(delivery amqp.Delivery)
}
