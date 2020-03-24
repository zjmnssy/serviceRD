package balancer

import (
	"hash/fnv"
	"sort"
	"strconv"
	"sync"
)

type hashFunc func(data []byte) uint32

// 默认参数
const (
	DefaultReplicas = 10
	Salt            = "n*@if09g3n"
)

// DefaultHash 默认hash函数
func DefaultHash(data []byte) uint32 {
	f := fnv.New32()
	f.Write(data)
	return f.Sum32()
}

// Ketama 哈希表存储器
type Ketama struct {
	sync.Mutex
	hash     hashFunc
	replicas int
	keys     []int //  Sorted keys
	hashMap  map[int]string
}

// NewKetama 新建哈希表存储器
func NewKetama(replicas int, fn hashFunc) *Ketama {
	h := &Ketama{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}

	if h.replicas <= 0 {
		h.replicas = DefaultReplicas
	}

	if h.hash == nil {
		h.hash = DefaultHash
	}

	return h
}

// IsEmpty hash map 是否为空
func (h *Ketama) IsEmpty() bool {
	h.Lock()
	defer h.Unlock()

	return len(h.keys) == 0
}

// Add 新增
func (h *Ketama) Add(nodes ...string) {
	h.Lock()
	defer h.Unlock()

	for _, node := range nodes {
		for i := 0; i < h.replicas; i++ {
			key := int(h.hash([]byte(Salt + strconv.Itoa(i) + node)))

			_, ok := h.hashMap[key]
			if !ok {
				h.keys = append(h.keys, key)
			}

			h.hashMap[key] = node
		}
	}

	sort.Ints(h.keys)
}

// Remove 删除
func (h *Ketama) Remove(nodes ...string) {
	h.Lock()
	defer h.Unlock()

	deletedKey := make([]int, 0)
	for _, node := range nodes {
		for i := 0; i < h.replicas; i++ {
			key := int(h.hash([]byte(Salt + strconv.Itoa(i) + node)))

			if _, ok := h.hashMap[key]; ok {
				deletedKey = append(deletedKey, key)
				delete(h.hashMap, key)
			}
		}
	}

	if len(deletedKey) > 0 {
		h.deleteKeys(deletedKey)
	}
}

func (h *Ketama) deleteKeys(deletedKeys []int) {
	sort.Ints(deletedKeys)

	index := 0
	count := 0
	for _, key := range deletedKeys {
		for ; index < len(h.keys); index++ {
			h.keys[index-count] = h.keys[index]

			if key == h.keys[index] {
				count++
				index++
				break
			}
		}
	}

	for ; index < len(h.keys); index++ {
		h.keys[index-count] = h.keys[index]
	}

	h.keys = h.keys[:len(h.keys)-count]
}

// Get 获取
func (h *Ketama) Get(key string) (string, bool) {
	if h.IsEmpty() {
		return "", false
	}

	hash := int(h.hash([]byte(key)))

	h.Lock()
	defer h.Unlock()

	idx := sort.Search(len(h.keys), func(i int) bool {
		return h.keys[i] >= hash
	})

	if idx == len(h.keys) {
		idx = 0
	}

	str, ok := h.hashMap[h.keys[idx]]

	return str, ok
}
