package eventbus

import (
	"context"
)

type PubSub[pubInput, pubOutput, subInput, subOutput any] interface {
	Publish(ctx context.Context, input pubInput) (output pubOutput, err error)
	Subscribe(ctx context.Context, input subInput) (output subOutput, err error)
}
