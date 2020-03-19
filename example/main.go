package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/zjmnssy/etcd"
	"github.com/zjmnssy/serviceRD/discover"
	"github.com/zjmnssy/serviceRD/register"
	"github.com/zjmnssy/serviceRD/service"
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

	kvs[fmt.Sprintf("/services/push/%s/%s/version", s.Type, s.ID)] = s.Version
	kvs[fmt.Sprintf("/services/push/%s/%s/address", s.Type, s.ID)] = s.Addr

	return kvs
}

// PauseKey 解析从etcd获取到的key
func PauseKey(key string) (string, string, string, error) {
	sl := strings.Split(key, "/")
	var serviceType string
	var serviceID string
	var serviceAttr string

	if len(sl) == 6 {
		serviceType = sl[3]
		serviceID = sl[4]
		serviceAttr = sl[5]
	} else {
		return "", "", "", fmt.Errorf("key is invalid")
	}

	return serviceType, serviceID, serviceAttr, nil
}

/********************************************　注册示例　************************************************/

func registerImpl() {
	var c etcd.Config
	c.NodeList = append(c.NodeList, "47.56.89.32:2379")
	c.UseTLS = true
	c.CaFile = "/home/nssy/.ssh/ca/etcd1/ca.pem"
	c.CertFile = "/home/nssy/.ssh/ca/etcd1/etcd1.pem"
	c.CertKeyFile = "/home/nssy/.ssh/ca/etcd1/etcd1-key.pem"
	c.ServerName = "www.3344.fun"
	c.DialTimeout = 1500
	c.DialKeepAlivePeriod = 5000
	c.DialKeepAliveTimeout = 2000

	s := HTTPService{ID: "test1", Type: "http", Version: "20190828005", Addr: "192.168.31.218:8080"}

	impl, err := register.New(c, &s, 10)
	if err != nil {
		zlog.Prints(zlog.Warn, "example", "create new register error : %s", err)
		return
	}

	impl.Start()

	go func(i *register.Register) {
		for {
			zlog.Prints(zlog.Info, "example", "is health = %v", i.IsHealth())

			time.Sleep(time.Duration(1) * time.Second)
		}
	}(impl)

	time.Sleep(time.Duration(10) * time.Second)
	zlog.Prints(zlog.Notice, "example", "register end")

	impl.Stop()
	zlog.Prints(zlog.Notice, "example", "stop")
	time.Sleep(time.Duration(10) * time.Second)

	err = impl.Register()
	zlog.Prints(zlog.Notice, "example", "register again")
	time.Sleep(time.Duration(10) * time.Second)

	di := discoverImpl()
	di.All.Show()
	di.All.Select("http", service.ServiceSelectMethodRandom)
	time.Sleep(time.Duration(10) * time.Second)

	impl.Stop()
	zlog.Prints(zlog.Notice, "example", "stop again")
	time.Sleep(time.Duration(10) * time.Second)
	di.All.Show()
	di.All.Select("http", service.ServiceSelectMethodRandom)

	err = impl.Register()
	zlog.Prints(zlog.Notice, "example", "register third")
	time.Sleep(time.Duration(10) * time.Second)
	di.All.Show()
	di.All.Select("http", service.ServiceSelectMethodRandom)

}

/******************************************** 服务发现 ******************************************************/

func discoverImpl() *discover.Discover {
	var c etcd.Config
	c.NodeList = append(c.NodeList, "47.56.89.32:2379")
	c.UseTLS = true
	c.CaFile = "/home/nssy/.ssh/ca/etcd1/ca.pem"
	c.CertFile = "/home/nssy/.ssh/ca/etcd1/etcd1.pem"
	c.CertKeyFile = "/home/nssy/.ssh/ca/etcd1/etcd1-key.pem"
	c.ServerName = "www.3344.fun"
	c.DialTimeout = 1500
	c.DialKeepAlivePeriod = 5000
	c.DialKeepAliveTimeout = 2000

	list := make([]string, 0)
	list = append(list, "/services")

	impl, err := discover.New(c, list, PauseKey)
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
