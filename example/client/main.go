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
	"google.golang.org/grpc/resolver"
)

const (
	serverType = "grpcTest"
)

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
	cc, err := grpc.Dial("grpc:///",
		grpc.WithInsecure(),
		grpc.WithBalancerName(balancer.RoundRobin)) // grpc.WithTimeout(time.Duration(3000)*time.Microsecond)
	if err != nil {
		zlog.Prints(zlog.Warn, "main", "grpc dial: %s", err)
		return
	}
	defer cc.Close()

	clientGrpc := proto.NewTestClient(cc)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(100000)*time.Second)
	defer cancel()

	for i := 0; i < 1000; i++ {
		resp, err := clientGrpc.Say(ctx, &proto.SayReq{Content: "round robin"})
		if err != nil {
			zlog.Prints(zlog.Warn, "main", "index = %d error = %s", i, err)
			time.Sleep(time.Second)
			continue
		}
		time.Sleep(time.Second)

		zlog.Prints(zlog.Info, "main", "index = %d response.Content: %s", i, resp.Content)
	}
}

func example2HTTP() {
	cc, err := grpc.Dial("http:///",
		grpc.WithInsecure(),
		grpc.WithBalancerName(balancer.RoundRobin)) // grpc.WithTimeout(time.Duration(3000)*time.Microsecond)
	if err != nil {
		zlog.Prints(zlog.Warn, "main", "grpc dial: %s", err)
		return
	}
	defer cc.Close()

	clientGrpc := proto.NewTestClient(cc)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(100000)*time.Second)
	defer cancel()

	for i := 0; i < 1000; i++ {
		resp, err := clientGrpc.Say(ctx, &proto.SayReq{Content: "round robin"})
		if err != nil {
			zlog.Prints(zlog.Warn, "main", "index = %d error = %s", i, err)
			time.Sleep(time.Second)
			continue
		}
		time.Sleep(time.Second)

		zlog.Prints(zlog.Info, "main", "index = %d response.Content: %s", i, resp.Content)
	}
}

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
	detector.RegisterResolver("http", c, "/services/push/httpTest", extractAddr)

	go exampleGRPC()
	go example2HTTP()

	system.SecurityExitProcess(quit)
}
