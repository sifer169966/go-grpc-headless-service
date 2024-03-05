package k8sreflector

import (
	"context"
	"time"

	"github.com/sifer169966/go-grpc-client-lb/xdsserver/snapshots"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/kubernetes"
	k8scache "k8s.io/client-go/tools/cache"
)

type ReflectorOptions func(r *Reflector)

func WithServiceResyncPeriod(d time.Duration) ReflectorOptions {
	return func(r *Reflector) {
		r.svc.resyncPeriod = d
	}
}

func WithEndpointResyncPeriod(d time.Duration) ReflectorOptions {
	return func(r *Reflector) {
		r.ep.resyncPeriod = d
	}
}

type Reflector struct {
	k8sClient kubernetes.Interface
	snap      snapshots.SnapshotSetter
	// services reflector
	svc watcher
	// endpoints reflector
	ep watcher
}

type watcher struct {
	k8sRefl *k8scache.Reflector
	// not thread-safe
	lastSnapshotSum uint64
	watcherConfig
}

type watcherConfig struct {
	resyncPeriod time.Duration
}

func New(s snapshots.SnapshotSetter, c kubernetes.Interface, opts ...ReflectorOptions) *Reflector {
	refl := &Reflector{
		snap:      s,
		k8sClient: c,
	}
	for _, opt := range opts {
		opt(refl)
	}
	setDefaultWatcherConfig(&refl.svc, &refl.ep)
	return refl
}

func setDefaultWatcherConfig(w ...*watcher) {
	for i := range w {
		if w[i].resyncPeriod == 0 {
			w[i].resyncPeriod = 5 * time.Minute
		}
	}
}

func (r *Reflector) Watch(stopCtx context.Context) error {
	g, ctx := errgroup.WithContext(stopCtx)
	g.Go(func() error {
		return r.watchServices(ctx)
	})
	g.Go(func() error {
		return r.watchEndpoints(ctx)
	})
	return g.Wait()
}
