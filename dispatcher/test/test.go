package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/go-redis/redis/v8"
	// "os"
	// "reflect"
	// "runtime"
	"sync"
	"time"
	// // "sync/atomic"
	// "github.com/nsqio/go-nsq"
	// "google.golang.org/grpc"
	// pb "poc.dispatcher/protos"
	// "time"
)

var (
	addr       = flag.String("s", "localhost:50051", "server address, `<addr>:<port>`")
	topic      = flag.String("t", "disp-topic", "working topic")
	channel    = flag.String("ch", "disp-channel", "working channel")
	nsqdAddr   = flag.String("nsqd", "localhost:4150", "lookupd address")
	duration   = flag.Duration("runfor", 30*time.Second, "duration of time to run, e.g `1h1m10s`, `10ms`")
	conn       = flag.Int("c", 10, "concurrent connections")
	numRPC     = flag.Int("r", 10, "The number of concurrent RPCs on each connection.")
	produceMsg = flag.Bool("p", false, "produceMsg mode, default false")

	wg     sync.WaitGroup
	cnts   = int32(0)
	fins   = int32(0)
	nomsgs = int32(0)
)

func main() {

	// fmt.Println(time.Now().UTC().UnixNano() / int64(time.Millisecond))

	opt := &redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	rdb := redis.NewClient(opt)
	// fmt.Println(opt.MinIdleConns)
	// testScript := redis.NewScript(`
	// if redis.call("hexists", "finish", KEYS[1]) == 1  then	return 0 end return redis.call("hset","finish",KEYS[1],ARGV[1])
	// `)
	// n, err := testScript.Run(context.Background(), rdb, []string{"234"}, "ma").Result()
	// fmt.Println(n, err)
	// testScript := redis.NewScript(`
	// local infinish = redis.call("hexists", "finish", KEYS[1])
	// local infailed = redis.call("hexists", "failed", KEYS[1])
	// return {infinish, infailed}
	// `)
	// n, err := testScript.Run(context.Background(), rdb, []string{"ma"}).Result()
	// tmp := n.([]interface{})
	// fmt.Println(tmp[0].(int64))
	// fmt.Println(tmp[1].(int64))
	// fmt.Println(tmp)
	// fmt.Println(reflect.TypeOf(n))
	// fmt.Println(n, err)
	// fmt.Println()
	// testScript := redis.NewScript(`
	// redis.call("hsetnx", "finish", KEYS[1], ARGV[1])
	// redis.call("hdel", "failed", KEYS[1])
	// return 1
	// `)
	// n, err := testScript.Run(context.Background(), rdb, []string{"mapei"}, "yanis").Result()
	// fmt.Println(n, err)

	// redisScript := redis.NewScript(`
	// if redis.call("hexists", "finish", KEYS[1]) == 0 then
	// redis.call("hsetnx", "failed", KEYS[1],ARGV[1])
	// end
	// return 1
	// `)
	// n, err := redisScript.Run(context.Background(), rdb, []string{"ma2"}, "yanis").Result()
	// fmt.Println(n, err)
	// // fmt.Println(runtime.NumCPU())
	// fmt.Println(time.Now().UTC().UnixNano() / int64(time.Millisecond))
	// keys := []string{}
	// redisScript := redis.NewScript(`
	// local batchResult = {}
	// for i = 1, table.getn(KEYS) do
	// local infinish = redis.call("hexists", "finish", KEYS[i])
	// local infailed = redis.call("hexists", "failed", KEYS[i])
	// table.insert(batchResult,{infinish, infailed})
	// end
	// return batchResult
	// `)
	// res, err := redisScript.Run(context.Background(), rdb, keys).Result()
	// if err != nil {
	// 	fmt.Println("Exist error : ", err)

	// }
	// keysSize := len(keys)
	// Result := make([][]int64, keysSize)
	// for i := range Result {
	// 	Result[i] = make([]int64, 2)
	// }
	// for i, va := range res.([]interface{}) {
	// 	Result[i][0] = va.([]interface{})[0].(int64)
	// 	Result[i][1] = va.([]interface{})[1].(int64)
	// }
	// for i := range Result {
	// 	fmt.Println(Result[i])
	// }

	// 	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	// 	if err != nil {
	// 		panic(err)

	// 	}

	// 	// // endTime := time.Now().Add(*duration)
	// 	client := pb.NewDispatcherClient(conn)
	// 	resp, err := client.GetTask(context.Background(), &pb.TaskRequest{Topic: *topic, Channel: *channel})
	// 	_, err = client.FinTask(context.Background(), &pb.FinTaskRequest{
	// 		Topic:   *topic,
	// 		Channel: *channel,
	// 		Result:  pb.TaskResult_FIN, MessageID: resp.MessageID})
	const maxRetries = 1000

	increment := func(key string) error {

		txf := func(tx *redis.Tx) error {

			n, err := tx.Get(context.Background(), key).Int()
			if err != nil && err != redis.Nil {
				return err
			}

			n++

			_, err = tx.TxPipelined(context.Background(), func(pipe redis.Pipeliner) error {
				pipe.Set(context.Background(), key, n, 0)
				return nil
			})
			return err
		}

		for i := 0; i < maxRetries; i++ {
			err := rdb.Watch(context.Background(), txf, key)
			if err == nil {

				return nil
			}
			if err == redis.TxFailedErr {

				continue
			}

			return err
		}

		return errors.New("increment reached maximum number of retries")
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := increment("counter3"); err != nil {
				fmt.Println("increment error:", err)
			}
		}()
	}
	wg.Wait()

	n, err := rdb.Get(context.Background(), "counter3").Int()
	fmt.Println("ended with", n, err)

}
