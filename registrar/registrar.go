package registrar

import (
	"context"
	"time"

	"github.com/zjmnssy/etcd"
	"github.com/zjmnssy/serviceRD/service"
	"go.etcd.io/etcd/clientv3"
)

const (
	defaultDialTimeout     = 1500 // per - Millisecond
	defaultLeaseTTL        = 5    // per - Second
	defaultSelfCheckPeriod = 3    // per - Second
)

// Registrar 注册器
type Registrar struct {
	client      *clientv3.Client
	serviceDesc service.Desc
	ttl         int64
	leaseID     clientv3.LeaseID
	cancel      context.CancelFunc
	stopCheck   bool
}

// NewRegistrar 创建注册实例
func NewRegistrar(c etcd.Config, desc service.Desc, ttl int64) (*Registrar, error) {
	client, err := etcd.Client(c)
	if err != nil {
		return nil, err
	}

	r := Registrar{client: client, serviceDesc: desc, ttl: ttl, stopCheck: false}

	return &r, nil
}

// Start 启动自检协程，开始注册和保活
func (r *Registrar) Start() {
	go r.selfCheck()
}

// Stop 停止服务注册和保活以及自检
func (r *Registrar) Stop() {
	r.stopCheck = true

	if r.cancel != nil {
		r.cancel()
	}

	r.leaseID = clientv3.NoLease
	r.cancel = nil
}

// Register 注册服务（非阻塞保活，异步）
func (r *Registrar) Register() error {
	var err error

	if r.leaseID != clientv3.NoLease {
		r.Stop()
	}

	r.stopCheck = false

	ctxTemp, cancel := context.WithTimeout(context.Background(), time.Duration(defaultDialTimeout)*time.Millisecond)
	defer cancel()

	if r.ttl <= 0 {
		r.ttl = defaultLeaseTTL
	}

	_, r.leaseID, err = etcd.CreateLease(ctxTemp, r.client, r.ttl)
	if err != nil {
		return err
	}

	_, err = etcd.TxnPutWithLease(ctxTemp, r.client, r.serviceDesc.GetServiceRegisterInfo(), r.leaseID)
	if err != nil {
		return err
	}

	ctxKeep, cancel := context.WithCancel(context.Background())
	if r.cancel != nil {
		r.cancel()
	}
	r.cancel = cancel

	_, err = etcd.KeepAliveAways(ctxKeep, r.client, r.leaseID)
	if err != nil {
		return err
	}

	return nil
}

// IsHealth 检查注册是否健康
func (r *Registrar) IsHealth() bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(defaultDialTimeout)*time.Millisecond)
	defer cancel()

	_, keys, err := etcd.LeaseTimeToLive(ctx, r.client, r.leaseID)
	if err != nil {
		return false
	}

	if len(keys) == 0 {
		return false
	}

	return true
}

func (r *Registrar) selfCheck() {
	for {
		if !r.IsHealth() && !r.stopCheck {
			r.Register()
		}

		time.Sleep(time.Duration(defaultSelfCheckPeriod) * time.Second)
	}
}
