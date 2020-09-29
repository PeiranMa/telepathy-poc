package server

import (
	// "fmt"
	// "errors"
	"time"

	"github.com/nsqio/go-nsq"
)

type Responser interface {
	Start(int, int) error
}

type nsqResponser struct {
	Responser
	consumers      []*nsq.Consumer
	msgCh          chan *nsq.Message
	topic          string
	channel        string
	lookupds       []string
	consumerConfig *nsq.Config
}

// func (f *nsqFetcher) Fetch() (Message, error) {
// 	select {
// 	case v := <-f.msgCh:
// 		return v, nil
// 	case <-time.After(time.Millisecond):
// 		return nil, NoMessageError
// 	}
// }

func (r *nsqResponser) HandleMessage(m *nsq.Message) error {
	//fmt.Println("handle message")
	if len(m.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		return nil
	}
	// nsq.NewMessage(id nsq.MessageID, body []byte)

	r.msgCh <- m
	return nil
}

func (r *nsqResponser) Start(parallel int, nsqMaxInflight int) error {

	nsqConfig := nsq.NewConfig()
	nsqConfig.MaxInFlight = nsqMaxInflight
	nsqConfig.MsgTimeout = 1 * time.Minute
	nsqConfig.MaxAttempts = 10
	nsqConfig.LowRdyIdleTimeout = 1 * time.Minute

	for i := 0; i < parallel; i++ {

		c, err := nsq.NewConsumer(r.topic, r.channel, nsqConfig)
		if err != nil {
			return err

		}

		c.AddHandler(r)
		r.consumers = append(r.consumers, c)
		c.ConnectToNSQLookupds(r.lookupds)

	}
	return nil

	// for _, c := range f.consumers {
	// 	c.ConnectToNSQLookupds(f.lookupds)
	// }
}

func NewResponser(topic string, channel string, lookupds []string, msgCh chan *nsq.Message) Responser {

	responser := &nsqResponser{
		msgCh:    msgCh,
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
	return responser
}
