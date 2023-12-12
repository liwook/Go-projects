package cache

import (
	"cacheV6/cache/consistenthash"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const DefaultBasePath = "/geecache/"

type HTTPPool struct {
	addr     string
	basePath string

	//新添加的，把Peers内容增添到HTTPPool中
	mutex         sync.Mutex
	peersHashRing *consistenthash.HashRing
	httpGetters   map[string]*httpGetter
}

func (pool *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, pool.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	// fmt.Println(r.URL.Path)

	parts := strings.SplitN(r.URL.Path[len(pool.basePath):], "/", 2)

	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(parts[1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

// 添加节点
// 使用用例：Set("http://localhost:8001","http://localhost:8001")
func (p *HTTPPool) Set(peers ...string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.peersHashRing = consistenthash.NewHash(50, nil)
	p.peersHashRing.Add(peers...) //在 hash 环上添加真实节点和虚拟节点
	//存储远端节点信息
	p.httpGetters = make(map[string]*httpGetter)
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

func (p *HTTPPool) PickPeer(key string) (*httpGetter, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	//这里返回的peer是个地址，可以查看(Peers).Set函数中的参数
	var peer string
	if peer = p.peersHashRing.Get(key); peer != "" && peer != p.addr {
		log.Println("pick peer ", peer, "   local=", p.addr)
		return p.httpGetters[peer], true
	}
	// if peer != "" && peer == p.addr {
	// 	log.Println("local ok, this peer=", peer)
	// }
	return &httpGetter{}, false
}

func NewHTTPPool(addr string, basePath string) *HTTPPool {
	return &HTTPPool{
		addr:     addr,
		basePath: basePath,
	}
}

type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	//QueryEscape 对字符串进行转义，以便可以将其安全地放置在 URL 查询中。
	u := fmt.Sprintf("%v%v/%v", h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key))

	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	return bytes, nil
}

// type Peers struct {
// 	addr          string //这个是用于进行选择节点时用来判断的
// 	basePath      string
// 	mutex         sync.Mutex
// 	peersHashRing *consistenthash.HashRing
// 	httpGetters   map[string]*httpGetter
// }

// // 添加节点
// // 使用用例：Set("http://localhost:8001","http://localhost:8001")
// func (p *Peers) Set(peers ...string) {
// 	p.mutex.Lock()
// 	defer p.mutex.Unlock()

// 	p.peersHashRing = consistenthash.NewHash(50, nil)
// 	p.peersHashRing.Add(peers...) //在 hash 环上添加真实节点和虚拟节点
// 	//存储远端节点信息
// 	p.httpGetters = make(map[string]*httpGetter)
// 	for _, peer := range peers {
// 		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
// 	}
// }

// func (p *Peers) PickPeer(key string) (*httpGetter, bool) {
// 	p.mutex.Lock()
// 	defer p.mutex.Unlock()
// 	//这里返回的peer是个地址，可以查看(Peers).Set函数中的参数
// 	if peer := p.peersHashRing.Get(key); peer != "" && peer != p.addr {
// 		fmt.Println("pick peer ", peer)
// 		return p.httpGetters[peer], true
// 	}
// 	return &httpGetter{}, false
// }
