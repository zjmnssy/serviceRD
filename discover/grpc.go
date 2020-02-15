package discover

import (
	"context"
	"encoding/json"
	"fmt"

	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/naming"
	"google.golang.org/grpc/status"
)

type gRPCWatcher struct {
	c      *clientv3.Client
	target string
	ctx    context.Context
	cancel context.CancelFunc
	wch    clientv3.WatchChan
	err    error
}

var errWatcherClosed = fmt.Errorf("naming: watch closed")

// 获取所有目标对象，并开启监测
func (gw *gRPCWatcher) firstNext() ([]*naming.Update, error) {
	// 使用序列化的请求，因此，如果目标etcd服务器从仲裁中分离出来，则解析仍然有效。
	resp, err := gw.c.Get(gw.ctx, gw.target, clientv3.WithPrefix(), clientv3.WithSerializable())
	if gw.err = err; err != nil {
		return nil, err
	}

	updates := make([]*naming.Update, 0, len(resp.Kvs))

	for _, kv := range resp.Kvs {
		var updateTmp naming.Update
		if err := json.Unmarshal(kv.Value, &updateTmp); err != nil {
			continue
		}

		updates = append(updates, &updateTmp)
	}

	opts := []clientv3.OpOption{clientv3.WithRev(resp.Header.Revision + 1), clientv3.WithPrefix(), clientv3.WithPrevKV()}
	gw.wch = gw.c.Watch(gw.ctx, gw.target, opts...)

	return updates, nil
}

// Next 从etcd解析器获取下一组更新.
// 对Next的调用应序列化； 并发调用并不安全，因为没有办法协调更新顺序。
func (gw *gRPCWatcher) Next() ([]*naming.Update, error) {
	if gw.wch == nil {
		return gw.firstNext()
	}

	if gw.err != nil {
		return nil, gw.err
	}

	// 处理监测到的目标事件
	wr, ok := <-gw.wch
	if !ok {
		gw.err = status.Error(codes.Unavailable, errWatcherClosed.Error())
		return nil, gw.err
	}

	gw.err = wr.Err()
	if gw.err != nil {
		return nil, gw.err
	}

	updates := make([]*naming.Update, 0, len(wr.Events))

	for _, e := range wr.Events {
		var updateTmp naming.Update
		var err error

		switch e.Type {
		case clientv3.EventTypePut:
			err = json.Unmarshal(e.Kv.Value, &updateTmp)
			updateTmp.Op = naming.Add
		case clientv3.EventTypeDelete:
			err = json.Unmarshal(e.PrevKv.Value, &updateTmp)
			updateTmp.Op = naming.Delete
		default:
			continue
		}

		if err == nil {
			updates = append(updates, &updateTmp)
		}
	}

	return updates, nil
}

func (gw *gRPCWatcher) Close() {
	gw.cancel()
}

// GRPCResolver 创建一个　grpc.Watcher　用来追踪目标的解析改变.
type GRPCResolver struct {
	Client *clientv3.Client
}

// Update 更新信息
func (gr *GRPCResolver) Update(ctx context.Context, target string, nm naming.Update, opts ...clientv3.OpOption) (err error) {
	switch nm.Op {
	case naming.Add:
		{
			var v []byte

			if v, err = json.Marshal(nm); err != nil {
				return status.Error(codes.InvalidArgument, err.Error())
			}

			_, err = gr.Client.KV.Put(ctx, target+"/"+nm.Addr, string(v), opts...)
		}
	case naming.Delete:
		{
			_, err = gr.Client.Delete(ctx, target+"/"+nm.Addr, opts...)
		}
	default:
		{
			return status.Error(codes.InvalidArgument, "naming: bad naming op")
		}
	}

	return err
}

// Resolve 解析.
func (gr *GRPCResolver) Resolve(target string) (naming.Watcher, error) {
	ctx, cancel := context.WithCancel(context.Background())

	w := &gRPCWatcher{c: gr.Client, target: target + "/", ctx: ctx, cancel: cancel}

	return w, nil
}
