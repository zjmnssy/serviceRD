package service

import (
	"fmt"

	"github.com/zjmnssy/codex"
)

// randomGet 随机获取一个服务实例
func (as *All) randomGet(ty string) (map[string]string, error) {
	as.lock.Lock()
	defer as.lock.Unlock()

	sl, ok := as.typeList[ty]
	if !ok {
		return nil, fmt.Errorf(fmt.Sprintf("%s type service is none", ty))
	}

	sl.lock.Lock()
	defer sl.lock.Unlock()

	total := len(sl.idUnit)
	if total == 0 {
		return nil, fmt.Errorf("the %s service list is empty", ty)
	}

	index, err := codex.Number(uint64(total))
	if err != nil {
		return nil, err
	}

	var i uint64
	var result map[string]string
	for _, v := range sl.idUnit {
		if i == index || i == uint64(len(sl.idUnit)) {
			result = v.info
			break
		}

		i++
	}

	return result, nil
}
