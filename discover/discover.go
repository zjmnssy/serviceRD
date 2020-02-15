package discover

import (
	"context"

	"github.com/zjmnssy/etcd"
	"github.com/zjmnssy/serviceRD/service"
	"github.com/zjmnssy/zlog"
	"go.etcd.io/etcd/clientv3"
)

type pauseEtcdKey func(key string) (string, string, string, error)

// Discover 服务发现
type Discover struct {
	All *service.All

	pauseEtcdKey pauseEtcdKey
	client       *clientv3.Client
	list         []string
	isGetFinish  chan struct{}
}

// New 创建发现实例
func New(c etcd.Config, l []string, pause pauseEtcdKey) (*Discover, error) {
	client, err := etcd.Client(c)
	if err != nil {
		return nil, err
	}

	d := Discover{client: client,
		All:          service.NewAll(),
		pauseEtcdKey: pause,
		list:         l,
		isGetFinish:  make(chan struct{})}

	return &d, nil
}

// Run 服务发现主流程
func (d *Discover) Run() {
	go d.listen()

	d.initialize()
}

func (d *Discover) initialize() {
	for _, v := range d.list {
		_, data, err := etcd.GetPrefix(context.Background(), d.client, v)
		if err != nil {
			zlog.Prints(zlog.Critical, "discover", "get %s error : %s", v, err)
			continue
		}

		for k, v := range data {
			serviceType, serviceID, serviceAttr, err := d.pauseEtcdKey(k)
			if err == nil {
				d.All.AddOne(serviceType, serviceID, serviceAttr, v)
			}
		}
	}

	d.isGetFinish <- struct{}{}
}

func (d *Discover) listen() {
	var data = make(chan etcd.WatchData, 100000)

	for _, v := range d.list {
		go etcd.WatchPrefix(context.Background(), d.client, v, data)
	}

	<-d.isGetFinish

	for {
		dc := <-data

		if dc.Operate == etcd.MethodPut {
			serviceType, serviceID, serviceAttr, err := d.pauseEtcdKey(dc.Key)
			if err == nil {
				d.All.AddOne(serviceType, serviceID, serviceAttr, dc.Value)
			}
		} else if dc.Operate == etcd.MethodCreate {
			serviceType, serviceID, serviceAttr, err := d.pauseEtcdKey(dc.Key)
			if err == nil {
				d.All.AddOne(serviceType, serviceID, serviceAttr, dc.Value)
			}
		} else if dc.Operate == etcd.MethodModify {
			serviceType, serviceID, serviceAttr, err := d.pauseEtcdKey(dc.Key)
			if err == nil {
				d.All.AddOne(serviceType, serviceID, serviceAttr, dc.Value)
			}
		} else if dc.Operate == etcd.MethodDelete {
			serviceType, serviceID, _, err := d.pauseEtcdKey(dc.Key)
			if err == nil {
				d.All.DeleteOne(serviceType, serviceID)
			}
		} else {
			zlog.Prints(zlog.Debug, "discover", "unknown op=%s, k=%s, v=%s", dc.Operate, dc.Key, dc.Value)
		}
	}
}
