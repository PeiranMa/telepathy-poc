package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	// "strings"

	"google.golang.org/grpc"
	pb "poc.dispatcher/protos"
	"poc.dispatcher/server"
)

var (
	port                 = flag.String("p", "50051", "svc port")
	publisherProducerNum = flag.Int("pn", 1, "num of producer num")
	fetcherConsumerNum   = flag.Int("cn", 1, "num of consumer num per fetcher")
	pulsarProxy          = flag.String("addr", "pulsar-proxy", "address of pulsar-proxy")
)

func main() {
	runtime.SetBlockProfileRate(1)
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()
	flag.Parse()
	grpcServer := grpc.NewServer()
	//NewServer(serverName string, nsqlookups []string, nsqdAddr string, fetcherBuf int, nsqMaxInflight int, publisherProducerNum int, fetcherConsumerNum int, responserConsumerNum int) pb.DispatcherServer {

	pb.RegisterDispatcherServer(grpcServer, server.NewServer(*pulsarProxy, *publisherProducerNum, *fetcherConsumerNum))
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Start Service")
	grpcServer.Serve(lis)
}
