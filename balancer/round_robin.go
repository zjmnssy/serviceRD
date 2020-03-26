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

// RoundRobin 轮询负载均衡方式名称
const RoundRobin = "round_robin"

func newRoundRobinBuilder() balancer.Builder {
	return base.NewBalancerBuilderWithConfig(RoundRobin, &roundRobinPickerBuilder{}, base.Config{HealthCheck: true})
}

// 注意：需要在包初始化的时候注册到grpc中
func init() {
	balancer.Register(newRoundRobinBuilder())
}

type roundRobinPickerBuilder struct{}

func (*roundRobinPickerBuilder) Build(readySCs map[resolver.Address]balancer.SubConn) balancer.Picker {
	if len(readySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	var scs []balancer.SubConn
	for addr, sc := range readySCs {
		// version filter

		// weight
		weight := 1
		if addr.Metadata != nil {
			if m, ok := addr.Metadata.(*map[string]string); ok {
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

	return &roundRobinPicker{
		subConns: scs,
		next:     rand.Intn(len(scs)),
	}
}

type roundRobinPicker struct {
	subConns []balancer.SubConn
	mu       sync.Mutex
	next     int
}

func (p *roundRobinPicker) Pick(ctx context.Context, opt balancer.PickInfo) (balancer.SubConn, func(balancer.DoneInfo), error) {
	p.mu.Lock()
	sc := p.subConns[p.next]
	p.next = (p.next + 1) % len(p.subConns)
	p.mu.Unlock()

	return sc, nil, nil
}
