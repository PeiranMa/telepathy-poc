package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/go-redis/redis/v8"
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
	// failedMsgs     Cache
	msgTimer   *twBatchTimer
	nsqlookups []string

	fetcherBuf     int
	nsqMaxInflight int
	consumerNum    int
	tickCh         chan Message
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
	fid := getFetcherID(in.Topic, in.Channel)
	var fetcher interface{}
	if s.fetchers.Exists(fid) {
		f, err := s.fetchers.Get(fid)
		if err != nil {
			fmt.Println("Get Fetcher failed", err)
			return nil, err
		}
		fetcher = f
	} else {
		f := NewFetcher(s.fetcherBuf, in.Topic, in.Channel, s.nsqlookups)
		err := s.fetchers.SetNX(fid, f, 0)
		if err == nil {
			err = f.Start(s.consumerNum, s.nsqMaxInflight)
			if err != nil {
				fmt.Println("Fetcher Start error", err)
				return nil, err
			}
		}
		fetcher = f

	}

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
		// fmt.Println("message: " + msg.GetID() + " is finished")
		// confirm the message directly
		msg.Finish()
		return nil, TaskAlreadyFinished
	}

	// The msg is in failed state, retry
	if infailed {
		// fmt.Println("message: " + msg.GetID() + "is failed, let us retry")
		s.finishedMsgs.Delete(string(msg.GetID()))
	}

	// refresh msg directly
	msg.Touch()

	//tick add  message to the tick channel
	tick := func() bool {
		s.tickCh <- msg
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

	// fmt.Println("message: " + msg.GetID() + " start working")
	resp := &pb.TaskResponse{Payload: msg.GetPayload(), MessageID: []byte(msg.GetID())}
	return resp, nil
}

//get messages from tick chan and process them with batch size
func (s *server) tickLoop(workerNum int, batchSize int) {
	for i := 0; i < workerNum; i++ {
		go func() {
			batchMsgs := make(chan Message, batchSize)
			for {
				select {
				//take message from the tick channel
				case msg := <-s.tickCh:
					select {
					//if batchMsgs chan isn't full, add the message to this chan
					case batchMsgs <- msg:
						continue
					//if batchMsgs chan is full, process messages, and then add the message to this chan
					default:
						s.batchProcess(batchMsgs, batchSize, msg)
					}
				//if can't get a message from the tickCh for a second, process the messages in batchMsgs
				case <-time.After(time.Second):
					//if the batchMsgs is empty, then there could be no more messages, block until get a new message
					if len(batchMsgs) == 0 {
						select {
						case msg := <-s.tickCh:
							batchMsgs <- msg
							continue
						}
					}
					//process the rest messages in the batchMsgs
					s.batchProcess(batchMsgs, len(batchMsgs), nil)
				}

			}
		}()
	}

}

func (s *server) batchProcess(batchMsgs chan Message, batchSize int, msg Message) {
	msgMap := make(map[string]Message)
	msgIDs := make([]string, batchSize)
	for i := 0; i < batchSize; i++ {
		tmpMsg := <-batchMsgs
		msgIDs[i] = tmpMsg.GetID().String()
		msgMap[msgIDs[i]] = tmpMsg
	}
	//add the msg to batchMsgs for the next iteration if message != nil
	if msg != nil {
		batchMsgs <- msg
	}
	//query the redis, res : [batchSize][infinish(bool), infailed(bool)]
	res := s.finishedMsgs.BatchExist(msgIDs)

	for i := 0; i < batchSize; i++ {
		msgID := msgIDs[i]
		batchMsg := msgMap[msgID]
		timerItem := s.msgTimer.GetItem(msgID)
		//if the message in finish, confirm and delete from timer
		if res[i][0] {
			batchMsg.Finish()
			// fmt.Println("Tick Message, Found Finished", msgID)
			s.msgTimer.Delete(msgID)
			continue
		}
		//if the message in failed, requeue and delete from timer
		if res[i][1] {
			batchMsg.Requeue(-1)
			// fmt.Println("Tick Message, Found Failed", msgID)
			s.msgTimer.Delete(msgID)
			continue
		}
		//check if the message is timeout
		if time.Now().After(timerItem.StartTime.Add(timerItem.ExpireDuration)) {
			batchMsg.Requeue(-1)
			// fmt.Println("Tick Message, Found Timeout", msgID)
			s.msgTimer.Delete(msgID)
			continue
		}

		// fmt.Println("Tick Message, Need Tick Again")
		batchMsg.Touch()
		// tick again
		s.msgTimer.SetTimer(msgID, s.msgTimer.AfterFunc(timerItem.TickDuration, func() {
			s.tickCh <- batchMsg
		}))

	}

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
		// fmt.Println("Finish Task, ID: ", msgID)
		break
	case pb.TaskResult_FAIL:
		// fmt.Println("Failed Task, ID: ", msgID)
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

func NewServer(nsqlookups []string, fetcherBuf int, nsqMaxInflight int, poolSize int, minIdleConns int, workerNum int, batchNum int, consumerNum int) pb.DispatcherServer {

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"

	}
	redisPass := os.Getenv("REDIS_PASSWORD")
	fmt.Println("redisAddr", redisAddr, redisPass)
	fmt.Println()

	if poolSize == 0 {
		poolSize = runtime.NumCPU() * 10
	}

	// opt := &redis.ClusterOptions{
	// 	Addrs:    []string{redisAddr},
	// 	Password: redisPass, // no password set
	// }

	opt := &redis.Options{
		Addr:         redisAddr,
		Password:     redisPass,
		DB:           0,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
	}

	s := &server{
		fetchers:     NewInMemoryCache(),
		finishedMsgs: NewRedisCache("finish", opt),
		// failedMsgs:     NewRedisCache("failed", opt),
		// msgTimer:       NewTimingWheel(),
		msgTimer:       NewTimingWheelBatch(),
		nsqlookups:     nsqlookups,
		fetcherBuf:     fetcherBuf,
		nsqMaxInflight: nsqMaxInflight,
		tickCh:         make(chan Message),
		consumerNum:    consumerNum,
	}
	s.tickLoop(workerNum, batchNum)
	return s

}
