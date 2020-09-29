package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/bmizerany/perks/quantile"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	pb "poc.dispatcher/protos"
)

var (
	addr        = flag.String("s", "localhost:50051", "server address, `<addr>:<port>`")
	topic       = flag.String("t", "disp-topic", "working topic")
	channel     = flag.String("ch", "disp-channel", "working channel")
	duration    = flag.Duration("runfor", 30*time.Second, "duration of time to run, e.g `1h1m10s`, `10ms`")
	conn        = flag.Int("c", 10, "concurrent connections")
	numRPC      = flag.Int("r", 10, "number of concurrent RPCs on each connection.")
	produceMsg  = flag.Bool("p", false, "produceMsg mode, default false")
	pulsarProxy = flag.String("addr", "pulsar-proxy", "address of pulsar-proxy")
	numTopic    = flag.Int("nt", 1, "number of topics")

	wg     sync.WaitGroup
	cnts   = int32(0)
	fins   = int32(0)
	nomsgs = int32(0)
)

func runWithConn() {
	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		panic(err)

	}

	endTime := time.Now().Add(*duration)

	for i := 0; i < *numRPC; i++ {
		client := pb.NewDispatcherClient(conn)
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cnt := int32(0)
			fin := int32(0)
			nomsg := int32(0)
			for {
			RetryGet:
				if time.Now().After(endTime) {
					atomic.AddInt32(&cnts, cnt)
					atomic.AddInt32(&fins, fin)
					atomic.AddInt32(&nomsgs, nomsg)
					break
				}
				// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				// defer cancel()
				t := time.Now()
				multi_topic := *topic + "-" + strconv.Itoa(int(t.Unix())%*numTopic)
				resp, err := client.GetTask(context.Background(), &pb.TaskRequest{Topic: multi_topic, Channel: *channel + strconv.Itoa(i)})
				fmt.Println("GetTask cost", time.Since(t))
				if err != nil {
					fmt.Println("GetTask failed, error: ", err)
					nomsg++
					//time.Sleep(time.Second)
					goto RetryGet
				}
				cnt++
				// fmt.Printf("Get task %s, cnt %d\n", resp.MessageID, cnt)
			RetryFinTask:
				// ctx1, cancel1 := context.WithTimeout(context.Background(), 1000*time.Second)
				// defer cancel1()
				// time.Sleep(1 * time.Second)
				t = time.Now()
				_, err = client.FinTask(context.Background(), &pb.FinTaskRequest{
					Topic:     multi_topic,
					Channel:   *channel,
					Result:    pb.TaskResult_FIN,
					MessageID: resp.MessageID})
				fmt.Println("FinTask cost", time.Since(t))
				if err != nil {
					fmt.Println("FinTask failed, error: ", err)
					//time.Sleep(time.Millisecond)
					goto RetryFinTask
				}
				fin++
				//fmt.Printf("Fin task %s, cnt %d\n", resp.MessageID, fin)
			}

		}(i)
	}
}

func workflow() {

	start := time.Now()
	for i := 0; i < *conn; i++ {
		runWithConn()
	}
	wg.Wait()

	milli := time.Now().Sub(start).Seconds()
	fmt.Printf("Workflow Duration: %.2f sec, Get Task %d, Finish Task %d, NoMsg %v, QPS: %.2f \n", milli, cnts, fins, nomsgs, float32(fins)/float32(milli))
	forever := make(chan int)
	<-forever
}

func produceWork() {

	client, err := NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	var producers []pulsar.Producer
	for i := 0; i < *numTopic; i++ {
		producer, err := client.CreateProducer(pulsar.ProducerOptions{
			Topic:                   *topic + "-" + strconv.Itoa(i),
			MaxPendingMessages:      1000,
			BatchingMaxPublishDelay: time.Millisecond * time.Duration(1),
			BatchingMaxSize:         uint(128 * 1024),
			// DisableBatching: true,
		})
		if err != nil {
			log.Fatal(err)
		}
		producers = append(producers, producer)
		defer producer.Close()
	}
	ctx := context.Background()

	payload := make([]byte, 100)

	ch := make(chan float64)

	go func(stopCh <-chan time.Time) {

		for {
			select {
			case <-stopCh:
				return
			default:
			}

			start := time.Now()
			producers[int(start.Unix())%*numTopic].SendAsync(ctx, &pulsar.ProducerMessage{
				Payload: payload,
			}, func(msgID pulsar.MessageID, message *pulsar.ProducerMessage, e error) {
				if e != nil {
					log.WithError(e).Fatal("Failed to publish")
				}

				latency := time.Since(start).Seconds()
				ch <- latency
			})
		}
	}(time.NewTimer(*duration).C)

	// Print stats of the publish rate and latencies
	tick := time.NewTicker(10 * time.Second)
	q := quantile.NewTargeted(0.50, 0.95, 0.99, 0.999, 1.0)
	messagesPublished := 0
	totalMessagesPublished := 0
	stop := time.NewTimer(*duration).C

	for {
		select {
		case <-stop:
			goto exit
		case <-tick.C:
			messageRate := float64(messagesPublished) / float64(10)
			log.Infof(`Stats - Publish rate: %6.1f msg/s - %6.1f Mbps - 
				Latency ms: 50%% %5.1f -95%% %5.1f - 99%% %5.1f - 99.9%% %5.1f - max %6.1f`,
				messageRate,
				messageRate*float64(100)/1024/1024*8,
				q.Query(0.5)*1000,
				q.Query(0.95)*1000,
				q.Query(0.99)*1000,
				q.Query(0.999)*1000,
				q.Query(1.0)*1000,
			)

			q.Reset()
			messagesPublished = 0
		case latency := <-ch:
			messagesPublished++
			totalMessagesPublished++
			q.Insert(latency)
		}
	}
exit:
	log.Infof(`Stats - Publish Total Number : %d msgs`, totalMessagesPublished)
	forever := make(chan int)
	<-forever
}

func NewClient() (pulsar.Client, error) {
	clientOpts := pulsar.ClientOptions{
		URL: "pulsar://" + *pulsarProxy + ":6650",
	}
	return pulsar.NewClient(clientOpts)
}

func main() {
	flag.Parse()
	if *produceMsg {
		produceWork()
	} else {
		workflow()
	}

}
