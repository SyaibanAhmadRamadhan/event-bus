package ekafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	"runtime/debug"
)

type TracerPub interface {
	TracePubStart(ctx context.Context, msg *kafka.Message) context.Context
	TracePubEnd(ctx context.Context, input PubOutput, err error)
}

type TracerSub interface {
	TraceSubStart(ctx context.Context, groupID string, msg *kafka.Message) context.Context
	TraceSubEnd(ctx context.Context, err error)
}

type TracerCommitMessage interface {
	TraceCommitMessagesStart(ctx context.Context, groupID string, messages ...kafka.Message) []context.Context
	TraceCommitMessagesEnd(ctx []context.Context, err error)
}

func findOwnImportedVersion() {
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		for _, dep := range buildInfo.Deps {
			if dep.Path == TracerName {
				kafkaLibVersion = dep.Version
			}
		}
	}
}

const TracerName = "github.com/SyaibanAhmadRamadhan/event-bus/kafka"
const InstrumentVersion = "v1.0.0"
const kafkaLibName = "github.com/segmentio/kafka-go"
const netProtocolVer = "0.9.1"

var (
	kafkaLibVersion = "unknown"
)
