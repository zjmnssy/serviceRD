package balancer

import (
	"context"
	"math/rand"
	"strconv"
	"sync"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

// Random 随机负载均衡方式名称
const Random = "random-my"

func newRandomBuilder() balancer.Builder {
	return base.NewBalancerBuilderWithConfig(Random, &randomPickerBuilder{}, base.Config{HealthCheck: true})
}

// 注意：需要在包初始化的时候注册到grpc中
func init() {
	balancer.Register(newRandomBuilder())
}

type randomPickerBuilder struct {
}

// Build 支持权重，权重的实现为：按权重值，向队列中增加权重数量的实例
func (*randomPickerBuilder) Build(readySCs map[resolver.Address]balancer.SubConn) balancer.Picker {
	if len(readySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	var scs []balancer.SubConn

	for addr, sc := range readySCs {
		weight := 1

		if addr.Metadata != nil {
			m, ok := addr.Metadata.(*map[string]string)
			if ok {
				w, ok := (*m)["weight"]
				if ok {
					n, err := strconv.Atoi(w)
					if err == nil && n > 0 {
						weight = n
					}
				}
			}
		}

		for i := 0; i < weight; i++ {
			scs = append(scs, sc)
		}
	}

	return &randomPicker{subConns: scs}
}

type randomPicker struct {
	subConns []balancer.SubConn
	mu       sync.Mutex
}

func (p *randomPicker) Pick(ctx context.Context, opt balancer.PickInfo) (balancer.SubConn, func(balancer.DoneInfo), error) {
	p.mu.Lock()
	sc := p.subConns[rand.Intn(len(p.subConns))]
	p.mu.Unlock()

	return sc, nil, nil
}
