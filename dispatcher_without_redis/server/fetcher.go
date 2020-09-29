package server

import (
	"errors"
	"time"

	"github.com/nsqio/go-nsq"
)

type Fetcher interface {
	Start(int, int) error
	Fetch() (Message, error)
}

var (
	NoMessageError = errors.New("No Message")
)

func getMessageID(topic string, channel string, msgID nsq.MessageID) string {
	return topic + "." + channel + "." + string(msgID[:])
}

type nsqFetcher struct {
	Fetcher
	consumers      []*nsq.Consumer
	msgCh          chan Message
	topic          string
	channel        string
	lookupds       []string
	consumerConfig *nsq.Config
}

func (f *nsqFetcher) Fetch() (Message, error) {
	select {
	case v := <-f.msgCh:
		return v, nil
	case <-time.After(time.Millisecond):
		return nil, NoMessageError
	}
}

func (f *nsqFetcher) HandleMessage(m *nsq.Message) error {
	//fmt.Println("handle message")
	if len(m.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		return nil
	}
	// nsq.NewMessage(id nsq.MessageID, body []byte)
	m.DisableAutoResponse()
	message := NewNsqMessage(f.topic, f.channel, m)
	f.msgCh <- message
	return nil
}

func (f *nsqFetcher) Start(parallel int, nsqMaxInflight int) error {

	nsqConfig := nsq.NewConfig()
	nsqConfig.MaxInFlight = nsqMaxInflight
	nsqConfig.MsgTimeout = 1 * time.Minute
	nsqConfig.MaxAttempts = 10
	nsqConfig.LowRdyIdleTimeout = 1 * time.Minute
	for i := 0; i < parallel; i++ {
		c, err := nsq.NewConsumer(f.topic, f.channel, nsqConfig)
		if err != nil {
			return err

		}
		c.AddHandler(f)
		f.consumers = append(f.consumers, c)
		c.ConnectToNSQLookupds(f.lookupds)
	}
	return nil

	// for _, c := range f.consumers {
	// 	c.ConnectToNSQLookupds(f.lookupds)
	// }
}

func NewFetcher(bufferCnt int, topic string, channel string, lookupds []string) Fetcher {

	fetcher := &nsqFetcher{
		msgCh:    make(chan Message, bufferCnt),
		topic:    topic,
		channel:  channel,
		lookupds: lookupds,
	}
	// for i := 0; i < parallel; i++ {
	// 	c, err := nsq.NewConsumer(topic, channel, config)
	// 	if err != nil {
	// 		return nil, err

	// 	}
	// 	c.AddHandler(fetcher)
	// 	fetcher.consumers = append(fetcher.consumers, c)
	// }
	return fetcher
}
