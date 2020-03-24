package detector

import (
	"github.com/zjmnssy/etcd"
	"google.golang.org/grpc/resolver"
)

type etcdResolver struct {
	scheme    string
	conf      etcd.Config
	watchPath string
	extract   extractAddr

	watcher  *Watcher
	updateCh chan []resolver.Address
	stopCh   chan struct{}
	cc       resolver.ClientConn
}

func (r *etcdResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	client, err := etcd.Client(r.conf)
	if err != nil {
		return nil, err
	}

	r.cc = cc
	r.watcher = NewWatcher(client, r.updateCh, r.extract, r.watchPath)
	r.start()

	return r, nil
}

func (r *etcdResolver) Scheme() string {
	return r.scheme
}

func (r *etcdResolver) start() {
	r.watcher.Run()

	go func() {
		for {
			select {
			case <-r.stopCh:
				{
					r.Close()
				}
			case addrs := <-r.updateCh:
				{
					r.cc.UpdateState(resolver.State{Addresses: addrs})
				}
			}
		}
	}()
}

func (r *etcdResolver) ResolveNow(o resolver.ResolveNowOption) {
}

func (r *etcdResolver) Close() {
	r.watcher.Close()
}
