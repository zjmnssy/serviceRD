module github.com/zjmnssy/serviceRD

go 1.13

replace go.etcd.io/etcd => go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738

require (
	github.com/golang/protobuf v1.3.2
	github.com/zjmnssy/codex v1.0.0 // indirect
	github.com/zjmnssy/etcd v1.0.1
	github.com/zjmnssy/system v1.0.2
	github.com/zjmnssy/zlog v1.0.2
	go.etcd.io/etcd v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20200114155413-6afb5195e5aa
	golang.org/x/sys v0.0.0-20200113162924-86b910548bc1 // indirect
	golang.org/x/tools v0.0.0-20200115044656-831fdb1e1868 // indirect
	google.golang.org/genproto v0.0.0-20200113173426-e1de0a7b01eb // indirect
	google.golang.org/grpc v1.26.0
)
