package callbacks

import (
	"context"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"k8s.io/klog/v2"
)

func New() xds.CallbackFuncs {
	return xds.CallbackFuncs{
		StreamOpenFunc: func(ctx context.Context, streamID int64, typeURL string) error {
			klog.V(4).InfoS("StreamOpen", "streamID", streamID, "type", typeURL)
			return nil
		},
		StreamClosedFunc: func(streamID int64, node *corev3.Node) {
			klog.V(4).InfoS("StreamClosed", "streamID", streamID)
		},
		DeltaStreamOpenFunc: func(ctx context.Context, streamID int64, typeURL string) error {
			klog.V(4).InfoS("DeltaStreamOpen", "streamID", streamID, "type", typeURL)
			return nil
		},
		DeltaStreamClosedFunc: func(streamID int64, node *corev3.Node) {
			klog.V(4).InfoS("DeltaStreamClosed", "streamID", streamID)
		},
		StreamRequestFunc: func(streamID int64, request *discoverygrpc.DiscoveryRequest) error {
			klog.V(4).InfoS("StreamRequest", "streamID", streamID, "request", request)
			return nil
		},
		StreamResponseFunc: func(ctx context.Context, streamID int64, request *discoverygrpc.DiscoveryRequest, response *discoverygrpc.DiscoveryResponse) {
			klog.V(4).InfoS("StreamResponse", "streamID", streamID, "resourceNames", request.ResourceNames, "response", response)
		},
	}
}
