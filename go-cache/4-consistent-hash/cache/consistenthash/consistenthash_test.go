package consistenthash

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	//创建哈希环，每个真实节点有三个虚拟节点
	hash := NewHash(2, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})

	//添加3个真实节点，哈希函数后，
	//"2"对应的虚拟节点是2/12/22,4的是/4/14/24
	hash.Add("2", "4")

	//map的key是缓存数据key,value是真实节点
	testCases := map[string]string{
		"4":  "4",
		"11": "2",
		"16": "2",
		"27": "2",
	}

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}
	//添加真实节点"8"，其对应的虚拟节点是8/18
	hash.Add("8")

	testCases["16"] = "8"
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}
}
