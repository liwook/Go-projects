package cache

import (
	"fmt"
	"log"
	"sync"
)

type Group struct {
	name      string
	mainCache cache
	getter    Getter

	// peers *Peers //添加了节点集合
	peers *HTTPPool //添加了节点集合

}

var (
	rwMu   sync.RWMutex
	groups = make(map[string]*Group)
)

// 往分组内注册节点集合
// func (g *Group) RegisterPeers(peers *Peers) {
func (g *Group) RegisterPeers(peers *HTTPPool) {

	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

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

func (g *Group) load(key string) (value ByteView, err error) {
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			if value, err = g.getFromPeer(peer, key); err == nil {
				return value, nil
			}
			log.Println("[GeeCache] Failed to get from peer", err)
		}
	}

	return g.getLocally(key)

}

func (g *Group) getFromPeer(peer *httpGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil

}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneByte(bytes)}
	g.mainCache.add(key, value)
	return value, nil
}

// func (g *Group) load(key string) (ByteView, error) {
// 	bytes, err := g.getter.Get(key)
// 	if err != nil {
// 		return ByteView{}, err
// 	}
// 	value := ByteView{b: cloneByte(bytes)}
// 	g.mainCache.add(key, value)
// 	return value, nil
// }
