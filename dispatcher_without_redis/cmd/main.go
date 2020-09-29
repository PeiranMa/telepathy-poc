package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"strings"

	"google.golang.org/grpc"
	pb "poc.dispatcher/protos"
	"poc.dispatcher/server"
)

var (
	servername           = flag.String("servername", "test", "the name of server")
	lookupds             = flag.String("lookupd", "localhost:4161", "lookupd address, `<addr>:<port> <addr>:<port>`")
	nsqdAddr             = flag.String("nsqdaddr", "localhost:4150", "nsqd address fo produer")
	port                 = flag.String("p", "50051", "svc port")
	fetcherBuf           = flag.Int("b", 10000, "fetcher buffer size")
	nsqMaxInflight       = flag.Int("maxinflight", 50000, "nsq inflight message size")
	consumerNum          = flag.Int("cn", 1, "num of consumers")
	publisherProducerNum = flag.Int("pn", 1, "num of producer num")
	batchSize            = flag.Int("batch-size", 1000, "batch size of producer")
	lookupdsD            = flag.String("lookupd-d", "localhost:4161", "lookupd address for dispatcher, `<addr>:<port> <addr>:<port>`")
)

func main() {
	runtime.SetBlockProfileRate(1)
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()
	flag.Parse()
	grpcServer := grpc.NewServer()
	//NewServer(serverName string, nsqlookups []string, nsqdAddr string, fetcherBuf int, nsqMaxInflight int, publisherProducerNum int, fetcherConsumerNum int, responserConsumerNum int) pb.DispatcherServer {

	pb.RegisterDispatcherServer(grpcServer, server.NewServer(*servername, strings.Split(*lookupds, " "), strings.Split(*lookupdsD, " "), *nsqdAddr, *fetcherBuf, *nsqMaxInflight, *publisherProducerNum, *consumerNum, *consumerNum, *batchSize))
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Start Service")
	grpcServer.Serve(lis)
}
