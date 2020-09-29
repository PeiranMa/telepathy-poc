package main

import (
	// "context"
	// "errors"
	"flag"
	"fmt"
	// "github.com/go-redis/redis/v8"
	// "strings"
	// "os"
	// "reflect"
	// "runtime"
	// "reflect"
	"sync"
	"time"
	// // "sync/atomic"
	// "github.com/nsqio/go-nsq"
	// "google.golang.org/grpc"
	// pb "poc.dispatcher/protos"
	// "time"
	// "github.com/sigurn/crc16"
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

	// opt := &redis.Options{
	// 	Addr:     "localhost:6379",
	// 	Password: "",
	// 	DB:       0,
	// }
	// rdb := redis.NewClient(opt)

	// opt := &redis.ClusterOptions{
	// 	Addrs:    []string{"localhost:7000"},
	// 	Password: "",
	// }
	// rdb := redis.NewClusterClient(opt)
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
	// pubsub := rdb.Subscribe(ctx, "")

	// redisScript := redis.NewScript(`
	// return redis.call("mget", KEYS[1],KEYS[2])
	// `)
	// n, err := redisScript.Run(context.Background(), rdb, []string{"{yanis}ma", "{yanis}mapeiran"}).Result()
	// fmt.Println(n, err)
	// x := "am.pei.ran"
	// xSplit := strings.Split(x, ".")
	// fmt.Println(strings.Join(xSplit[0:2], "."))
	// fmt.Println(xSplit[2])
	// fmt.Println(xSplit)
	// fmt.Println(len(xSplit), cap(xSplit))
	// nsqConfig := nsq.NewConfig()
	// producer, err := nsq.NewProducer("localhost:4150", nsqConfig)
	// if err != nil {
	// 	fmt.Println("Can't create producer")

	// }
	// err = producer.Publish("test", []byte("hahahah"))
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// a := time.Now()
	// time.Sleep(5 * time.Second)
	// b := time.Since(a)
	// fmt.Println(float64(b.Nanoseconds()) / float64((50 * time.Second).Nanoseconds()))
	// var batch [][]byte
	// for i := 0; i < 100; i++ {
	// 	batch = append(batch, []byte("asd"))

	// }
	// fmt.Println(batch)
	test := make(map[string]string)
	test["a"] = "a"
	test["b"] = "b"
	for k, v := range test {
		fmt.Println(k, " ", v)
	}
}

// func sub(key string) {
// 	pubsub := rdb.Subscribe(ctx, "mychannel1")

// 	// Wait for confirmation that subscription is created before publishing anything.
// 	_, err := pubsub.Receive(ctx)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Go channel which receives messages.
// 	ch := pubsub.Channel()

// 	// Publish a message.
// 	err = rdb.Publish(ctx, "mychannel1", "hello").Err()
// 	if err != nil {
// 		panic(err)
// 	}

// 	time.AfterFunc(time.Second, func() {
// 		// When pubsub is closed channel is closed too.
// 		_ = pubsub.Close()
// 	})

// 	// Consume messages.
// 	for msg := range ch {
// 		fmt.Println(msg.Channel, msg.Payload)
// 	}
// }
