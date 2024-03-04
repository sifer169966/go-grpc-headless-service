package resources

import (
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	authv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	runtimev3 "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

func typeOf(res types.Resource) resource.Type {
	switch res.(type) {
	case *listenerv3.Listener:
		return resource.ListenerType
	case *routev3.RouteConfiguration:
		return resource.RouteType
	case *clusterv3.Cluster:
		return resource.ClusterType
	case *endpointv3.ClusterLoadAssignment:
		return resource.EndpointType
	case *routev3.ScopedRouteConfiguration:
		return resource.ScopedRouteType
	case *authv3.Secret:
		return resource.SecretType
	case *runtimev3.Runtime:
		return resource.RuntimeType
	case *corev3.TypedExtensionConfig:
		return resource.ExtensionConfigType
	default:
		return ""
	}
}

func ToMap(resources []types.Resource) map[string][]types.Resource {
	out := map[string][]types.Resource{}
	for _, res := range resources {
		rt := typeOf(res)
		if _, ok := out[rt]; !ok {
			out[rt] = []types.Resource{res}
		} else {
			out[rt] = append(out[rt], res)
		}
	}
	return out
}
