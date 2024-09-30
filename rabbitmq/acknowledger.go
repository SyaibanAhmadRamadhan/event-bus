package erabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type otelAck struct {
	otel  *opentelemetry
	acker amqp.Acknowledger // The real otelAck is amqp091.Channel
	span  trace.Span
}

func (a *otelAck) Ack(tag uint64, multiple bool) error {
	err := a.acker.Ack(tag, multiple)
	if multiple {
		a.endMultiple(tag, codes.Ok, "ack", err)
	} else {
		a.endOne(tag, codes.Ok, "ack", err)
	}
	return err
}

func (a *otelAck) Nack(tag uint64, multiple, requeue bool) error {
	err := a.acker.Nack(tag, multiple, requeue)
	if multiple {
		a.endMultiple(tag, codes.Error, "nack", err)
	} else {
		a.endOne(tag, codes.Error, "nack", err)
	}
	return err
}

func (a *otelAck) Reject(tag uint64, requeue bool) error {
	err := a.acker.Reject(tag, requeue)
	a.endOne(tag, codes.Error, "reject", err)
	return err
}

func (a *otelAck) endMultiple(lastTag uint64, code codes.Code, desc string, err error) {
	a.otel.m.Lock()
	defer a.otel.m.Unlock()

	for tag, span := range a.otel.spanMap {
		span.SetAttributes(semconv.MessagingOperationName(desc))
		if tag <= lastTag {
			if err != nil {
				span.RecordError(err)
			}
			span.SetStatus(code, desc)
			span.End()
			delete(a.otel.spanMap, tag)
		}
	}
}

func (a *otelAck) endOne(tag uint64, code codes.Code, desc string, err error) {
	a.otel.m.Lock()
	defer a.otel.m.Unlock()

	a.span.SetAttributes(semconv.MessagingOperationName(desc))
	if err != nil {
		a.span.RecordError(err)
	}
	a.span.SetStatus(code, desc)
	a.span.End()
	delete(a.otel.spanMap, tag)
}
