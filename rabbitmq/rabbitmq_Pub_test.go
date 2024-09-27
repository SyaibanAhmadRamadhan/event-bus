package erabbitmq_test

import (
	"context"
	"fmt"
	"github.com/guregu/null/v5"
	amqp "github.com/rabbitmq/amqp091-go"
	erabbitmq "go-event-bus/rabbitmq"
	"log"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	main()
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	r := erabbitmq.New("amqp://rabbitmq:pas12345@localhost:5672/", erabbitmq.WithReconnect(time.Second, null.Int{}))

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
		time.Sleep(2 * time.Second)
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
	r.Close()

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
	//time.Sleep(10 * time.Second)
}
