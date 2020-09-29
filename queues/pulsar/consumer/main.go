package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/go-redis/redis/v8"
)

func main() {

	opt := &redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	rdb := redis.NewClient(opt)
	rdb.SetNX(context.Background(), "1", 1, time.Hour)

	client, err := pulsar.NewClient(pulsar.ClientOptions{URL: "pulsar://localhost:6650"})
	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()

	channel := make(chan pulsar.ConsumerMessage, 100)

	options := pulsar.ConsumerOptions{
		Topic:               "persistent://public/default/disp-topic",
		SubscriptionName:    "my-subscription",
		NackRedeliveryDelay: 5 * time.Second,
		Type:                pulsar.Shared,
		// RetryEnable:         true,
		// SubscriptionInitialPosition: pulsar.SubscriptionPositionEarliest,
	}
	fmt.Println(options.ReplicateSubscriptionState)
	options.MessageChannel = channel

	consumer, err := client.Subscribe(options)
	if err != nil {
		log.Fatal(err)
	}

	defer consumer.Close()

	for cm := range channel {
		msg := cm.Message
		fmt.Printf("Received message  msgId: %v -- content: '%s'   -- redelivery: %d \n",
			msg.ID(), string(msg.Payload()), msg.RedeliveryCount())
		// rdb.Set(context.Background(), "test", msg.ID().Serialize(), time.Hour)
		// consumer.ReconsumeLater(msg, time.Second)
		consumer.AckID(msg.ID())
	}
}
