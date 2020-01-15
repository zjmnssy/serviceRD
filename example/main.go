package main

import (
	"fmt"
	"time"

	"github.com/zjmnssy/etcd"
	"github.com/zjmnssy/serviceRD/discover"
	"github.com/zjmnssy/serviceRD/register"
	"github.com/zjmnssy/system"
	"github.com/zjmnssy/zlog"
)

/******************************************** 定义服务 ************************************************/

// HTTPService http service示例
type HTTPService struct {
	ID      string `json:"ID"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Addr    string `json:"addr"`
}

// GetServiceRegisterInfo 服务描述
func (s *HTTPService) GetServiceRegisterInfo() map[string]string {
	var kvs = make(map[string]string)

	kvs[fmt.Sprintf("/services/%s/%s/put/ID", s.Type, s.ID)] = s.ID
	kvs[fmt.Sprintf("/services/%s/%s/put/version", s.Type, s.ID)] = s.Version
	kvs[fmt.Sprintf("/services/%s/%s/put/addr", s.Type, s.ID)] = s.Addr

	return kvs
}

/********************************************　注册示例　************************************************/

func registerImpl() {
	var c etcd.Config
	c.NodeList = append(c.NodeList, "192.168.31.218:2379")
	c.UseTLS = true
	c.CaFile = "/home/nssy/Work/0-Software/cfssl/root.crt"
	c.CertFile = "/home/nssy/Work/0-Software/cfssl/test3-config-server/etcd.pem"
	c.CertKeyFile = "/home/nssy/Work/0-Software/cfssl/test3-config-server/etcd-key.pem"
	c.ServerName = "config-server"
	c.DialTimeout = 1500
	c.DialKeepAlivePeriod = 5000
	c.DialKeepAliveTimeout = 2000

	s := HTTPService{ID: "test1", Type: "http", Version: "10.000.000.001", Addr: "192.168.31.218:8080"}

	impl, err := register.New(c, &s, 10)
	if err != nil {
		zlog.Prints(zlog.Warn, "example", "create new register error : %s", err)
		return
	}

	impl.Start()

	go func(i *register.Register) {
		for {
			zlog.Prints(zlog.Info, "example", "is health = %v", i.IsHealth())

			time.Sleep(time.Duration(3) * time.Second)
		}
	}(impl)

	time.Sleep(time.Duration(10) * time.Second)

	impl.Stop()

	time.Sleep(time.Duration(10) * time.Second)

	err = impl.Register()

	time.Sleep(time.Duration(10) * time.Second)

	zlog.Prints(zlog.Notice, "example", "register end")

	di := discoverImpl()
	di.All.GetOneRandom("http")

	time.Sleep(time.Duration(10) * time.Second)
	di.All.GetOneRandom("http")

	impl.Stop()

	time.Sleep(time.Duration(10) * time.Second)
	di.All.GetOneRandom("http")

	err = impl.Register()

	di.All.GetOneRandom("http")
}

/******************************************** 服务发现 ******************************************************/

func discoverImpl() *discover.Discover {
	var c etcd.Config
	c.NodeList = append(c.NodeList, "192.168.31.218:2379")
	c.UseTLS = true
	c.CaFile = "/home/nssy/Work/0-Software/cfssl/root.crt"
	c.CertFile = "/home/nssy/Work/0-Software/cfssl/test3-config-server/etcd.pem"
	c.CertKeyFile = "/home/nssy/Work/0-Software/cfssl/test3-config-server/etcd-key.pem"
	c.ServerName = "config-server"
	c.DialTimeout = 1500
	c.DialKeepAlivePeriod = 5000
	c.DialKeepAliveTimeout = 2000

	list := make([]string, 0)
	list = append(list, "/services")

	impl, err := discover.New(c, list)
	if err != nil {
		zlog.Prints(zlog.Warn, "example", "create new discover error : %s", err)
		return nil
	}

	impl.Run()

	return impl
}

/******************************************** main ******************************************************/

func quit() {

}

func main() {
	registerImpl()

	system.SecurityExitProcess(quit)
}
