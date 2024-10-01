package ekafka

import (
	"context"
	"github.com/segmentio/kafka-go"
)

func (b *broker) Subscribe(ctx context.Context, input SubInput) (output SubOutput, err error) {
	reader := kafka.NewReader(input.Config)

	output = SubOutput{
		Reader: Reader{
			R:            reader,
			subTracer:    b.subTracer,
			commitTracer: b.commitTracer,
			groupID:      input.Config.GroupID,
		},
	}
	return
}
