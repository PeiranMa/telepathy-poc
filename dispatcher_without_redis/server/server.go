package server

import (
	"context"
	"errors"
	"fmt"
	"strings"
	// "os"
	// "runtime"
	// "strings"
	"sync"
	"time"

	// "github.com/go-redis/redis/v8"

	// "github.com/google/uuid"
	"github.com/RussellLuo/timingwheel"
	"github.com/nsqio/go-nsq"
	pb "poc.dispatcher/protos"
)

var (

	// errors
	TaskAlreadyInWorkingState = errors.New("TaskAlreadyInWorkingState")
	TaskAlreadyFinished       = errors.New("TaskAlreadyFinished")
	TaskAlreadyFailed         = errors.New("TaskAlreadyFailed")
)

type server struct {
	pb.DispatcherServer
	serverName    string
	fetchers      Cache
	results       Cache
	msgDispatched Cache // local state, message has been dispatched but not confirm
	// msgStates      *memoryStateCache
	msgStates Cache

	nsqlookups     []string
	nsqdAddr       string
	fetcherBuf     int
	nsqMaxInflight int
	consumerNum    int
	publisher      Publisher
	responser      Responser
	responseCh     chan *nsq.Message
	mux            sync.Mutex
	verbose        bool

	timers      *timingwheel.Timer
	nsqlookupsD []string
	batchSize   int
	timer       Timer
}

func getFetcherID(topic string, channel string) string {
	return topic + "." + channel
}

func getMessageState(msg string) (string, string) {
	tmp := strings.Split(msg, ".")
	if len(tmp) != 4 {
		panic("can't get message state")
	}
	return strings.Join(tmp[0:3], "."), tmp[3]
}

func (s *server) println(sent ...interface{}) {
	if s.verbose {
		fmt.Println(sent)
	}
}

func duationDivision(a time.Duration, b time.Duration) int {

	return int(float64(a.Nanoseconds()) / float64(b.Nanoseconds()) * 100)

}

func (s *server) GetTask(ctx context.Context, in *pb.TaskRequest) (*pb.TaskResponse, error) {
	// start := time.Now()
	// defer func() {
	// 	s.println("GetTask Cost", time.Now().Sub(start))
	// }()

	// Fetch msg from Fetcher
	// If exists, put it into `working_msg` cache.
	start := time.Now()
	fid := getFetcherID(in.Topic, in.Channel)
	var fetcher interface{}
	if s.fetchers.Exists(fid) {
		f, err := s.fetchers.Get(fid)
		if err != nil {
			s.println("Get Fetcher failed", err)

			return nil, err
		}
		fetcher = f
	} else {
		f := NewFetcher(s.fetcherBuf, in.Topic, in.Channel, s.nsqlookups)
		err := s.fetchers.SetNX(fid, f, 0)
		if err == nil {
			err = f.Start(s.consumerNum, s.nsqMaxInflight)
			if err != nil {
				s.println("Fetcher Start error", err)
				return nil, err
			}
		}
		fetcher = f

	}

	// fetch the msg

	msg, err := fetcher.(Fetcher).Fetch()

	if err != nil {
		// s.println("Fetch msg Failed, reason: ", err)
		return nil, err
	}

	msgID := msg.GetID().String()
	err = s.msgStates.SetDiff(msgID, "process", -1)

	if err != nil {
		if err == TaskAlreadyFinished {
			s.println("Message already finish")
			msg.Finish()
			return nil, TaskAlreadyFinished
		} else {
			s.println("Message is processing")
			msg.Touch()
			return nil, TaskAlreadyInWorkingState
		}

	}

	//if SetDiff succeed, then touch, publish, and add to  msgDispatched

	msg.Touch()

	pubStart := time.Now()
	s.publisher.Publish(msgID + ".process")

	pubDur := time.Since(pubStart)
	s.msgDispatched.Set(msgID, msg, -1)
	resp := &pb.TaskResponse{Payload: msg.GetPayload(), MessageID: []byte(msgID)}

	timeout := func() {
		err = s.msgStates.SetDiff(msgID, "failed", -1)
		if err != nil {
			goto exit
		}
		msg.Requeue(-1)
		s.msgDispatched.Delete(msgID)
	exit:
		s.timer.Delete(msgID)
		return

	}
	item := &TimerItem{
		Timeout:        timeout,
		ExpireDuration: time.Minute,
		ID:             msgID,
	}
	s.timer.Add(item)

	// pubDur := time.Since(start) - setDiffDur
	total := time.Since(start)
	if total > 1*time.Millisecond {
		fmt.Println("pub_duration: ", pubDur, " ", duationDivision(pubDur, total), " ", "total", " ", total)
	}
	return resp, nil

}

func (s *server) FinTask(ctx context.Context, in *pb.FinTaskRequest) (*pb.FinTaskResponse, error) {
	// start := time.Now()
	// defer func() {
	// 	s.println("FinTask Cost", time.Now().Sub(start))
	// }()
	// TODO: transaction between caches

	msgID := string(in.MessageID)
	msg := NewMessage(msgID, in.Payload, -1)

	switch in.Result {
	// If the result is success, take a radical approach
	case pb.TaskResult_FIN:
		err := s.msgStates.SetDiff(msgID, "finish", -1)
		if err != nil {
			s.println("Message already finish : ", err)
			break
		}

		//if SetDiff succeed, then publish, try finish msg, set the result
		s.publisher.Publish(msgID + ".finish")

		//confirm the message if this dispatcher dispatch this msg
		if val, err := s.msgDispatched.Get(msgID); err == nil {
			val.(Message).Finish()
			s.msgDispatched.Delete(msgID)
		}
		err = s.results.SetNX(msgID, msg, -1)
		if err != nil {
			s.println("result SetNX err: ", err)
			break
		}
		s.println("Finish Task, ID: ", msgID)

		break
	case pb.TaskResult_FAIL:
		s.println("Failed Task, ID: ", msgID)
		err := s.msgStates.SetDiff(msgID, "failed", -1)
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
		s.publisher.Publish(msgID + ".failed")

		if val, err := s.msgDispatched.Get(msgID); err == nil {
			val.(Message).Requeue(-1)
			s.msgDispatched.Delete(msgID)
		}

	}
	return &pb.FinTaskResponse{}, nil
}

func (s *server) getResponseLoop() {
	go func() {
		for {

			select {
			case msg := <-s.responseCh:
				msgID, recState := getMessageState(string(msg.Body))
				err := s.msgStates.SetDiff(msgID, recState, -1)
				if err != nil {
					if err == TaskAlreadyFinished {
						s.println("Message already finish")
						continue
					} else {
						s.println("Message is already in the state received")
						continue
					}

				}
				//if SetDiff succeed
				switch recState {

				case "finish":
					//try to finish msg
					if val, err := s.msgDispatched.Get(msgID); err == nil {
						val.(Message).Finish()
						s.msgDispatched.Delete(msgID)
					}
					continue
				case "failed":
					//try to reque msg
					if val, err := s.msgDispatched.Get(msgID); err == nil {
						val.(Message).Requeue(-1)
						s.msgDispatched.Delete(msgID)
					}
				case "process":
					//do nothing
					continue

				}

			}
		}

	}()

}

func NewServer(serverName string, nsqlookups []string, nsqlookupsD []string, nsqdAddr string, fetcherBuf int, nsqMaxInflight int, publisherProducerNum int, fetcherConsumerNum int, responserConsumerNum int, batchSize int) pb.DispatcherServer {

	s := &server{
		serverName:     serverName,
		fetchers:       NewInMemoryCache(),
		results:        NewInMemoryCache(),
		msgStates:      NewInMemoryCache(),
		msgDispatched:  NewInMemoryCache(),
		nsqlookups:     nsqlookups,
		nsqdAddr:       nsqdAddr,
		fetcherBuf:     fetcherBuf,
		nsqMaxInflight: nsqMaxInflight,
		consumerNum:    fetcherConsumerNum,
		responseCh:     make(chan *nsq.Message),
		verbose:        false,
		batchSize:      batchSize,
		nsqlookupsD:    nsqlookupsD,
		timer:          NewTimingWheel(),
	}
	// nsqConfig := nsq.NewConfig()
	// producer, _ := nsq.NewProducer(s.nsqdAddr, nsqConfig)

	s.publisher = NewPublisher("sync", s.nsqdAddr, s.batchSize)
	s.responser = NewResponser("sync", s.serverName, s.nsqlookupsD, s.responseCh)
	s.publisher.Start(publisherProducerNum)
	err := s.responser.Start(1, s.nsqMaxInflight)
	if err != nil {
		s.println("start responser err", err)
	}
	s.getResponseLoop()

	return s

}
