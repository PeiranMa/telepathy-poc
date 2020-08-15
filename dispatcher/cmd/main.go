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
	lookupds         = flag.String("lookupd", "localhost:4161", "lookupd address, `<addr>:<port> <addr>:<port>`")
	port             = flag.String("p", "50051", "svc port")
	fetcherBuf       = flag.Int("b", 20000, "fetcher buffer size")
	nsqMaxInflight   = flag.Int("maxinflight", 20000, "nsq inflight message size")
	redisPoolSize    = flag.Int("poolsize", 0, "max num of connections to redis, default 0 for runtime.NumCPU * 10")
	minIdleConnsSize = flag.Int("idlesize", 0, "min num of idle connections to redis")
)

func main() {
	runtime.SetBlockProfileRate(1)
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()
	flag.Parse()
	grpcServer := grpc.NewServer()
	pb.RegisterDispatcherServer(grpcServer, server.NewServer(strings.Split(*lookupds, " "), *fetcherBuf, *nsqMaxInflight, *redisPoolSize, *minIdleConnsSize))
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Start Service")
	grpcServer.Serve(lis)
}
