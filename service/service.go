package service

import (
	"fmt"
	"sync"

	"github.com/zjmnssy/codex"
)

// 服务获取方法
const (
	ServiceSelectMethodRandom = "random"
)

// Service 服务
type Service interface {
	GetServiceRegisterInfo() map[string]string // 用于服务描述自己的信息
}

// Unit 服务
type Unit struct {
	lock sync.Mutex
	info map[string]string
}

// List 服务列表
type List struct {
	lock   sync.Mutex
	idUnit map[string]*Unit // key - service ID
}

// All 所有服务表
type All struct {
	lock     sync.Mutex
	typeList map[string]*List // key - service type
}

// NewAll 创建服务存储实例
func NewAll() *All {
	s := &All{typeList: make(map[string]*List)}

	return s
}

// AddOne 添加服务属性
func (as *All) AddOne(ty string, ID string, key string, value string) {
	as.lock.Lock()
	defer as.lock.Unlock()

	sl, ok := as.typeList[ty]
	if ok {
		sl.lock.Lock()
		defer sl.lock.Unlock()

		su, ok := sl.idUnit[ID]
		if ok {
			su.lock.Lock()
			defer su.lock.Unlock()

			su.info[key] = value
		} else {
			su = &Unit{info: make(map[string]string)}
			su.info[key] = value

			sl.idUnit[ID] = su
		}
	} else {
		sl = &List{idUnit: make(map[string]*Unit)}

		su := &Unit{info: make(map[string]string)}
		su.info[key] = value

		sl.idUnit[ID] = su
		as.typeList[ty] = sl
	}
}

// DeleteOne 删除服务
func (as *All) DeleteOne(ty string, ID string) {
	as.lock.Lock()
	defer as.lock.Unlock()

	sl, ok := as.typeList[ty]
	if !ok {
		return
	}

	sl.lock.Lock()
	defer sl.lock.Unlock()

	delete(sl.idUnit, ID)
}

// Select 选择一个服务
func (as *All) Select(ty string, method string) (map[string]string, error) {

	return as.randomGet(ty)
}

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

// Show 展示全部服务
func (as *All) Show() {
	as.lock.Lock()
	defer as.lock.Unlock()

	for serviceType, list := range as.typeList {
		fmt.Printf("/******************************** %s ********************************/\n", serviceType)

		list.lock.Lock()
		defer list.lock.Unlock()

		for serviceID, ut := range list.idUnit {
			ut.lock.Lock()
			defer ut.lock.Unlock()

			fmt.Printf("---------------- %s ----------------\n", serviceID)

			for k, v := range ut.info {
				fmt.Printf("%s = %s\n", k, v)
			}

			fmt.Printf("----------------------------------------\n\n")
		}
		fmt.Printf("/**********************************************************************/\n\n")
	}
}
