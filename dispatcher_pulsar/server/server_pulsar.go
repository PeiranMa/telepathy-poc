package server

import (
	"context"
	"errors"
	"fmt"
	// "strings"
	"sync"
	// "time"

	"github.com/apache/pulsar-client-go/pulsar"
	pb "poc.dispatcher/protos"
)

var (

	// errors
	TaskAlreadyInWorkingState = errors.New("TaskAlreadyInWorkingState")
	TaskAlreadyFinished       = errors.New("TaskAlreadyFinished")
	TaskAlreadyFailed         = errors.New("TaskAlreadyFailed")
)

const (
	process uint8 = 0
	finish  uint8 = 1
	failed  uint8 = 2
)

type server struct {
	pb.DispatcherServer

	fetchers     Cache
	results      Cache
	msgStates    Cache
	publisher    *pulsarPublisher
	reader       *pulsarReader
	pulsarProxy  string
	publisherNum int
	consumerNum  int

	readCh  chan pulsar.Message
	mux     sync.Mutex
	verbose bool
}

func (s *server) println(sent ...interface{}) {
	if s.verbose {
		fmt.Println(sent)
	}
}

func (s *server) createFetcherIfNotExist(topic string) (Fetcher, error) {
	if s.fetchers.Exists(topic) {
		f, err := s.fetchers.Get(topic)
		if err != nil {
			s.println("Get Fetcher failed", err)

			return nil, err
		}
		return f.(Fetcher), nil
	} else {
		f := NewFetcher(topic, s.pulsarProxy)
		err := s.fetchers.SetNX(topic, f, 0)
		if err == nil {
			err = f.Start(s.consumerNum)
			if err != nil {
				s.println("Fetcher Start error", err)
				return nil, err
			}
		}
		return f, nil

	}

}

func (s *server) createResultIfNotExist(topic string) (Publisher, error) {
	if s.results.Exists(topic) {
		p, err := s.results.Get(topic)
		if err != nil {
			s.println("Get Publisher failed", err)

			return nil, err
		}
		return p.(Publisher), nil

	} else {
		p := NewPublisher(topic, s.pulsarProxy)
		err := s.results.SetNX(topic, p, 0)
		if err == nil {
			err = p.Start(s.publisherNum)
			if err != nil {
				s.println("Fetcher Start error", err)
				return nil, err
			}
		}
		return p, err

	}
}

func (s *server) GetTask(ctx context.Context, in *pb.TaskRequest) (*pb.TaskResponse, error) {
	topic := in.Topic
	f, err := s.createFetcherIfNotExist(topic)
	if err != nil {
		return nil, err
	}

	// fetch the msg
	// fmt.Println("Before Fetch", in.Channel)
	cMsg, err := f.Fetch()
	// fmt.Println("After Fetch", in.Channel)
	if err != nil {
		return nil, err
	}
	msg := cMsg.Message
	msgID := msg.ID().Serialize()
	// fmt.Println("Before SetDiff", in.Channel)
	err = s.msgStates.SetDiff(string(msgID), "process", -1)
	// fmt.Println("After SetDiff", in.Channel)
	if err != nil {
		if err == TaskAlreadyFinished {
			s.println("Message already finish", in.Channel)
			f.AckID(msg.ID())
			return nil, TaskAlreadyFinished
		} else {
			s.println("Message is processing", in.Channel)
			return nil, TaskAlreadyInWorkingState
		}

	}
	// fmt.Println("Before Publish", in.Channel)
	s.publisher.Publish(append(msgID, process))
	// fmt.Println("After Publish", in.Channel)
	resp := &pb.TaskResponse{Payload: msg.Payload(), MessageID: msgID}
	return resp, nil

}

func (s *server) FinTask(ctx context.Context, in *pb.FinTaskRequest) (*pb.FinTaskResponse, error) {
	msgID := in.MessageID

	switch in.Result {
	// If the result is success, take a radical approach
	case pb.TaskResult_FIN:
		err := s.msgStates.SetDiff(string(msgID), "finish", -1)
		if err != nil {
			s.println("Message already finish : ", err)
			break
		}
		//if SetDiff succeed, then publish, try finish msg, set the result

		f, err := s.createFetcherIfNotExist(in.Topic)
		if err != nil {
			return nil, err
		}
		idForAck, err := pulsar.DeserializeMessageID(msgID)
		if err != nil {
			return nil, err
		}
		f.AckID(idForAck)
		s.publisher.Publish(append(msgID, finish))
		r, err := s.createResultIfNotExist(in.Topic + ".result")
		if err != nil {
			return nil, err
		}
		r.Publish(in.Payload)
		break
	case pb.TaskResult_FAIL:
		s.println("Failed Task, ID: ", msgID)
		err := s.msgStates.SetDiff(string(msgID), "failed", -1)
		if err != nil {
			if err == TaskAlreadyFinished {
				s.println("Message already finish")
				return nil, TaskAlreadyFinished
			} else {
				s.println("Message is already failed")
				return nil, TaskAlreadyFailed
			}

		}
		//if SetDiff succeed, then publish,  try reque
		s.publisher.Publish(append(msgID, failed))
		f, err := s.createFetcherIfNotExist(in.Topic)
		if err != nil {
			return nil, err
		}
		idForAck, err := pulsar.DeserializeMessageID(msgID)
		if err != nil {
			return nil, err
		}
		f.NackID(idForAck)
	}
	return &pb.FinTaskResponse{}, nil
}

func (s *server) getStatesLoop() {
	go func() {
		for {
			select {
			case msg := <-s.readCh:
				msgState := msg.Payload()
				msgID := msgState[:len(msgState)-1]
				state := msgState[len(msgState)-1]
				var stateStr string
				if state == process {
					stateStr = "process"
				} else if state == finish {
					stateStr = "finish"
				} else if state == failed {
					stateStr = "failed"
				} else {
					panic("Message state not exist")
				}

				err := s.msgStates.SetDiff(string(msgID), stateStr, -1)
				if err != nil {
					// log.Info("getStatesLoop :", err)

				}

			}
		}
	}()

}

func NewServer(pulsarProxy string, publisherProducerNum int, fetcherConsumerNum int) pb.DispatcherServer {

	s := &server{
		fetchers:     NewInMemoryCache(),
		results:      NewInMemoryCache(),
		msgStates:    NewInMemoryCache(),
		pulsarProxy:  pulsarProxy,
		publisherNum: publisherProducerNum,
		consumerNum:  fetcherConsumerNum,
		readCh:       make(chan pulsar.Message, 1000),
		verbose:      false,
	}
	s.publisher = NewPublisher("States_test", s.pulsarProxy)
	s.reader = NewReader("States_test", s.readCh, s.pulsarProxy)
	err := s.publisher.Start(publisherProducerNum)
	if err != nil {
		panic(err.Error())
	}
	err = s.reader.Start()
	if err != nil {
		panic(err.Error())
	}
	s.getStatesLoop()

	return s

}
