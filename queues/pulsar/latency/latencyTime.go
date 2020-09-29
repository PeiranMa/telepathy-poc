package main

import (
	"context"
	"fmt"
	// "log"
	// "math"
	"sync"

	"time"

	. "github.com/apache/pulsar-client-go/pulsar"
	"github.com/bmizerany/perks/quantile"
	log "github.com/sirupsen/logrus"
)

var (
	serviceURL = "pulsar://pulsar-proxy:6650"
)
var mux sync.Mutex

func main() {
	fmt.Println("Start!")

	client, err := NewClient(ClientOptions{
		URL: "pulsar://pulsar-proxy:6650",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	pubTimes := make(map[string]time.Time)
	channel := make(chan ConsumerMessage, 100)

	options := ConsumerOptions{
		Topic:               "persistent://public/default/latency",
		SubscriptionName:    "my-subscription",
		NackRedeliveryDelay: time.Duration(5) * time.Second,
		Type:                Shared,
		// RetryEnable:         true,
		// SubscriptionInitialPosition: pulsar.SubscriptionPositionEarliest,
	}

	options.MessageChannel = channel

	consumer, err := client.Subscribe(options)
	if err != nil {
		log.Fatal(err)
	}
	q := quantile.NewTargeted(0.50, 0.95, 0.99, 0.999, 1.0)
	go func() {
		for cm := range channel {
			msg := cm.Message
			mux.Lock()
			msgStartTime, ok := pubTimes[string(msg.ID().Serialize())]
			mux.Unlock()
			if !ok {
				// log.Infof("msg not in pubs")
				continue
			}
			q.Insert(time.Since(msgStartTime).Seconds())
			consumer.AckID(msg.ID())
		}
	}()
	defer consumer.Close()
	producer, err := client.CreateProducer(ProducerOptions{
		Topic: "latency",
	})

	var start time.Time
	go func() {
		for {
			start = time.Now()
			producer.SendAsync(context.Background(), &ProducerMessage{
				Payload: make([]byte, 100)},
				func(msgID MessageID, message *ProducerMessage, e error) {
					if e != nil {
						panic("Send Async fail")
					}
					mux.Lock()
					pubTimes[string(msgID.Serialize())] = start
					mux.Unlock()
				})
		}
	}()
	defer producer.Close()

	tick := time.NewTicker(10 * time.Second)
	for {
		select {

		case <-tick.C:

			log.Infof(`Latency ms: 50%% %5.1f -95%% %5.1f - 99%% %5.1f - 99.9%% %5.1f - max %6.1f`,

				q.Query(0.5)*1000,
				q.Query(0.95)*1000,
				q.Query(0.99)*1000,
				q.Query(0.999)*1000,
				q.Query(1.0)*1000,
			)

			q.Reset()

		}
	}

}
