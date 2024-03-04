package k8sreflector

import (
	"context"
	"time"

	"github.com/sifer169966/go-grpc-client-lb/xdsserver/snapshots"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/kubernetes"
)

type ReflectorOptions func(r *Reflector)

func WithResyncPeriod(d time.Duration) ReflectorOptions {
	return func(r *Reflector) {
		r.resyncPeriod = d
	}
}

type Reflector struct {
	k8sClient    kubernetes.Interface
	resyncPeriod time.Duration
	// not thread-safe
	lastSnapshotSum uint64
	snap            snapshots.SnapshotSetter
}

func New(s snapshots.SnapshotSetter, opts ...ReflectorOptions) *Reflector {
	refl := &Reflector{
		snap: s,
	}
	for _, opt := range opts {
		opt(refl)
	}
	if refl.resyncPeriod == 0 {
		refl.resyncPeriod = 5 * time.Minute
	}
	return refl
}

func (r *Reflector) Watch(stopCtx context.Context) error {
	g, ctx := errgroup.WithContext(stopCtx)
	g.Go(func() error {
		return r.watchServices(ctx)
	})
	// g.Go(func() error {
	// 	return s.watchEndpoints(ctx)
	// })
	return g.Wait()
}
