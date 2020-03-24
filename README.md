# serviceRD
a service register and discover, base on etcd

# description
## etcd信息存储示例如下:  
```
/services
        /pull
                /serviceType
                        /common                      
                                /ca                  sssssssssssssssssssssssssssssssssss
                                /cert                sssssssssssssssssssssssssssssssssss
                                /key                 sssssssssssssssssssssssssssssssssss
                        /serviceID1
                                /weight         10
        /push
                /serviceType
                        /serviceID1             {"address":"192.168.1.128:8080", "version":"20190828001", "weight":"10"}
                        /serviceID2             {"address":"192.168.1.128:8080", "version":"20190828001", "weight":"10"}
```
## 说明
- 1.版本组成为：年月日＋三位序号，方便比较计算  
- 2./services/pull此目录下存储此类服务共用的信息，由各自服务监控common和自己的serviceID下的配置更新自己的现有配置  
- 3./services/push为服务注册目录
- 4.服务的粒度目前只定义到提供服务的程序，未精确到单个服务方法
- 5.

# 服务定义



# 服务注册



# 服务发现


