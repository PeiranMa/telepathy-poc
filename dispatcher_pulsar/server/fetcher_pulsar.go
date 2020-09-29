package server

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	log "github.com/sirupsen/logrus"
)

type Fetcher interface {
	Start(int) error
	Fetch() (*pulsar.ConsumerMessage, error)
	AckID(pulsar.MessageID)
	NackID(pulsar.MessageID)
}

var (
	NoMessageError = errors.New("No Message")
)

type pulsarFetcher struct {
	Fetcher
	pulsarConsumers []pulsar.Consumer
	msgCh           chan pulsar.ConsumerMessage
	topic           string
	pulsarProxy     string
	pulsarClient    pulsar.Client
	consumerConfig  pulsar.ConsumerOptions
}

func (f *pulsarFetcher) Fetch() (*pulsar.ConsumerMessage, error) {
	count := 0
	for {
		select {
		case v := <-f.msgCh:
			return &v, nil
		case <-time.After(100 * time.Millisecond):
			if count < 10 {
				count++
				continue
			} else {
				fmt.Println("No Message")
				return nil, NoMessageError
			}

		}
	}
}

func (f *pulsarFetcher) Start(parallel int) error {

	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ticker.C:
				fmt.Println(len(f.msgCh))
			}
		}
	}()

	options := pulsar.ConsumerOptions{
		Topic:                       "persistent://public/default/" + f.topic,
		SubscriptionName:            "my-subscription",
		Type:                        pulsar.Shared,
		MessageChannel:              f.msgCh,
		SubscriptionInitialPosition: pulsar.SubscriptionPositionEarliest,
		ReceiverQueueSize:           1000,
	}
	f.consumerConfig = options
	for i := 0; i < parallel; i++ {
		consumer, err := f.pulsarClient.Subscribe(f.consumerConfig)
		if err != nil {
			log.Fatal(err)
		}
		f.pulsarConsumers = append(f.pulsarConsumers, consumer)
	}
	return nil

}

func (f *pulsarFetcher) AckID(id pulsar.MessageID) {
	f.pulsarConsumers[rand.Intn(len(f.pulsarConsumers))].AckID(id)
}
func (f *pulsarFetcher) NackID(id pulsar.MessageID) {
	f.pulsarConsumers[rand.Intn(len(f.pulsarConsumers))].NackID(id)
}

func NewFetcher(topic string, pulsarProxy string) *pulsarFetcher {
	client, err := pulsar.NewClient(pulsar.ClientOptions{URL: "pulsar://" + pulsarProxy + ":6650"})
	if err != nil {
		log.Fatal(err)
	}
	fetcher := &pulsarFetcher{
		msgCh:        make(chan pulsar.ConsumerMessage, 10000),
		topic:        topic,
		pulsarProxy:  pulsarProxy,
		pulsarClient: client,
	}
	return fetcher

}
