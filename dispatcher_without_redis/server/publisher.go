package server

import (
	"errors"
	"fmt"
	"time"

	"github.com/nsqio/go-nsq"
)

type Publisher interface {
	Start(int) error
	Publish(string)
}

var (
	PublishError = errors.New("Can't Publish Message")
)

type nsqPublisher struct {
	Publisher

	msgCh    chan string
	topic    string
	nsqdAddr string

	doneCh    chan *nsq.ProducerTransaction
	batchSize int
}

func (p *nsqPublisher) Start(parallel int) error {

	nsqConfig := nsq.NewConfig()
	for i := 0; i < parallel; i++ {
		go func() {
			producer, err := nsq.NewProducer(p.nsqdAddr, nsqConfig)
			if err != nil {
				panic("Can't create producer")
			}
			batchCh := make(chan string, p.batchSize)

			for {
				var batch [][]byte
				select {
				case msg := <-p.msgCh:
					select {
					case batchCh <- msg:
						continue
					default:
						start := time.Now()
						batchSize := len(batchCh)
						for i := 0; i < batchSize; i++ {
							if val, ok := <-batchCh; ok {
								batch = append(batch, []byte(val))
							} else {
								panic("unknow")
							}

						}

						batchCh <- msg

						producer.MultiPublishAsync(p.topic, batch, p.doneCh)
						// if err != nil {
						// 	panic("can't publish msg : " + err.Error())
						// }
						fmt.Println("batch time : ", time.Since(start))

					}

				case <-time.NewTimer(time.Millisecond).C:

					if len(batchCh) == 0 {
						select {
						case msg := <-p.msgCh:
							batchCh <- msg
							continue
						}
					}
					batchSize := len(batchCh)
					for i := 0; i < batchSize; i++ {

						if val, ok := <-batchCh; ok {
							batch = append(batch, []byte(val))
						} else {
							panic("unknow")
						}

					}

					producer.MultiPublishAsync(p.topic, batch, p.doneCh)
					// if err != nil {
					// 	panic("can't publish msg : " + err.Error())
					// }

				}
			}

			// producer, err := nsq.NewProducer(p.nsqdAddr, nsqConfig)
			// if err != nil {
			// 	panic("Can't create producer")
			// }

			// for {
			// 	select {

			// 	case msg := <-p.msgCh:

			// 		producer.PublishAsync("sync", []byte(msg), p.doneCh)
			// 	}
			// }

		}()

	}
	go func() {
		for {
			var pt *nsq.ProducerTransaction
			select {
			case pt = <-p.doneCh:
				if pt.Error != nil {
					panic("Publish failed " + pt.Error.Error())
				}
			}
		}
	}()
	return nil
}

func (p *nsqPublisher) Publish(msg string) {
	// start := time.Now()
	select {
	case p.msgCh <- msg:
	}

	// fmt.Println(time.Since(start))
	// return p.producer.PublishAsync(topic, []byte(msg), p.doneCh)

}

func NewPublisher(topic string, nsqdAddr string, batchSize int) Publisher {

	publisher := &nsqPublisher{
		msgCh:     make(chan string, 1000),
		topic:     topic,
		nsqdAddr:  nsqdAddr,
		doneCh:    make(chan *nsq.ProducerTransaction),
		batchSize: batchSize,
	}

	// publisher.producer, _ = nsq.NewProducer(publisher.nsqdAddr, nsq.NewConfig())
	return publisher
}
