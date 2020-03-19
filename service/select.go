package service

// Select 选择一个服务
func (as *All) Select(ty string, method string) (map[string]string, error) {
	switch method {
	case ServiceSelectMethodRandom:
		{
			return as.randomGet(ty)
		}
	case ServiceSelectMethodPoll:
		{
			return as.randomGet(ty)
		}
	case ServiceSelectMethodWeight:
		{
			return as.randomGet(ty)
		}
	default:
		{
			return as.randomGet(ty)
		}
	}
}
