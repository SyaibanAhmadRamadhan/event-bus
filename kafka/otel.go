package ekafka

import (
	"context"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"strconv"
	"sync"
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

func NewOtel() *opentelemetry {
	tp := otel.GetTracerProvider()
	findOwnImportedVersion()
	return &opentelemetry{
		tracer:      tp.Tracer(TracerName, trace.WithInstrumentationVersion(InstrumentVersion)),
		propagators: otel.GetTextMapPropagator(),
		attrs: []attribute.KeyValue{
			semconv.ServiceName(kafkaLibName),
			semconv.ServiceVersion(kafkaLibVersion),
			semconv.MessagingSystemKafka,
			semconv.NetworkProtocolVersion(netProtocolVer),
			semconv.NetworkTransportTCP,
		},
		spanMap: make(map[uint64]trace.Span),
	}
}

func (r *opentelemetry) TracePubStart(ctx context.Context, msg *kafka.Message) context.Context {
	carrier := NewMsgCarrier(msg)
	ctx = r.propagators.Extract(ctx, carrier)

	attrs := []attribute.KeyValue{
		semconv.MessagingKafkaMessageKey(string(msg.Key)),
		semconv.MessagingDestinationName(msg.Topic),
		semconv.MessagingOperationTypePublish,
		semconv.MessagingOperationName("send"),
		semconv.MessagingMessageBodySize(len(msg.Value)),
	}
	attrs = append(attrs, r.attrs...)
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindProducer),
	}

	name := fmt.Sprintf("%s send", msg.Topic)
	ctx, _ = r.tracer.Start(ctx, name, opts...)
	r.propagators.Inject(ctx, carrier)
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
func (r *opentelemetry) TraceSubStart(ctx context.Context, groupID string, msg *kafka.Message) context.Context {
	carrier := NewMsgCarrier(msg)
	ctx = r.propagators.Extract(ctx, carrier)

	attrs := []attribute.KeyValue{
		semconv.MessagingKafkaMessageKey(string(msg.Key)),
		semconv.MessagingDestinationName(msg.Topic),
		semconv.MessagingOperationTypeReceive,
		semconv.MessagingOperationName("poll"),
		semconv.MessagingEventhubsConsumerGroup(groupID),
		semconv.MessagingDestinationPartitionID(strconv.FormatInt(int64(msg.Partition), 10)),
		semconv.MessagingMessageBodySize(len(msg.Value)),
		semconv.MessagingKafkaMessageOffset(int(msg.Offset)),
	}
	attrs = append(attrs, r.attrs...)
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindConsumer),
	}

	name := fmt.Sprintf("fetch from topic %s", msg.Topic)
	ctx, _ = r.tracer.Start(ctx, name, opts...)
	r.propagators.Inject(ctx, carrier)
	return ctx
}

func (r *opentelemetry) TraceSubEnd(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, ErrJsonUnmarshal) {
			span.SetAttributes(semconv.ErrorTypeKey.String(ErrJsonUnmarshal.Error()))
		}
	}

	span.End()
}

func (r *opentelemetry) TraceCommitMessagesStart(ctx context.Context, groupID string, messages ...kafka.Message) []context.Context {
	if messages == nil {
		return make([]context.Context, 0)
	}

	contexts := make([]context.Context, 0)

	for _, msg := range messages {
		carrier := NewMsgCarrier(&msg)
		ctx = r.propagators.Extract(ctx, carrier)

		attrs := []attribute.KeyValue{
			semconv.MessagingKafkaMessageKey(string(msg.Key)),
			semconv.MessagingDestinationName(msg.Topic),
			semconv.MessagingOperationTypeSettle,
			semconv.MessagingOperationName("commit"),
			semconv.MessagingEventhubsConsumerGroup(groupID),
			semconv.MessagingDestinationPartitionID(strconv.FormatInt(int64(msg.Partition), 10)),
			semconv.MessagingMessageBodySize(len(msg.Value)),
			semconv.MessagingKafkaMessageOffset(int(msg.Offset)),
		}
		attrs = append(attrs, r.attrs...)
		opts := []trace.SpanStartOption{
			trace.WithAttributes(attrs...),
			trace.WithSpanKind(trace.SpanKindConsumer),
		}

		name := fmt.Sprintf("commit from topic %s", msg.Topic)
		ctx, _ = r.tracer.Start(ctx, name, opts...)
		contexts = append(contexts, ctx)
	}

	return contexts
}

func (r *opentelemetry) TraceCommitMessagesEnd(ctx []context.Context, err error) {
	for _, c := range ctx {
		span := trace.SpanFromContext(c)
		if !span.IsRecording() {
			return
		}

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			span.SetAttributes(semconv.ErrorTypeKey.String("failed commit message"))
		}

		span.End()
	}
}
