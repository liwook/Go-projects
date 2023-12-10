package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type HashFunc func(data []byte) uint32

// 定义哈希环
type HashRing struct {
	hashFunc HashFunc       //定义的哈希算法
	replicas int            //虚拟节点的数量
	keys     []int          //排序好的哈希环
	hashMap  map[int]string //虚拟节点与真实节点的映射关系，key是虚拟节点的哈希值，value是真实节点名称
}

func NewHash(replicas int, fn HashFunc) *HashRing {
	h := &HashRing{
		replicas: replicas,
		hashFunc: fn,
		hashMap:  make(map[int]string),
	}
	if h.hashFunc == nil {
		h.hashFunc = crc32.ChecksumIEEE
	}
	return h
}

func (h *HashRing) Add(realNodeName ...string) {
	for _, name := range realNodeName {
		for i := 0; i < h.replicas; i++ {
			hash := int(h.hashFunc([]byte(strconv.Itoa(i) + name)))
			h.keys = append(h.keys, hash)
			h.hashMap[hash] = name
		}
	}
	sort.Ints(h.keys)
}

// 选择节点
func (h *HashRing) Get(key string) string {
	if len(h.keys) == 0 {
		return ""
	}

	hash := int(h.hashFunc([]byte(key)))

	idx := sort.Search(len(h.keys), func(i int) bool {
		return h.keys[i] >= hash
	})

	return h.hashMap[h.keys[idx%len(h.keys)]]
}
