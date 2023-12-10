package cache

import (
	"fmt"
	"sync"
)

type Group struct {
	name      string
	mainCache cache
	getter    Getter
}

var (
	rwMu   sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	rwMu.Lock()
	defer rwMu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// 获取 Group 对象的方法
func GetGroup(name string) *Group {
	rwMu.RLock()
	defer rwMu.RUnlock()
	g := groups[name]
	return g
}

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneByte(bytes)}
	g.mainCache.add(key, value)
	return value, nil
}
