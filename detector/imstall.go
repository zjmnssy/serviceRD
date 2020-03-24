package detector

import (
	"github.com/zjmnssy/etcd"
	"google.golang.org/grpc/resolver"
)

// RegisterResolver 注册解析器
func RegisterResolver(scheme string, conf etcd.Config, watchPath string, extract extractAddr) {
	resolver.Register(&etcdResolver{
		scheme:    scheme,
		conf:      conf,
		watchPath: watchPath,
		extract:   extract,
		updateCh:  make(chan []resolver.Address, 1000),
		stopCh:    make(chan struct{}),
	})
}
