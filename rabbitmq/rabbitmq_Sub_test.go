package erabbitmq_test

import (
	"context"
	"fmt"
	erabbitmq "github.com/SyaibanAhmadRamadhan/event-bus/rabbitmq"
	"github.com/guregu/null/v5"
	"testing"
	"time"
)

func TestSub(t *testing.T) {
	NewOtel("sub")
	time.Sleep(2 * time.Second)

	r := erabbitmq.New("amqp://rabbitmq:pas12345@localhost:5672/", erabbitmq.WithReconnect(time.Second, null.Int{}), erabbitmq.WithOtel("amqp://rabbitmq:pas12345@localhost:5672/"))

	output, err := r.Subscribe(context.Background(), erabbitmq.SubInput{
		QueueName:    "notification_type_email",
		ConsumerName: "",
		AutoAck:      true,
		Exclusive:    false,
		NoLocal:      false,
		NoWait:       false,
		Args:         nil,
	})
	if err != nil {
		t.Error(err)
	}

	for delivery := range output.Deliveries {
		fmt.Println(string(delivery.Body))
		//delivery.Ack(false)
		//err := delivery.Acknowledger.Ack(delivery.DeliveryTag, false)
		//if err != nil {
		//	fmt.Println(err)
		//}
	}

	time.Sleep(10 * time.Second)
	r.Close()
}
