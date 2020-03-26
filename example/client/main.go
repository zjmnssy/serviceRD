package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/zjmnssy/etcd"
	"github.com/zjmnssy/serviceRD/balancer"
	"github.com/zjmnssy/serviceRD/detector"
	"github.com/zjmnssy/serviceRD/example/proto"
	"github.com/zjmnssy/system"
	"github.com/zjmnssy/zlog"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/health"
	"google.golang.org/grpc/resolver"
)

/***************************************** 获取grpc连接 **************************************************/

// NameUnit 服务方法
type NameUnit struct {
	Service string `json:"service"` // package.Service
	//Method  string `json:"method"`
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxAttempts          int    `json:"maxAttempts"` // >= 2
	InitialBackoff       string `json:"initialBackoff"`
	MaxBackoff           string `json:"maxBackoff"`
	BackoffMultiplier    int    `json:"backoffMultiplier"`    // > 0
	RetryableStatusCodes []int  `json:"retryableStatusCodes"` // eg. [14]
}

// MethodConfigUnit 方法配置
type MethodConfigUnit struct {
	Name                    []NameUnit  `json:"name"`
	RetryPolicy             RetryPolicy `json:"retryPolicy"`
	WaitForReady            bool        `json:"waitForReady"`
	Timeout                 string      `json:"timeout"`
	MaxRequestMessageBytes  int         `json:"maxRequestMessageBytes"`
	MaxResponseMessageBytes int         `json:"maxResponseMessageBytes"`
}

// RetryThrottling 重试阈值控制
type RetryThrottling struct {
	MaxTokens  uint `json:"maxTokens"`  // (0, 1000]
	TokenRatio uint `json:"tokenRatio"` // (0, 1]
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	ServiceName string `json:"serviceName"` // package.Service
}

// ServerConfig grpc服务配置
type ServerConfig struct {
	LoadBalancingPolicy string             `json:"loadBalancingPolicy"`
	MethodConfig        []MethodConfigUnit `json:"methodConfig"`
	RetryThrottling     RetryThrottling    `json:"retryThrottling"`
	HealthCheckConfig   HealthCheckConfig  `json:"healthCheckConfig"`
}

func getGRPCConn(serviceName string, scheme string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(2)*time.Second)
	defer cancel()

	var serverConfig ServerConfig
	serverConfig.LoadBalancingPolicy = balancer.Random
	var methodConfig = MethodConfigUnit{Name: make([]NameUnit, 0, 0)}
	var nameUnit = NameUnit{Service: serviceName}
	methodConfig.Name = append(methodConfig.Name, nameUnit)
	var retryPolicy = RetryPolicy{RetryableStatusCodes: make([]int, 0, 0)}
	retryPolicy.RetryableStatusCodes = append(retryPolicy.RetryableStatusCodes, 14)
	retryPolicy.MaxAttempts = 2
	retryPolicy.InitialBackoff = "0.1s"
	retryPolicy.MaxBackoff = "1s"
	retryPolicy.BackoffMultiplier = 1
	methodConfig.RetryPolicy = retryPolicy
	methodConfig.WaitForReady = true
	methodConfig.Timeout = "1.5s"
	methodConfig.MaxRequestMessageBytes = 1024 * 1024 * 1024
	methodConfig.MaxResponseMessageBytes = 1024 * 1024 * 1024
	serverConfig.MethodConfig = make([]MethodConfigUnit, 0, 0)
	serverConfig.MethodConfig = append(serverConfig.MethodConfig, methodConfig)
	var retryThrottling = RetryThrottling{MaxTokens: 1000, TokenRatio: 1}
	serverConfig.RetryThrottling = retryThrottling
	var healthCheckConfig = HealthCheckConfig{ServiceName: serviceName}
	serverConfig.HealthCheckConfig = healthCheckConfig

	bytes, err := json.Marshal(serverConfig)
	if err != nil {
		return nil, err
	}

	cc, err := grpc.DialContext(ctx,
		fmt.Sprintf("%s:///", scheme),
		//grpc.WithBlock(), // 如果使用WithBlock()， 此接口返回失败，导致外面调用不好处理， 可能进入不了服务发现和负载均衡
		grpc.WithInsecure(),
		grpc.WithBackoffMaxDelay(time.Second),
		grpc.WithDisableServiceConfig(),
		grpc.WithDefaultServiceConfig(string(bytes)),
	)
	if err != nil {
		zlog.Prints(zlog.Warn, "main", "grpc dial: %s", err)
		return nil, err
	}

	return cc, nil
}

/***************************************** grpc client **************************************************/

// GrpcService http service示例
type GrpcService struct {
	Addr       string `json:"address"`
	Version    string `json:"version"`
	Weight     string `json:"weight"`
	ServerID   string `json:"serverID"`
	ServerType string `json:"serverType"`
}

func extractAddr(key string, value string) (resolver.Address, string, error) {
	var addr resolver.Address
	var serverID string
	var s GrpcService
	metaData := make(map[string]string)

	if value == "" {
		strList := strings.Split(key, "/")
		if len(strList) != 5 {
			return addr, serverID, fmt.Errorf("key = %s error", key)
		}
		serverID = strList[4]
	} else {
		err := json.Unmarshal([]byte(value), &s)
		if err != nil {
			return addr, serverID, err
		}
	}

	addr.Addr = s.Addr

	metaData["version"] = s.Version
	metaData["weight"] = s.Weight
	metaData["serverID"] = s.ServerID
	metaData["serverType"] = s.ServerType
	addr.Metadata = &metaData

	serverID = s.ServerID

	return addr, serverID, nil
}

func exampleGRPC() {
	cc, err := getGRPCConn("proto.Test", "grpc")
	if err != nil {
		zlog.Prints(zlog.Warn, "main", "grpc dial: %s", err)
		return
	}
	defer cc.Close()

	clientGrpc := proto.NewTestClient(cc)

	for i := 0; i < 1000; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(2)*time.Second)
		defer cancel()

		resp, err := clientGrpc.Say(ctx, &proto.SayReq{Content: "round robin"})
		if err != nil {
			zlog.Prints(zlog.Warn, "main", "grpc index = %d error = %s", i, err)
			time.Sleep(time.Second)
			continue
		}
		time.Sleep(time.Second)

		zlog.Prints(zlog.Info, "main", "grpc index = %d response.Content: %s", i, resp.Content)
	}
}

/***************************************** main **************************************************/

func quit() {

}

func main() {
	var c etcd.Config
	c.NodeList = append(c.NodeList, "127.0.0.1:2379")
	c.UseTLS = true
	c.CaFile = "/home/nssy/Work/4-zjmnssy/serviceRD/example/server/ca.pem"
	c.CertFile = "/home/nssy/Work/4-zjmnssy/serviceRD/example/server/etcd.pem"
	c.CertKeyFile = "/home/nssy/Work/4-zjmnssy/serviceRD/example/server/etcd-key.pem"
	c.ServerName = "etcd1"
	c.DialTimeout = 1500
	c.DialKeepAlivePeriod = 5000
	c.DialKeepAliveTimeout = 2000

	// 依赖多个服务的情况下，每种服务集群注册一个对应的解析器，然后在构建客户端请求的时候，调用对应的scheme:///即可
	detector.RegisterResolver("grpc", c, "/services/push/grpcTest", extractAddr)
	//detector.RegisterResolver("http", c, "/services/push/httpTest", extractAddr)

	go exampleGRPC()

	system.SecurityExitProcess(quit)
}
