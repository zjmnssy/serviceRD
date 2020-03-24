package detector

import (
	"sync"
	"time"

	"github.com/zjmnssy/etcd"
	"github.com/zjmnssy/zlog"
	"go.etcd.io/etcd/clientv3"
	"golang.org/x/net/context"
	"google.golang.org/grpc/resolver"
)

type extractAddr func(key string, value string) (resolver.Address, string, error)

// Watcher 服务监控器
type Watcher struct {
	client      *clientv3.Client
	updateCh    chan []resolver.Address
	extract     extractAddr
	watchPrefix string

	ctx        context.Context
	cancel     context.CancelFunc
	alladdrs   []resolver.Address
	initFinish chan struct{}
	lock       sync.Mutex
}

// NewWatcher 创建服务监控器实例
func NewWatcher(client *clientv3.Client,
	update chan []resolver.Address,
	extract extractAddr, prefix string) *Watcher {

	ctx, cancel := context.WithCancel(context.Background())

	w := &Watcher{
		client:      client,
		updateCh:    update,
		extract:     extract,
		watchPrefix: prefix,
		ctx:         ctx,
		cancel:      cancel,
		alladdrs:    make([]resolver.Address, 0, 0),
		initFinish:  make(chan struct{}),
	}

	return w
}

func (w *Watcher) initialize() []resolver.Address {
	retAddrs := make([]resolver.Address, 0, 0)

	ctxNow, cancel := context.WithTimeout(w.ctx, time.Duration(2)*time.Second)
	defer cancel()

	_, dataMap, err := etcd.GetPrefix(ctxNow, w.client, w.watchPrefix)
	if err != nil {
		zlog.Prints(zlog.Warn, "watcher", "etcd get error = %s", err)
		w.reset(retAddrs)
		w.initFinish <- struct{}{}
		return retAddrs
	}

	for k, v := range dataMap {
		addr, _, err := w.extract(k, v)
		if err != nil {
			zlog.Prints(zlog.Warn, "watcher", "extract addr error = %s", err)
			continue
		}

		retAddrs = append(retAddrs, addr)
	}

	w.reset(retAddrs)
	w.initFinish <- struct{}{}

	return retAddrs
}

func (w *Watcher) watch() {
	etcdData := make(chan etcd.WatchData, 10000)

	go etcd.WatchPrefix(w.ctx, w.client, w.watchPrefix, etcdData)

	<-w.initFinish

	for data := range etcdData {
		switch data.Operate {
		case etcd.MethodCreate:
			{
				addr, _, err := w.extract(data.Key, data.Value)
				if err == nil {
					w.add(addr)
				} else {
					zlog.Prints(zlog.Warn, "watcher", "extract addr error = %s", err)
				}
			}
		case etcd.MethodPut:
			{
				addr, _, err := w.extract(data.Key, data.Value)
				if err == nil {
					w.add(addr)
				} else {
					zlog.Prints(zlog.Warn, "watcher", "extract addr error = %s", err)
				}
			}
		case etcd.MethodDelete:
			{
				_, serverID, err := w.extract(data.Key, data.Value)
				if err == nil {
					w.delete(serverID)
				} else {
					zlog.Prints(zlog.Warn, "watcher", "extract data.Key = %s , data.Value = %s, error = %s", data.Key, data.Value, err)
				}
			}
		case etcd.MethodModify:
			{
				addr, serverID, err := w.extract(data.Key, data.Value)
				if err == nil {
					w.modify(serverID, addr)
				} else {
					zlog.Prints(zlog.Warn, "watcher", "extract addr error = %s", err)
				}
			}
		default:
			{
				zlog.Prints(zlog.Warn, "watcher", "not suport method = %s", data.Operate)
			}
		}
	}
}

func (w *Watcher) add(addr resolver.Address) {
	w.lock.Lock()
	defer w.lock.Unlock()

	newServerID, ok := getDataFromMeta(addr, "serverID")
	if !ok {
		return
	}

	for _, v := range w.alladdrs {
		oldServerID, ok := getDataFromMeta(v, "serverID")
		if !ok {
			return
		}

		if newServerID == oldServerID {
			return
		}
	}

	w.alladdrs = append(w.alladdrs, addr)
	zlog.Prints(zlog.Debug, "watcher", "add w.alladdrs : %v", w.alladdrs)
	w.updateCh <- w.alladdrs
	zlog.Prints(zlog.Debug, "watcher", "add ok ")
}

func (w *Watcher) delete(serverID string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	for i, v := range w.alladdrs {
		oldServerID, ok := getDataFromMeta(v, "serverID")
		if !ok {
			return
		}

		if serverID == oldServerID {
			w.alladdrs = append(w.alladdrs[:i], w.alladdrs[i+1:]...)
			w.updateCh <- w.alladdrs
			return
		}
	}
}

func (w *Watcher) modify(serverID string, addr resolver.Address) {
	w.lock.Lock()
	defer w.lock.Unlock()

	for i, v := range w.alladdrs {
		oldServerID, ok := getDataFromMeta(v, "serverID")
		if !ok {
			return
		}

		if serverID == oldServerID {
			w.alladdrs = append(w.alladdrs[:i], w.alladdrs[i+1:]...)
			w.alladdrs = append(w.alladdrs, addr)
			w.updateCh <- w.alladdrs
			return
		}
	}
}

func (w *Watcher) reset(list []resolver.Address) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.alladdrs = list
	w.updateCh <- w.alladdrs
}

// Run 启动监控器
func (w *Watcher) Run() {
	go w.watch()
	w.initialize()
}

// Close 关闭监控器
func (w *Watcher) Close() {
	w.cancel()
	w.client.Close()
}
