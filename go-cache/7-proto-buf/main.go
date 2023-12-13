package main

import (
	"cacheV6/cache"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

// 缓存中没有的话，就从该db中查找
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	gee := createGroup()
	if api {
		go startAPIServer(apiAddr, gee)
	}
	startCacheServer(addrMap[port], addrs, gee)
	time.Sleep(time.Second * 1000)
}

func createGroup() *cache.Group {
	return cache.NewGroup("scores", 2<<10, cache.GetterFunc(func(key string) ([]byte, error) {
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exit", key)
	}))
}

func startCacheServer(addr string, addrs []string, groups *cache.Group) {
	//HTTPPool是节点结合和HTTP服务端
	peers := cache.NewHTTPPool(addr, cache.DefaultBasePath)
	peers.Set(addrs...)         //添加节点
	groups.RegisterPeers(peers) //注册节点集合
	log.Println("geecache is running at", addr)
	http.ListenAndServe(addr[7:], peers)
}

func startAPIServer(apiAddr string, groups *cache.Group) {
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		view, err := groups.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(view.ByteSlice())
	})

	log.Println("fontend server is running at", apiAddr)
	http.ListenAndServe(apiAddr[7:], nil)
}
