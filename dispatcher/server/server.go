package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nsqio/go-nsq"
	pb "poc.dispatcher/protos"
)

var (

	// errors
	TaskAlreadyInWorkingState = errors.New("TaskAlreadyInWorkingState")
	TaskAlreadyFinished       = errors.New("TaskAlreadyFinished")
)

type server struct {
	pb.DispatcherServer
	fetchers     Cache
	finishedMsgs Cache
	failedMsgs   Cache
	msgTimer     Timer
	nsqlookups   []string
}

func getFetcherID(topic string, channel string) string {
	return topic + "." + channel
}

func (s *server) GetTask(ctx context.Context, in *pb.TaskRequest) (*pb.TaskResponse, error) {
	// start := time.Now()
	// defer func() {
	// 	fmt.Println("GetTask Cost", time.Now().Sub(start))
	// }()

	// Fetch msg from Fetcher
	// If exists, put it into `working_msg` cache.
	// Else: return empty error.
	fid := getFetcherID(in.Topic, in.Channel)
	// nsqConfig := nsq.NewConfig()
	// nsqConfig.MaxInFlight = 5000
	// nsqConfig.MsgTimeout = 1 * time.Minute
	// nsqConfig.MaxAttempts = 10
	// f, err := NewFetcher(1, 5000, in.Topic, in.Channel, s.nsqlookups, nsqConfig)
	// if err != nil {
	// 	fmt.Println("NewFetcher error", err)
	// 	return nil, err
	// }
	var fetcher interface{}

	if s.fetchers.Exists(fid) {
		f, err := s.fetchers.Get(fid)
		if err != nil {
			fmt.Println("Get Fetcher failed", err)
			return nil, err
		}
		fetcher = f
	} else {

		nsqConfig := nsq.NewConfig()
		nsqConfig.MaxInFlight = 7000
		nsqConfig.MsgTimeout = 1 * time.Minute
		nsqConfig.MaxAttempts = 10
		f, err := NewFetcher(1, 7000, in.Topic, in.Channel, s.nsqlookups, nsqConfig)
		if err != nil {
			fmt.Println("NewFetcher error", err)
			return nil, err
		}
		err = s.fetchers.SetNX(fid, f, 0)
		if err == nil {
			f.Start()
		}
		fetcher = f

	}

	// fetcher, err := s.fetchers.Get(fid)
	// if err != nil {
	// 	fmt.Println("Get Fetcher failed", err)
	// 	return nil, err
	// }
	// fetch the msg
	msg, err := fetcher.(Fetcher).Fetch()
	if err != nil {
		fmt.Println("Fetch msg Failed, reason: ", err)
		return nil, err
	}
	msgID := msg.GetID().String()

	//Use get msg state
	infinish, infailed := s.finishedMsgs.Exist(msgID)
	// The msg is in success state
	if infinish {
		fmt.Println("message: " + msg.GetID() + " is finished")
		// confirm the message directly
		msg.Finish()
		return nil, TaskAlreadyFinished
	}

	// The msg is in failed state, retry
	if infailed {
		fmt.Println("message: " + msg.GetID() + "is failed, let us retry")
		s.failedMsgs.Delete(string(msg.GetID()))
	}
	// refresh msg directly
	msg.Touch()

	// AddMsg to Timer

	tick := func() bool {
		infinish, infailed := s.finishedMsgs.Exist(msgID)
		if infinish {
			msg.Finish()
			fmt.Println("Tick Message, Found Finished", msgID)
			return false
		}
		if infailed {
			msg.Requeue(-1)
			fmt.Println("Tick Message, Found Failed", msgID)
			return false
		}
		msg.Touch()
		return true
	}
	timeout := func() {
		msg.Requeue(-1)
	}

	timerItem := &TimerItem{
		Tick:           tick,
		Timeout:        timeout,
		TickDuration:   time.Duration(time.Second),
		ExpireDuration: 10 * time.Second,
		ID:             msgID,
	}

	s.msgTimer.Add(timerItem)

	fmt.Println("message: " + msg.GetID() + " start working")
	resp := &pb.TaskResponse{Payload: msg.GetPayload(), MessageID: []byte(msg.GetID())}
	return resp, nil
}

func (s *server) FinTask(ctx context.Context, in *pb.FinTaskRequest) (*pb.FinTaskResponse, error) {
	// start := time.Now()
	// defer func() {
	// 	fmt.Println("FinTask Cost", time.Now().Sub(start))
	// }()
	// TODO: transaction between caches
	msgID := string(in.MessageID)
	msg := NewMessage(msgID, in.Payload, -1)
	switch in.Result {
	// If the result is success, take a radical approach
	case pb.TaskResult_FIN:
		err := s.finishedMsgs.SetAndDel(msgID, msg)
		if err != nil {
			return nil, err
		}
		// err := s.finishedMsgs.Set(msgID, msg, -1)
		// if err != nil {
		// 	fmt.Println("finishedMsgs.Set error: ", err)
		// 	return nil, err
		// }
		// err = s.failedMsgs.Delete(msgID)
		// if err != nil {
		// 	fmt.Println("failedMsgs.Set Delete: ", err)
		// 	return nil, err
		// }
		fmt.Println("Finish Task, ID: ", msgID)
		break
	case pb.TaskResult_FAIL:
		fmt.Println("Failed Task, ID: ", msgID)
		// Set the msg failed only if it is not in finished state
		// if s.finishedMsgs.Exists(msgID) == false {
		// 	err := s.failedMsgs.Set(msgID, msg, -1)
		// 	if err != nil {
		// 		fmt.Println("failedMsgs.Set error: ", err)
		// 		return nil, err
		// 	}
		// }
		err := s.finishedMsgs.ExistAndSet(msgID, msg)
		if err != nil {
			return nil, err
		}
		break
	}
	return &pb.FinTaskResponse{}, nil

}

func NewServer(nsqlookups []string) pb.DispatcherServer {

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"

	}
	redisPass := os.Getenv("REDIS_PASSWORD")
	fmt.Println("redisAddr", redisAddr, redisPass)

	// opt := &redis.ClusterOptions{
	// 	Addrs:    []string{redisAddr},
	// 	Password: redisPass, // no password set
	// }

	opt := &redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       0,
	}

	s := &server{
		fetchers:     NewInMemoryCache(),
		finishedMsgs: NewRedisCache("finish", opt),
		failedMsgs:   NewRedisCache("failed", opt),
		msgTimer:     NewTimingWheel(),
		nsqlookups:   nsqlookups,
	}
	return s

}
