module github.com/zjmnssy/serviceRD

go 1.13

replace go.etcd.io/etcd => go.etcd.io/etcd v0.0.0-20200324205056-bbb0fcfae986

require (
	github.com/golang/protobuf v1.3.5
	github.com/zjmnssy/codex v1.0.1 // indirect
	github.com/zjmnssy/etcd v1.0.2
	github.com/zjmnssy/system v1.0.2
	github.com/zjmnssy/zlog v1.0.3
	go.etcd.io/etcd v0.0.0-20200324205056-bbb0fcfae986
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e
	google.golang.org/grpc v1.28.0
)
