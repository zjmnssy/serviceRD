# serviceRD
a service register and discover, base on etcd

# description
etcd信息存储示例如下:  
```
/services
        /serviceType
                /common(此目录下存储此类服务共用的信息，由各自服务自行拉取使用)
                        /ca                  sssssssssssssssssssssssssssssssssss
                        /cert                sssssssssssssssssssssssssssssssssss
                        /key                 sssssssssssssssssssssssssssssssssss
                /serviceID1
                        /get(此目录下存储此服务要使用的信息，由此服务自行拉取使用)
                                /logPath     /home/test/log1
                                /netEvn      internet
                        /put(此目录下存储此类服务注册的信息，由其他依赖此类服务的服务自行拉取使用)
                                /addr        192.168.1.128:8080
                                /version     10.000.000.001
                /serviceID2
                        /get 
                                /logPath     /home/test/log2
                                /netEvn      all
                        /put 
                                /addr        192.168.1.129:8080
                                /version     10.000.000.002
```
