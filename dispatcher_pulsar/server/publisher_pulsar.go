package server

import (
	"context"
	"errors"
	// "fmt"
	// "time"

	"github.com/apache/pulsar-client-go/pulsar"
	log "github.com/sirupsen/logrus"
)

type Publisher interface {
	Start(int) error
	Publish([]byte)
}

var (
	PublishError = errors.New("Can't Publish Message")
)

type pulsarPublisher struct {
	Publisher
	pulsarClient    pulsar.Client
	pulsarProducers []pulsar.Producer
	msgCh           chan []byte
	topic           string
	pulsarProxy     string
	producerConfig  pulsar.ProducerOptions
}

func (p *pulsarPublisher) Start(parallel int) error {
	p.producerConfig = pulsar.ProducerOptions{
		Topic: p.topic,
	}
	for i := 0; i < parallel; i++ {
		go func() {
			producer, err := p.pulsarClient.CreateProducer(p.producerConfig)
			if err != nil {
				panic("Can't create producer")
			}
			p.pulsarProducers = append(p.pulsarProducers, producer)
			for {
				select {
				case v := <-p.msgCh:
					producer.SendAsync(context.Background(), &pulsar.ProducerMessage{
						Payload: v,
					}, func(msgID pulsar.MessageID, message *pulsar.ProducerMessage, e error) {
						if e != nil {
							log.WithError(e).Fatal("Failed to publish, try publish again")
							p.msgCh <- message.Payload

						}

					})

				}
			}

		}()

	}
	return nil
}

func (p *pulsarPublisher) Publish(msg []byte) {
	// start := time.Now()
	select {
	case p.msgCh <- msg:
	}

	// fmt.Println(time.Since(start))
	// return p.producer.PublishAsync(topic, []byte(msg), p.doneCh)
}

func NewPublisher(topic string, pulsarProxy string) *pulsarPublisher {
	client, err := pulsar.NewClient(pulsar.ClientOptions{URL: "pulsar://" + pulsarProxy + ":6650"})
	if err != nil {
		log.Fatal(err)
	}
	publisher := &pulsarPublisher{

		topic:        topic,
		msgCh:        make(chan []byte),
		pulsarProxy:  pulsarProxy,
		pulsarClient: client,
	}
	return publisher

}
