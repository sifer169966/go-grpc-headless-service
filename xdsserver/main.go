package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/sifer169966/go-grpc-client-lb/xdsserver/callbacks"
	"github.com/sifer169966/go-grpc-client-lb/xdsserver/k8sreflector"
	"github.com/sifer169966/go-grpc-client-lb/xdsserver/snapshots"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func main() {
	snap := snapshots.New()
	k8sClientConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), nil).ClientConfig()
	if err != nil {
		klog.Fatal("could not create k8s client configuration: ", err)
	}
	k8sClient, err := kubernetes.NewForConfig(k8sClientConfig)
	if err != nil {
		klog.Fatal("could not create k8s client: ", err)
	}

	stopCtx, stop := context.WithCancel(context.Background())
	refl := k8sreflector.New(snap, k8sClient)
	grpcSrv := grpc.NewServer()
	healthSrv := health.NewServer()
	xdsSrv := xds.NewServer(stopCtx, snap.MuxCache(), callbacks.New())
	go func() {
		err := refl.Watch(stopCtx)
		if err != nil {
			klog.Fatal(err)
		}
	}()
	lis, err := net.Listen("tcp4", ":9090")
	if err != nil {
		klog.Fatal(err)
	}
	grpc_health_v1.RegisterHealthServer(grpcSrv, healthSrv)
	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(grpcSrv, xdsSrv)
	go func() {
		err = grpcSrv.Serve(lis)
		if err != nil {
			klog.Fatal(err)
		}
	}()

	klog.Infoln("server is ready")
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	<-sigchan
	klog.Infoln("server is shutting down...")
	stop()
	healthSrv.Shutdown()
	grpcSrv.GracefulStop()
	lis.Close()
	klog.Infoln("server was gracefully stopped")

}
