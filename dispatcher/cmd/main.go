package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"google.golang.org/grpc"

	pb "poc.dispatcher/protos"
	"poc.dispatcher/server"
)

var (
	lookupds = flag.String("lookupd", "localhost:4161", "lookupd address, `<addr>:<port> <addr>:<port>`")
	port     = flag.String("p", "50051", "svc port")
)

func main() {
	runtime.SetBlockProfileRate(1)
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	flag.Parse()
	grpcServer := grpc.NewServer()
	pb.RegisterDispatcherServer(grpcServer, server.NewServer())
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Start Service")
	grpcServer.Serve(lis)
}
