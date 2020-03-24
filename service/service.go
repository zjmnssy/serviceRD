package service

// Desc 服务描述接口
type Desc interface {
	GetServiceRegisterInfo() map[string]string // 获取服务描述自己的信息，用于注册使用
}
