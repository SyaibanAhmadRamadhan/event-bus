package ekafka_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/SyaibanAhmadRamadhan/event-bus/debezium"
	ekafka "github.com/SyaibanAhmadRamadhan/event-bus/kafka"
	"github.com/guregu/null/v5"
	"github.com/segmentio/kafka-go"
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

type UserValue struct {
	ID              int64     `json:"id"`
	Email           string    `json:"email"`
	Password        string    `json:"password"`
	RegisterAs      int16     `json:"register_as"`
	IsEmailVerified bool      `json:"is_email_verified"`
	CreatedAt       int64     `json:"created_at"`
	UpdatedAt       int64     `json:"updated_at"`
	DeletedAt       null.Time `json:"deleted_at"`
}

func TestName(t *testing.T) {
	NewOtel("testing pub sub kafka")
	k := ekafka.New(ekafka.KafkaCustomWriter(&kafka.Writer{
		Addr:  kafka.TCP("localhost:9092"),
		Topic: "testtopic",
	}), ekafka.WithOtel())

	//_, err := k.Publish(context.Background(), ekafka.PubInput{
	//	Messages: []kafka.Message{
	//		{
	//			Key:       []byte(fmt.Sprintf("address-%s", " req.RemoteAddr")),
	//			Value:     []byte("body1"),
	//			Partition: 1,
	//		},
	//	},
	//})
	//if err != nil {
	//	fmt.Println(err)
	//}
	//_, err = k.Publish(context.Background(), ekafka.PubInput{
	//	Messages: []kafka.Message{
	//		{
	//			Key:       []byte(fmt.Sprintf("address-%s", " req.RemoteAddr")),
	//			Value:     []byte("body2"),
	//			Partition: 1,
	//		},
	//	},
	//})
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	r, err := k.Subscribe(context.Background(), ekafka.SubInput{
		Config: kafka.ReaderConfig{
			Brokers: []string{"localhost:9092"},
			GroupID: "test-group",
			Topic:   "usersvc.public.users",
		},
	})
	if err != nil {
		fmt.Println(err)
	}

	msgs := make([]kafka.Message, 0)
	for {
		msgStruct := debezium.Envelope[UserValue, UserValue]{}
		msg, err := r.Reader.FetchMessage(context.Background(), &msgStruct)
		if err != nil {
			t.Log(err)
		}

		fmt.Println(msgStruct.Payload.Op)
		fmt.Println(msgStruct.Payload.Before.Ptr())

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
