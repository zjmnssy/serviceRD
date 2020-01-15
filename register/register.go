package register

import (
	"context"
	"time"

	"github.com/zjmnssy/etcd"
	"github.com/zjmnssy/serviceRD/service"
	"go.etcd.io/etcd/clientv3"
)

const (
	defaultDailTimeout     = 1500 // per - Millisecond
	defaultLeaseTTL        = 3
	defaultSelfCheckPeriod = 5 // per - Second
)

// Register 注册
type Register struct {
	client    *clientv3.Client
	service   service.Service
	ttl       int64
	leaseID   clientv3.LeaseID
	cancel    context.CancelFunc
	stopCheck bool
}

// New 创建注册实例
func New(c etcd.Config, s service.Service, t int64) (*Register, error) {
	client, err := etcd.Client(c)
	if err != nil {
		return nil, err
	}

	i := Register{client: client, service: s, ttl: t, stopCheck: false}

	return &i, nil
}

// Start 启动自检协程，开始注册和保活
func (i *Register) Start() {
	go i.selfCheck()
}

// Stop 停止服务注册和保活以及自检
func (i *Register) Stop() {
	i.stopCheck = true

	if i.cancel != nil {
		i.cancel()
	}

	i.leaseID = clientv3.NoLease
	i.cancel = nil
}

// Register 注册服务（非阻塞保活，异步）
func (i *Register) Register() error {
	var err error

	if i.leaseID != clientv3.NoLease {
		i.Stop()
	}

	i.stopCheck = false

	ctxTemp, cancel := context.WithTimeout(context.Background(), time.Duration(defaultDailTimeout)*time.Millisecond)
	defer cancel()

	if i.ttl <= 0 {
		i.ttl = defaultLeaseTTL
	}

	_, i.leaseID, err = etcd.CreateLease(ctxTemp, i.client, i.ttl)
	if err != nil {
		return err
	}

	_, err = etcd.TxnPutWithLease(ctxTemp, i.client, i.service.GetServiceRegisterInfo(), i.leaseID)
	if err != nil {
		return err
	}

	ctxKeep, cancel := context.WithCancel(context.Background())
	if i.cancel != nil {
		i.cancel()
	}
	i.cancel = cancel

	_, err = etcd.KeepAliveAways(ctxKeep, i.client, i.leaseID)
	if err != nil {
		return err
	}

	return nil
}

// IsHealth 检查注册是否健康
func (i *Register) IsHealth() bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(defaultDailTimeout)*time.Millisecond)
	defer cancel()

	_, keys, err := etcd.LeaseTimeToLive(ctx, i.client, i.leaseID)
	if err != nil {
		return false
	}

	if len(keys) == 0 {
		return false
	}

	return true
}

func (i *Register) selfCheck() {
	for {
		if !i.IsHealth() && !i.stopCheck {
			i.Register()
		}

		time.Sleep(time.Duration(defaultSelfCheckPeriod) * time.Second)
	}
}
