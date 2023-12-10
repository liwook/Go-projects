package lru

import (
	"fmt"
	"reflect"
	"testing"
)

// 该类型实现了NodeValue接口
type String string

func (d String) Len() int {
	return len(d)
}

func TestGet(t *testing.T) {
	lru := New(0, nil)
	lru.Add("key1", String("abc"))
	if v, ok := lru.Get("key1"); !ok || String(v.(String)) != "abc" {
		t.Fatalf("cache hit key1=1234 failed")
	}

	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestRemoveoldest(t *testing.T) {
	k1, k2, k3 := "k1", "k2", "k3"
	v1, v2, v3 := "v1", "v2", "v3"

	cap := len(k1 + k2 + v1 + v2)
	lru := New(int64(cap), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))
	//容量只够k1和k2,现在也添加了k3，那需要把k1删除
	if _, ok := lru.Get("k1"); ok || lru.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}

// 测试回调函数
func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value NodeValue) {
		keys = append(keys, key)
		fmt.Println("callback..")
	}
	lru := New(10, callback)
	lru.Add("key1", String("123456"))
	lru.Add("k2", String("k2"))
	lru.Add("k3", String("k3"))
	lru.Add("k4", String("k4"))

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}
