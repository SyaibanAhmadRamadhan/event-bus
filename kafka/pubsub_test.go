package ekafka_test

import (
	"context"
	"encoding/base64"
	"fmt"
	ekafka "github.com/SyaibanAhmadRamadhan/event-bus/kafka"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	NewOtel("testing pub sub kafka")
	mechanism := plain.Mechanism{
		Username: "kafka",
		Password: "pass12345",
	}
	sharedTransport := &kafka.Transport{
		SASL: mechanism,
	}
	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: mechanism,
	}
	k := ekafka.New(ekafka.KafkaCustomWriter(&kafka.Writer{
		Addr:      kafka.TCP("localhost:8003"),
		Topic:     "testtopic",
		Transport: sharedTransport,
	}), ekafka.WithOtel())

	_, err := k.Publish(context.Background(), ekafka.PubInput{
		Messages: []kafka.Message{
			{
				Key:       []byte(fmt.Sprintf("address-%s", " req.RemoteAddr")),
				Value:     []byte("body1"),
				Partition: 1,
			},
		},
	})
	if err != nil {
		fmt.Println(err)
	}
	_, err = k.Publish(context.Background(), ekafka.PubInput{
		Messages: []kafka.Message{
			{
				Key:       []byte(fmt.Sprintf("address-%s", " req.RemoteAddr")),
				Value:     []byte("body2"),
				Partition: 1,
			},
		},
	})
	if err != nil {
		fmt.Println(err)
	}

	r, err := k.Subscribe(context.Background(), ekafka.SubInput{
		Config: kafka.ReaderConfig{
			Brokers: []string{"localhost:8003"},
			GroupID: "test-group",
			Topic:   "testtopic",
			Dialer:  dialer,
		},
	})
	if err != nil {
		fmt.Println(err)
	}

	msgs := make([]kafka.Message, 0)
	for {
		msg, err := r.Reader.FetchMessage(context.Background())
		if err != nil {
			t.Log(err)
		}

		msgs = append(msgs, msg)
		if len(msgs) > 1 {
			err = r.Reader.CommitMessages(context.Background(), msgs...)
			if err != nil {
				t.Log(err)
			}
		}
	}
}

func NewOtel(tracerName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", "user", "pw")))
	traceCli := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithHeaders(map[string]string{
			"Authorization": authHeader,
		}),
		otlptracegrpc.WithEndpoint("0.0.0.0:4317"),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	)

	traceExp, err := otlptrace.New(ctx, traceCli)
	if err != nil {
		panic(err)
	}

	otelProvider := &otelProvider{
		name: tracerName,
	}

	traceProvide, err := otelProvider.start(traceExp)

	if err != nil {
		panic(err)
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(traceProvide)

	return
}

type otelProvider struct {
	name     string
	exporter trace.SpanExporter
}

func (t *otelProvider) start(exp trace.SpanExporter) (*trace.TracerProvider, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(t.name),
		),
	)
	if err != nil {
		err = fmt.Errorf("failed to created resource: %w", err)
		return nil, err
	}

	t.exporter = exp
	bsp := trace.NewBatchSpanProcessor(t.exporter)

	provider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
		trace.WithSpanProcessor(bsp),
	)

	return provider, err
}
