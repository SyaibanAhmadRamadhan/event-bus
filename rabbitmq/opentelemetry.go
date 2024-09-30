package erabbitmq

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"strings"
	"sync"
	"time"
)

type opentelemetry struct {
	tracer      trace.Tracer
	propagators propagation.TextMapPropagator
	attrs       []attribute.KeyValue

	// When ack multiple, we need to end spans of every delivery before the tag,
	// so we keep a map of every span that haven't ended.
	spanMap map[uint64]trace.Span
	m       sync.Mutex
}

func nameWhenPublish(exchange string) string {
	if exchange == "" {
		exchange = "(default)"
	}
	return "publish " + exchange
}

func queueAnonymous(queue string) bool {
	return strings.HasPrefix(queue, "amq.gen-")
}

func nameWhenConsume(queue string) string {
	if queueAnonymous(queue) {
		queue = "(anonymous)"
	}
	return "process " + queue
}

func NewOtel(uri amqp.URI) *opentelemetry {
	tp := otel.GetTracerProvider()
	findOwnImportedVersion()
	return &opentelemetry{
		tracer:      tp.Tracer(TracerName, trace.WithInstrumentationVersion(InstrumentVersion)),
		propagators: otel.GetTextMapPropagator(),
		attrs: []attribute.KeyValue{
			semconv.ServiceName(amqpLibName),
			semconv.ServiceVersion(amqpLibVersion),
			semconv.MessagingSystemRabbitmq,
			semconv.NetworkProtocolName(uri.Scheme),
			semconv.NetworkProtocolVersion(netProtocolVer),
			semconv.NetworkTransportTCP,
			semconv.ServerAddress(uri.Host),
			semconv.ServerPort(uri.Port),
		},
		spanMap: make(map[uint64]trace.Span),
	}
}

func (r *opentelemetry) RecordRetryPub(ctx context.Context, attempt int, err error) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.Int("messaging.retry.attempt", attempt),
	}

	if err != nil {
		attrs = append(attrs, attribute.String("messaging.message.retry.error", err.Error()))
	}

	span.AddEvent(fmt.Sprintf("Retry Publish attempt %d", attempt), trace.WithAttributes(attrs...))
}

func (r *opentelemetry) TracePubStart(ctx context.Context, input PubInput) context.Context {
	attrs := []attribute.KeyValue{
		semconv.MessagingRabbitmqDestinationRoutingKey(input.RoutingKey),
		semconv.MessagingOperationTypePublish,
		semconv.MessagingOperationName("publish"),
		semconv.MessagingMessageBodySize(len(input.Msg.Body)),
		semconv.MessagingMessageConversationID(input.Msg.CorrelationId),
		semconv.MessagingMessageID(input.Msg.MessageId),
		semconv.MessagingDestinationPublishAnonymous(input.ExchangeName == ""),
		semconv.MessagingDestinationPublishName(input.ExchangeName),
		attribute.Int("messaging.message.max-retry", input.MaxRetry),
		attribute.String("messaging.message.delay-retry", input.DelayRetry.String()),
	}
	if input.Msg.CorrelationId != "" {
		attrs = append(attrs, semconv.MessagingMessageConversationID(input.Msg.CorrelationId))
	}
	if input.Msg.MessageId != "" {
		attrs = append(attrs, semconv.MessagingMessageID(input.Msg.MessageId))
	}
	attrs = append(attrs, r.attrs...)
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindConsumer),
	}

	ctx, _ = r.tracer.Start(ctx, nameWhenPublish(input.ExchangeName), opts...)
	return ctx
}

func (r *opentelemetry) TracePubEnd(ctx context.Context, input PubOutput, err error) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	span.End()
}

func (r *opentelemetry) TraceReConnStart(ctx context.Context, reconDelay time.Duration, attempt *int64) context.Context {
	attrs := []attribute.KeyValue{
		attribute.String("messaging.reconnection.delay", reconDelay.String()),
	}
	if attempt != nil {
		attrs = append(attrs, attribute.Int64("messaging.reconnect.attempts", *attempt))
	} else {
		attrs = append(attrs, attribute.String("messaging.reconnect.attempts", "unlimited"))
	}

	attrs = append(attrs, r.attrs...)
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindConsumer),
	}

	ctx, _ = r.tracer.Start(ctx, "rabbitmq-reconnection", opts...)
	return ctx
}

func (r *opentelemetry) TraceReConnEnd(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(
			attribute.String("messaging.reconnect.status", "failed"),
		)
	} else {
		span.SetStatus(codes.Ok, "Reconnection attempt successful")
		span.SetAttributes(
			attribute.String("messaging.reconnect.status", "successful"),
		)
	}

	span.End()
}

func (r *opentelemetry) RecordRetryReConn(ctx context.Context, attempt int64, err error) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.Int64("messaging.reconnect.attempt", attempt),
	}

	if err != nil {
		attrs = append(attrs, attribute.String(fmt.Sprintf("messaging.reconnect.attempt-%d.error", attempt), err.Error()))
	}

	span.AddEvent(fmt.Sprintf("Reconnect attempt %d", attempt), trace.WithAttributes(attrs...))
}

func (r *opentelemetry) TraceSubStart(ctx context.Context, input SubInput, msg *amqp.Delivery, ch *amqp.Channel) {
	attrs := []attribute.KeyValue{
		semconv.MessagingOperationTypeDeliver,
		semconv.MessagingOperationName("process"),
		semconv.MessagingDestinationAnonymous(queueAnonymous(input.QueueName)),
		semconv.MessagingDestinationName(input.QueueName),
		semconv.MessagingDestinationPublishAnonymous(msg.Exchange == ""),
		semconv.MessagingDestinationPublishName(msg.Exchange),
		// todo messaging.client.id
		semconv.MessagingRabbitmqDestinationRoutingKey(msg.RoutingKey),
	}
	if msg.MessageCount != 0 {
		attrs = append(attrs, semconv.MessagingBatchMessageCount(int(msg.MessageCount)))
	}
	if msg.CorrelationId != "" {
		attrs = append(attrs, semconv.MessagingMessageConversationID(msg.CorrelationId))
	}
	if msg.MessageId != "" {
		attrs = append(attrs, semconv.MessagingMessageID(msg.MessageId))
	}
	if msg.DeliveryTag != 0 {
		//nolint:gosec // overflow here is relatively safe and unlikely to happen
		attrs = append(attrs, semconv.MessagingRabbitmqMessageDeliveryTag(int(msg.DeliveryTag)))
	}
	attrs = append(attrs, r.attrs...)
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindConsumer),
	}

	_, span := r.tracer.Start(ctx,
		nameWhenConsume(input.QueueName), opts...)
	msg.Acknowledger = &otelAck{
		otel:  r,
		acker: ch,
		span:  span,
	}

	r.m.Lock()
	defer r.m.Unlock()
	r.spanMap[msg.DeliveryTag] = span
}

func (r *opentelemetry) EndAllSpan() {
	r.m.Lock()
	defer r.m.Unlock()

	for k, span := range r.spanMap {
		span.End()
		delete(r.spanMap, k)
	}
}
