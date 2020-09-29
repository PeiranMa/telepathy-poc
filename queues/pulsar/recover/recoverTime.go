package main

import (
	"context"
	"fmt"
	// "log"
	// "math"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	. "github.com/apache/pulsar-client-go/pulsar"
	log "github.com/sirupsen/logrus"
)

var (
	serviceURL = "pulsar://pulsar-proxy:6650"
)

func main() {
	fmt.Println("Start!")
	runtime.SetBlockProfileRate(1)
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()
	client, err := NewClient(ClientOptions{
		URL: "pulsar://localhost:6650",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	states := make(map[string][]byte)
	msgChannel := make(chan ReaderMessage)
	reader, err := client.CreateReader(ReaderOptions{
		Topic:             "latency",
		StartMessageID:    EarliestMessageID(),
		ReceiverQueueSize: 10000,
		MessageChannel:    msgChannel,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	// channel := make(chan ConsumerMessage, 100)
	// options := ConsumerOptions{
	// 	Topic:            "persistent://public/default/latency",
	// 	SubscriptionName: "my-subscription-new",

	// 	SubscriptionInitialPosition: SubscriptionPositionEarliest,
	// }
	// fmt.Println(options.ReplicateSubscriptionState)
	// options.MessageChannel = channel

	// consumer, err := client.Subscribe(options)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer consumer.Close()

	count := 0
	start := time.Now()
	last := time.Now()
	var batchCount float64 = 0
	var totalTime float64 = 0

	// tick := time.NewTicker(10 * time.Second)
	// for {
	// 	select {
	// 	case cm, ok := <-consumer.Chan():
	// 		if !ok {
	// 			return
	// 		}
	// 		count++
	// 		states[string(cm.Message.ID().Serialize())] = cm.Message.Payload()
	// 		consumer.Ack(cm.Message)

	// 		if count > 1000000 {
	// 			log.Infof("Recovery time consumed of 1 millions messages is  %f s.", time.Now().Sub(last).Seconds())
	// 			last = time.Now()
	// 			batchCount++
	// 			totalTime = time.Now().Sub(start).Seconds()
	// 			log.Infof("The average recovery time consumed of 1 millions messages is %f s.", totalTime/batchCount)
	// 			count = 0

	// 		}
	// 	case <-tick.C:
	// 		currentMsgReceived := count

	// 		msgRate := float64(currentMsgReceived) / time.Now().Sub(start).Seconds()

	// 		log.Infof(`Stats - Consume rate: %6.1f msg/s`,
	// 			msgRate)

	// 	}
	// }

	for reader.HasNext() {
		msg, err := reader.Next(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		count++
		states[string(msg.ID().Serialize())] = msg.Payload()
		if count > 1000000 {
			log.Infof("Recovery time consumed of 1 millions messages is  %f s.", time.Now().Sub(last).Seconds())
			last = time.Now()
			batchCount++
			totalTime = time.Now().Sub(start).Seconds()
			log.Infof("The average recovery time consumed of 1 millions messages is %f s.", totalTime/batchCount)
			count = 0

		}

	}

}
