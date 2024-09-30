package erabbitmq_test

import (
	"context"
	"encoding/base64"
	"fmt"
	erabbitmq "github.com/SyaibanAhmadRamadhan/event-bus/rabbitmq"
	"github.com/guregu/null/v5"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"

	"log"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	main()
}

func main() {
	NewOtel("pub")
	time.Sleep(2 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	r := erabbitmq.New("amqp://rabbitmq:pas12345@localhost:5672/", erabbitmq.WithReconnect(time.Second, null.Int{}), erabbitmq.WithOtel("amqp://rabbitmq:pas12345@localhost:5672/"))

	//var mu sync.Mutex
	successCount := 0

	//for i := 0; i < 1; i++ {
	//go func(msg string) {

	go func() {
		output, err := r.Publish(context.Background(), erabbitmq.PubInput{
			ExchangeName: "notifications",
			RoutingKey:   "notification.type.email",
			Mandatory:    false,
			Immediate:    false,
			MaxRetry:     3,
			DelayRetry:   time.Second * 2,
			Msg: amqp.Publishing{
				Body: []byte("msg"),
				//Headers: amqp.Table{
				//	"corre": "123",
				//},
			},
		})
		if err != nil {
			log.Println(err)
		} else {
			successCount++
			fmt.Println("send msg: ", "msg")
		}
		log.Println(output)
	}()

	go func() {
		time.Sleep(10 * time.Second)
		output, err := r.Publish(context.Background(), erabbitmq.PubInput{
			ExchangeName: "notifications",
			RoutingKey:   "notification.type.email",
			Mandatory:    false,
			Immediate:    false,
			MaxRetry:     3,
			Msg: amqp.Publishing{
				Body: []byte("msg"),
			},
		})
		if err != nil {
			log.Println(err)
		} else {
			successCount++
			fmt.Println("send msg: ", "msg")
		}
		log.Println(output)
	}()

	<-ctx.Done()
	//r.Close()

	//}(fmt.Sprintf("Message %d", 1))

	//if i == 50 {
	//	time.Sleep(10 * time.Second)
	//}
	//if i == 70 {
	//	time.Sleep(10 * time.Second)
	//}
	//r.Close()
	//}

	// Print total success count
	fmt.Printf("Total messages successfully sent: %d\n", successCount)
	time.Sleep(30 * time.Hour)
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

	//closeFn := func(ctx context.Context) (err error) {
	//	ctxClosure, cancelClosure := context.WithTimeout(ctx, 5*time.Second)
	//	defer cancelClosure()
	//
	//	if err = t.exporter.Shutdown(ctxClosure); err != nil {
	//		return err
	//	}
	//
	//	if err = provider.Shutdown(ctxClosure); err != nil {
	//		return err
	//	}
	//
	//	return
	//}

	return provider, err
}
