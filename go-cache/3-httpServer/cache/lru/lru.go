package lru

import "container/list"

type Cache struct {
	maxBytes  int64      //允许的能使用的最大内存
	nbytes    int64      //已使用的内存
	ll        *list.List //双向链表
	cache     map[string]*list.Element
	OnEvicted func(key string, value NodeValue)
}

type NodeValue interface {
	Len() int
}

//代表双向链表的数据类型
type entry struct {
	key   string
	value NodeValue
}

func New(maxBytes int64, onEvicted func(string, NodeValue)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

//查找
func (c *Cache) Get(key string) (v NodeValue, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele) //取到该元素,成为了热点数据, 所以要往队首放
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

//添加
func (c *Cache) Add(key string, value NodeValue) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else { //还没存在该数据的情况
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}

	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele == nil {
		return
	}

	c.ll.Remove(ele)
	kv := ele.Value.(*entry)
	delete(c.cache, kv.key) //从map中删除
	c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())

	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value) //执行被删除时的回调函数
	}
}

// Len the number of cache entries
func (c *Cache) Len() int {
	return c.ll.Len()
}
