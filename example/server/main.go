package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/zjmnssy/etcd"
	"github.com/zjmnssy/serviceRD/example/proto"
	"github.com/zjmnssy/serviceRD/health"
	"github.com/zjmnssy/serviceRD/registrar"
	"github.com/zjmnssy/serviceRD/service"
	"github.com/zjmnssy/system"
	"github.com/zjmnssy/zlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

/************************************************* service desc ***************************************************/

// serviceDesc 服务描述
type serviceDesc struct {
	Addr       string `json:"address"`
	Version    string `json:"version"`
	Weight     string `json:"weight"`
	ServerID   string `json:"serverID"`
	ServerType string `json:"serverType"`
}

// GetServiceRegisterInfo 服务描述
func (s *serviceDesc) GetServiceRegisterInfo() map[string]string {
	var kvs = make(map[string]string)

	bytes, err := json.Marshal(s)
	if err != nil {
		zlog.Prints(zlog.Warn, "main", "json marshal error = %s", err)
	}

	kvs[fmt.Sprintf("/services/push/%s/%s", s.ServerType, s.ServerID)] = string(bytes)

	return kvs
}

/*************************************************** test server **************************************************/

// RPCServer rpc服务
type RPCServer struct {
	info      serviceDesc
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
	addr, err := GetClientIP(ctx)
	if err != nil {
		zlog.Prints(zlog.Warn, "main", "GetClientIP error = %s", err)
	}

	text := "Hello " + addr + ", " + req.Content + ", I am " + s.info.ServerID

	zlog.Prints(zlog.Info, "main", "response : %s", text)

	return &proto.SayResp{Content: text}, nil
}

/******************************************************** start ***************************************************/

func getGrpcServer(c etcd.Config, desc service.Desc, serviceName string, ttl int64) (*grpc.Server, *registrar.Registrar, error) {
	impl, err := registrar.NewRegistrar(c, desc, ttl)
	if err != nil {
		zlog.Prints(zlog.Warn, "example", "create new register error = %s", err)
		return nil, nil, err
	}

	s := grpc.NewServer()

	manager := health.GetManager()
	manager.Register(s, serviceName)

	return s, impl, nil
}

// GetClientIP 获取请求客户端的远程地址, 通过从metadata中获取远程地址信息
func GetClientIP(ctx context.Context) (string, error) {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("[getClinetIP] invoke FromContext() failed")
	}

	if pr.Addr == net.Addr(nil) {
		return "", fmt.Errorf("[getClientIP] peer.Addr is nil")
	}

	return pr.Addr.String(), nil
}

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

	serviceDesc := serviceDesc{
		Addr:       "0.0.0.0:10001",
		Version:    "20190828001",
		Weight:     "1",
		ServerID:   "node1",
		ServerType: "grpcTest",
	}

	s, impl, err := getGrpcServer(c, &serviceDesc, "proto.Test", 5)
	if err != nil {
		zlog.Prints(zlog.Warn, "example", "getGRPCServer error = %s", err)
		return
	}

	server := &RPCServer{info: serviceDesc, registrar: impl, s: s}
	go server.Run()
}

/**************************************************** main *************************************************/

func quit() {

}

func main() {
	startGRPC()

	system.SecurityExitProcess(quit)
}
