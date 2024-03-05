package k8sreflector

import (
	"context"
	"fmt"
	"net"
	"strconv"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	routerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	managerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	k8scache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

type servicesReflector struct {
	k8sRefl *k8scache.Reflector

	// not thread-safe
	lastSnapshotSum uint64
}

func (r *Reflector) watchServices(ctx context.Context) error {
	store := k8scache.NewUndeltaStore(r.servicesPushFunc(ctx), k8scache.DeletionHandlingMetaNamespaceKeyFunc)
	r.svc.k8sRefl = k8scache.NewReflector(&k8scache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return r.k8sClient.CoreV1().Services("").List(ctx, options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return r.k8sClient.CoreV1().Services("").Watch(ctx, options)
		},
	}, &corev1.Service{}, store, r.svc.resyncPeriod)
	r.svc.k8sRefl.Run(ctx.Done())
	klog.Warning("services reflector has been stopped")
	return nil
}

func (r *Reflector) servicesPushFunc(ctx context.Context) func(v []interface{}) {
	return func(v []interface{}) {
		if r.svc.k8sRefl == nil {
			klog.Warning("reflector is not ready yet")
			return
		}
		latestVersion := r.svc.k8sRefl.LastSyncResourceVersion()
		svcs := sliceToServices(v)
		res := servicesToResources(svcs)
		r.snap.Set(ctx, latestVersion, res)
		klog.Infof("set snapshot version to %s from servicesPushFunc()", latestVersion)
	}
}

// servicesToResources ...
// creating lds, rds, and cds resources from k8s services
func servicesToResources(svcs []*corev1.Service) []types.Resource {
	var out []types.Resource
	klog.Infof("\ndebug services: %+v\n", svcs)
	for _, svc := range svcs {
		serviceToOutResources(&out, svc)
	}
	return out
}

// serviceToOutResources ...
// push lds, rds, and cds resources from k8s service to `out`
func serviceToOutResources(out *[]types.Resource, svc *corev1.Service) {
	router, _ := anypb.New(&routerv3.Router{})
	host := fmt.Sprintf("%s.%s", svc.Name, svc.Namespace)
	for _, port := range svc.Spec.Ports {
		hostWithPortName := net.JoinHostPort(host, port.Name)
		hostWithPortNumber := net.JoinHostPort(host, strconv.Itoa(int(port.Port)))
		rds := &routev3.RouteConfiguration{
			Name: hostWithPortNumber,
			VirtualHosts: []*routev3.VirtualHost{
				{
					Name:    hostWithPortName,
					Domains: []string{host, hostWithPortName, hostWithPortNumber, svc.Name},
					Routes: []*routev3.Route{{
						Name: "default",
						Match: &routev3.RouteMatch{
							PathSpecifier: &routev3.RouteMatch_Prefix{},
						},
						Action: &routev3.Route_Route{
							Route: &routev3.RouteAction{
								ClusterSpecifier: &routev3.RouteAction_Cluster{
									Cluster: hostWithPortName,
								},
							},
						},
					}},
				},
			},
		}

		hcm, _ := anypb.New(&managerv3.HttpConnectionManager{
			HttpFilters: []*managerv3.HttpFilter{
				{
					Name: wellknown.Router,
					ConfigType: &managerv3.HttpFilter_TypedConfig{
						TypedConfig: router,
					},
				},
			},
			RouteSpecifier: &managerv3.HttpConnectionManager_RouteConfig{
				RouteConfig: rds,
			},
		})

		lds := &listenerv3.Listener{
			Name: hostWithPortNumber,
			ApiListener: &listenerv3.ApiListener{
				ApiListener: hcm,
			},
		}

		cds := &clusterv3.Cluster{
			Name:                 hostWithPortName,
			ClusterDiscoveryType: &clusterv3.Cluster_Type{Type: clusterv3.Cluster_EDS},
			LbPolicy:             clusterv3.Cluster_ROUND_ROBIN,
			EdsClusterConfig: &clusterv3.Cluster_EdsClusterConfig{
				EdsConfig: &corev3.ConfigSource{
					ConfigSourceSpecifier: &corev3.ConfigSource_Ads{
						Ads: &corev3.AggregatedConfigSource{},
					},
				},
			},
		}
		*out = append(*out, lds, rds, cds)
	}
}

func sliceToServices(svcs []interface{}) []*corev1.Service {
	out := make([]*corev1.Service, len(svcs))
	for i, svc := range svcs {
		out[i] = svc.(*corev1.Service)
	}
	return out
}
