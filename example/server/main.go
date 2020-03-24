package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/zjmnssy/etcd"
	"github.com/zjmnssy/serviceRD/example/proto"
	"github.com/zjmnssy/serviceRD/registrar"
	"github.com/zjmnssy/system"
	"github.com/zjmnssy/zlog"
	"google.golang.org/grpc"
)

// serviceInfo 服务描述
type serviceInfo struct {
	Addr       string `json:"address"`
	Version    string `json:"version"`
	Weight     string `json:"weight"`
	ServerID   string `json:"serverID"`
	ServerType string `json:"serverType"`
}

// GetServiceRegisterInfo 服务描述
func (s *serviceInfo) GetServiceRegisterInfo() map[string]string {
	var kvs = make(map[string]string)

	bytes, err := json.Marshal(s)
	if err != nil {
		zlog.Prints(zlog.Warn, "main", "json marshal error = %s", err)
	}

	kvs[fmt.Sprintf("/services/push/%s/%s", s.ServerType, s.ServerID)] = string(bytes)

	return kvs
}

/***********************************************************************************************************/

// RPCServer rpc服务
type RPCServer struct {
	info      serviceInfo
	registrar *registrar.Registrar
	s         *grpc.Server
}

// Run 启动
func (s *RPCServer) Run() {
	listener, err := net.Listen("tcp", s.info.Addr)
	if err != nil {
		log.Printf("failed to listen: %v", err)
		return
	}

	s.registrar.Start()

	zlog.Prints(zlog.Info, "main", "rpc listening on:%s", s.info.Addr)

	proto.RegisterTestServer(s.s, s)
	s.s.Serve(listener)
}

// Stop 停止
func (s *RPCServer) Stop() {
	s.s.GracefulStop()
}

// Say 远程调用方法
func (s *RPCServer) Say(ctx context.Context, req *proto.SayReq) (*proto.SayResp, error) {
	text := "Hello " + req.Content + ", I am " + s.info.ServerID

	zlog.Prints(zlog.Info, "main", "response : %s", text)

	return &proto.SayResp{Content: text}, nil
}

/***********************************************************************************************************/
func startGRPC() {
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

	serviceDesc := serviceInfo{
		Addr:       "0.0.0.0:10001",
		Version:    "20190828001",
		Weight:     "1",
		ServerID:   "node1",
		ServerType: "grpcTest",
	}

	impl, err := registrar.NewRegistrar(c, &serviceDesc, 5)
	if err != nil {
		zlog.Prints(zlog.Warn, "example", "create new register error : %s", err)
		return
	}

	s := grpc.NewServer()
	server := &RPCServer{info: serviceDesc, registrar: impl, s: s}
	go server.Run()
}

func start2HTTP() {
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

	serviceDesc := serviceInfo{
		Addr:       "0.0.0.0:10002",
		Version:    "20190828001",
		Weight:     "1",
		ServerID:   "node2",
		ServerType: "httpTest",
	}

	impl, err := registrar.NewRegistrar(c, &serviceDesc, 5)
	if err != nil {
		zlog.Prints(zlog.Warn, "example", "create new register error : %s", err)
		return
	}

	s := grpc.NewServer()
	server := &RPCServer{info: serviceDesc, registrar: impl, s: s}
	go server.Run()
}

func quit() {

}

func main() {
	startGRPC()
	start2HTTP()

	system.SecurityExitProcess(quit)
}
