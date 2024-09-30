package erabbitmq

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"runtime/debug"
	"time"
)

type TracerPub interface {
	TracePubStart(ctx context.Context, input PubInput) context.Context
	TracePubEnd(ctx context.Context, input PubOutput, err error)
	RecordRetryPub(ctx context.Context, attempt int, err error)
}

type TracerSub interface {
	TraceSubStart(ctx context.Context, input SubInput, msg *amqp.Delivery, ch *amqp.Channel)
}
type TracerReconnection interface {
	TraceReConnStart(ctx context.Context, reconDelay time.Duration, attempt *int64) context.Context
	TraceReConnEnd(ctx context.Context, err error)
	RecordRetryReConn(ctx context.Context, attempt int64, err error)
}

type TracerCleanUpSpan interface {
	EndAllSpan()
}

func findOwnImportedVersion() {
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		for _, dep := range buildInfo.Deps {
			if dep.Path == TracerName {
				amqpLibVersion = dep.Version
			}
		}
	}
}

const TracerName = "github.com/SyaibanAhmadRamadhan/event-bus/rabbitmq"
const InstrumentVersion = "v1.0.0"
const amqpLibName = "github.com/rabbitmq/amqp091-go"
const netProtocolVer = "0.9.1"

var (
	amqpLibVersion = "unknown"
)
