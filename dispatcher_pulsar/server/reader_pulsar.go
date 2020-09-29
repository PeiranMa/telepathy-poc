package server

import (
	"context"
	// "fmt"
	// "errors"
	// "strings"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	log "github.com/sirupsen/logrus"
)

type Reader interface {
	Start() error
}

type pulsarReader struct {
	Reader

	msgCh        chan pulsar.Message
	pulsarClient pulsar.Client
	reader       pulsar.Reader
	topic        string
	pulsarProxy  string
	readerConfig pulsar.ReaderOptions
}

func (r *pulsarReader) Start() error {
	reader, err := r.pulsarClient.CreateReader(pulsar.ReaderOptions{
		Topic:          r.topic,
		StartMessageID: pulsar.EarliestMessageID(),
	})
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			//No Channel for reader, need to change
			for reader.HasNext() {
				msg, err := reader.Next(context.Background())
				if err != nil {
					log.Fatal(err)
				}
				r.msgCh <- msg
			}
			select {
			case <-time.After(1000 * time.Millisecond):
			}
		}
	}()
	return nil
}

func NewReader(topic string, msgCh chan pulsar.Message, pulsarProxy string) *pulsarReader {
	client, err := pulsar.NewClient(pulsar.ClientOptions{URL: "pulsar://" + pulsarProxy + ":6650"})
	if err != nil {
		log.Fatal(err)
	}
	reader := &pulsarReader{

		topic:        topic,
		msgCh:        msgCh,
		pulsarProxy:  pulsarProxy,
		pulsarClient: client,
	}
	return reader

}
