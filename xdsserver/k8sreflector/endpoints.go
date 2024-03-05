package k8sreflector

import (
	"context"
	"fmt"
	"sort"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"google.golang.org/protobuf/types/known/wrapperspb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	k8scache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func (r *Reflector) watchEndpoints(ctx context.Context) error {
	store := k8scache.NewUndeltaStore(r.endpointsPushFunc(ctx), k8scache.DeletionHandlingMetaNamespaceKeyFunc)
	r.ep.k8sRefl = k8scache.NewReflector(&k8scache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return r.k8sClient.CoreV1().Endpoints("").List(ctx, options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return r.k8sClient.CoreV1().Endpoints("").Watch(ctx, options)
		},
	}, &corev1.Endpoints{}, store, r.ep.resyncPeriod)
	r.ep.k8sRefl.Run(ctx.Done())
	klog.Warning("endpoints reflector has been stopped")
	return nil
}

func (r *Reflector) endpointsPushFunc(ctx context.Context) func(v []interface{}) {
	return func(v []interface{}) {
		if r.ep.k8sRefl == nil {
			klog.Warning("reflector is not ready yet")
			return
		}
		latestVersion := r.ep.k8sRefl.LastSyncResourceVersion()
		eps := sliceToEndpoints(v)
		res := endpointsToResources(eps)
		r.snap.Set(ctx, latestVersion, res)
		klog.Infof("set snapshot version to %s from endpointsPushFunc()", latestVersion)
	}
}

// endpointsToResources ...
// creating eds resources from k8s endpoints
func endpointsToResources(eps []*corev1.Endpoints) []types.Resource {
	var out []types.Resource
	for _, ep := range eps {
		endpointToOutResources(&out, ep)
	}
	return out
}

// endpointToOutResources ...
// push eds resources from k8s endpoint to `out`
func endpointToOutResources(out *[]types.Resource, ep *corev1.Endpoints) {
	klog.Infof("\ndebug subsets: %+v\n", ep.Subsets)
	for _, subset := range ep.Subsets {
		klog.Infof("\ndebug subset.port: %+v\n", subset.Ports)
		for _, port := range subset.Ports {
			var claName string
			if port.Name == "" {
				claName = fmt.Sprintf("%s.%s:%d", ep.Name, ep.Namespace, port.Port)
			} else {
				claName = fmt.Sprintf("%s.%s:%s", ep.Name, ep.Namespace, port.Name)
			}
			cla := &endpointv3.ClusterLoadAssignment{
				ClusterName: claName,
				Endpoints: []*endpointv3.LocalityLbEndpoints{
					{
						LoadBalancingWeight: wrapperspb.UInt32(1),
						Locality:            &corev3.Locality{},
						LbEndpoints:         []*endpointv3.LbEndpoint{},
					},
				},
			}
			appendLbEndpoints(subset.Addresses, cla, port)
			*out = append(*out, cla)
		}
	}
}

func appendLbEndpoints(addrs []corev1.EndpointAddress, cla *endpointv3.ClusterLoadAssignment, port corev1.EndpointPort) {
	sort.SliceStable(addrs, func(i, j int) bool {
		left := addrs[i].IP
		right := addrs[j].IP
		return left < right
	})
	klog.Infof("debug address: %v", addrs)
	for _, addr := range addrs {
		hostname := addr.Hostname
		if hostname == "" && addr.TargetRef != nil {
			hostname = fmt.Sprintf("%s.%s", addr.TargetRef.Name, addr.TargetRef.Namespace)
		}
		if hostname == "" && addr.NodeName != nil {
			hostname = *addr.NodeName
		}
		cla.Endpoints[0].LbEndpoints = append(cla.Endpoints[0].LbEndpoints, &endpointv3.LbEndpoint{
			HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
				Endpoint: &endpointv3.Endpoint{
					Address: &corev3.Address{
						Address: &corev3.Address_SocketAddress{
							SocketAddress: &corev3.SocketAddress{
								Protocol: corev3.SocketAddress_TCP,
								Address:  addr.IP,
								PortSpecifier: &corev3.SocketAddress_PortValue{
									PortValue: uint32(port.Port),
								},
							},
						},
					},
					Hostname: hostname,
				},
			},
		})
	}
}

func sliceToEndpoints(eps []interface{}) []*corev1.Endpoints {
	out := make([]*corev1.Endpoints, len(eps))
	for i, ep := range eps {
		out[i] = ep.(*corev1.Endpoints)
	}
	return out
}
