package health

import (
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Manager grpc健康检查管理器
type Manager struct {
	DefaultHealthServer *health.Server
}

func newManager() *Manager {
	// GRPC已实现的健康检查服务，支持并发，同时支持同一个健康检查服务绑定多个GRPC服务
	defaultServer := health.NewServer()

	m := &Manager{
		DefaultHealthServer: defaultServer,
	}

	return m
}

// Register 注册健康检查服务
func (m *Manager) Register(s *grpc.Server, serviceName string) {
	m.DefaultHealthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(s, m.DefaultHealthServer)
}

var instanceManager *Manager
var instanceManagerOnce sync.Once

// GetManager 获取grpc健康检查管理器
func GetManager() *Manager {
	instanceManagerOnce.Do(func() {
		instanceManager = newManager()
	})

	return instanceManager
}
