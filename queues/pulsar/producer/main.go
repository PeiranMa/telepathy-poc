package main

import (
	"context"
	"fmt"
	"log"
	// "time"

	. "github.com/apache/pulsar-client-go/pulsar"
	// "github.com/go-redis/redis/v8"
)

var (
	serviceURL = "pulsar://localhost:6650"
)

func main() {

	client, err := NewClient(ClientOptions{
		URL: "pulsar://localhost:6650",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	producer, err := client.CreateProducer(ProducerOptions{
		Topic: "disp-topic",
	})

	for i := 0; i < 10; i++ {
		msgID, err := producer.Send(context.Background(), &ProducerMessage{
			Payload: []byte("hello"),
		})
		if err != nil {
			panic("Publish failed")
		}
		fmt.Println("Message ID :", msgID)
	}

	defer producer.Close()

}
