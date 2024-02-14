package main

import (
	"fmt"
	"net"

	"github.com/sifer169966/go-grpc-client-lb/server/apis/pb"
	"github.com/sifer169966/go-grpc-client-lb/server/config"
	"github.com/sifer169966/go-grpc-client-lb/server/handler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	err := config.Init()
	if err != nil {
		panic(err)
	}
	serverHost := fmt.Sprintf("%s:%s", config.Get().App.GRPCHost, config.Get().App.GRPCPort)
	lis, err := net.Listen("tcp4", serverHost)
	if err != nil {
		panic(err)
	}
	hdl := handler.NewGRPC()
	var opts []grpc.ServerOption
	srv := grpc.NewServer(opts...)
	defer srv.Stop()
	hs := health.NewServer()                     // will default to respond with SERVING
	grpc_health_v1.RegisterHealthServer(srv, hs) // registration
	pb.RegisterDeviceInteractionServiceServer(srv, hdl)
	err = srv.Serve(lis)
	if err != nil {
		panic(err)
	}
}
