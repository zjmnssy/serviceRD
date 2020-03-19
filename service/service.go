package service

// Service 服务
type Service interface {
	GetServiceRegisterInfo() map[string]string // 用于服务描述自己的信息
}
