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

func (r *Reflector) watchServices(ctx context.Context) error {
	var refl *k8scache.Reflector
	store := k8scache.NewUndeltaStore(r.servicesPushFunc(ctx, refl), k8scache.DeletionHandlingMetaNamespaceKeyFunc)
	refl = k8scache.NewReflector(&k8scache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return r.k8sClient.CoreV1().Services("").List(ctx, options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return r.k8sClient.CoreV1().Services("").Watch(ctx, options)
		},
	}, &corev1.Service{}, store, r.resyncPeriod)

	refl.Run(ctx.Done())
	return nil
}

func (r *Reflector) servicesPushFunc(ctx context.Context, refl *k8scache.Reflector) func(items []interface{}) {
	return func(services []interface{}) {
		if refl == nil {
			klog.Warning("reflector is not ready yet")
			return
		}
		latestVersion := refl.LastSyncResourceVersion()
		svcs := sliceToServices(services)
		res := servicesToResources(svcs)
		r.snap.Set(ctx, latestVersion, res)
	}
}

// servicesToResources ...
// creating lds, rds, and cds resources from k8s services
func servicesToResources(svcs []*corev1.Service) []types.Resource {
	var out []types.Resource
	for _, svc := range svcs {
		serviceToResources(out, svc)
	}
	return out
}

// serviceToResources ...
// creating lds, rds, and cds resources from k8s service
func serviceToResources(out []types.Resource, svc *corev1.Service) {
	router, _ := anypb.New(&routerv3.Router{})
	serviceFullName := fmt.Sprintf("%s.%s", svc.Name, svc.Namespace)
	for _, port := range svc.Spec.Ports {
		targetHostPort := net.JoinHostPort(serviceFullName, port.Name)
		targetHostPortNumber := net.JoinHostPort(serviceFullName, strconv.Itoa(int(port.Port)))
		routeConfig := &routev3.RouteConfiguration{
			Name: targetHostPortNumber,
			VirtualHosts: []*routev3.VirtualHost{
				{
					Name:    targetHostPort,
					Domains: []string{serviceFullName, targetHostPort, targetHostPortNumber, svc.Name},
					Routes: []*routev3.Route{{
						Name: "default",
						Match: &routev3.RouteMatch{
							PathSpecifier: &routev3.RouteMatch_Prefix{},
						},
						Action: &routev3.Route_Route{
							Route: &routev3.RouteAction{
								ClusterSpecifier: &routev3.RouteAction_Cluster{
									Cluster: targetHostPort,
								},
							},
						},
					}},
				},
			},
		}

		manager, _ := anypb.New(&managerv3.HttpConnectionManager{
			HttpFilters: []*managerv3.HttpFilter{
				{
					Name: wellknown.Router,
					ConfigType: &managerv3.HttpFilter_TypedConfig{
						TypedConfig: router,
					},
				},
			},
			RouteSpecifier: &managerv3.HttpConnectionManager_RouteConfig{
				RouteConfig: routeConfig,
			},
		})

		svcListener := &listenerv3.Listener{
			Name: targetHostPortNumber,
			ApiListener: &listenerv3.ApiListener{
				ApiListener: manager,
			},
		}

		svcCluster := &clusterv3.Cluster{
			Name:                 targetHostPort,
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
		out = append(out, svcListener, routeConfig, svcCluster)
	}
}

func sliceToServices(services []interface{}) []*corev1.Service {
	out := make([]*corev1.Service, len(services))
	for i, v := range services {
		out[i] = v.(*corev1.Service)
	}
	return out
}
