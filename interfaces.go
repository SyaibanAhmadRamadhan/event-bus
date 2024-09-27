package eventbus

import (
	"context"
)

//type RabbitMQPubSub = pubSub[rabbitmq.PubInput, rabbitmq.PubOutput, rabbitmq.SubInput, rabbitmq.SubOutput]

type pubSub[pubInput, pubOutput, subInput, subOutput any] interface {
	Publish(ctx context.Context, input pubInput) (output pubOutput, err error)
	Subscribe(ctx context.Context, input subInput) (output subOutput, err error)
}

//func NewMockRabbitMQPubSub(mock *gomock.Controller) *MockpubSub[rabbitmq.PubInput, rabbitmq.PubOutput, rabbitmq.SubInput, rabbitmq.SubOutput] {
//	return NewMockpubSub[rabbitmq.PubInput, rabbitmq.PubOutput, rabbitmq.SubInput, rabbitmq.SubOutput](mock)
//}
