package snapshots

import (
	"context"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/sifer169966/go-grpc-client-lb/xdsserver/pkg/resources"
	"k8s.io/client-go/kubernetes"
)

const (
	resourceKindLDS   = "LDS"
	resourceKindRDS   = "RDS"
	resourceKindCDS   = "CDS"
	resourceKindEDS   = "EDS"
	resourceKindMixed = "LDS/RDS/CDS"
)

type SnapshotSetter interface {
	Set(ctx context.Context, version string, res []types.Resource)
}

// Snapshot ...
type Snapshot struct {
	muxCache      cachev3.MuxCache
	mixedSnapshot cachev3.SnapshotCache
	edsSnapshot   cachev3.SnapshotCache
	k8sClient     kubernetes.Interface
}

func getResourceKeyName(typeURL string) string {
	switch typeURL {
	case resource.ListenerType, resource.RouteType, resource.ClusterType:
		return resourceKindMixed
	case resource.EndpointType:
		return resourceKindEDS
	default:
		return ""
	}
}

// New ...
// create a new instance of snapshot to capture and hold the discovery information at a point of time
func New() *Snapshot {
	mixedSnapshot := cachev3.NewSnapshotCache(false, DefaultNodeID{}, nil)
	edsSnapshot := cachev3.NewSnapshotCache(false, DefaultNodeID{}, nil)
	muxCache := cachev3.MuxCache{
		Classify: func(r *cachev3.Request) string {
			return getResourceKeyName(r.TypeUrl)
		},
		ClassifyDelta: func(r *cachev3.DeltaRequest) string {
			return getResourceKeyName(r.TypeUrl)
		},
		Caches: map[string]cachev3.Cache{
			resourceKindMixed: mixedSnapshot,
			resourceKindEDS:   edsSnapshot,
		},
	}
	return &Snapshot{
		muxCache:      muxCache,
		mixedSnapshot: mixedSnapshot,
		edsSnapshot:   edsSnapshot,
	}
}

func (s *Snapshot) MuxCache() *cachev3.MuxCache {
	return &s.muxCache
}

// Set ...
// set the snapshot for lds, rds, cds and the separate snapshot for eds
func (s *Snapshot) Set(ctx context.Context, version string, res []types.Resource) {
	resMap := resources.ToMap(res)
	snapshot, err := cachev3.NewSnapshot(version, resMap)
	if err != nil {
		panic(err)
	}
	// set resource eds snapshot, otherwise, set to the mixedSnapshot instead
	if _, ok := resMap[resource.EndpointType]; ok {
		s.edsSnapshot.SetSnapshot(ctx, "", snapshot)
	} else {
		s.mixedSnapshot.SetSnapshot(ctx, "", snapshot)
	}

}
