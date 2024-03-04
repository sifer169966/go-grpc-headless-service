package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/sifer169966/go-grpc-client-lb/xdsserver/k8sreflector"
	"github.com/sifer169966/go-grpc-client-lb/xdsserver/snapshots"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"k8s.io/klog/v2"
)

func main() {
	snap := snapshots.New()
	refl := k8sreflector.New(snap)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		err := refl.Watch(ctx)
		if err != nil {
			klog.Fatal(err)
		}
	}()
	grpcSrv := grpc.NewServer()
	lis, err := net.Listen("tcp4", ":9090")
	if err != nil {
		klog.Fatal(err)
	}
	go func() {
		err = grpcSrv.Serve(lis)
		if err != nil {
			klog.Fatal(err)
		}
	}()
	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcSrv, healthSrv)
	klog.Infoln("server is ready")

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	<-sigchan

	klog.Infoln("server is shutting down...")
	cancel()
	healthSrv.Shutdown()
	grpcSrv.GracefulStop()
	lis.Close()
	klog.Infoln("server was gracefully stopped")

}
