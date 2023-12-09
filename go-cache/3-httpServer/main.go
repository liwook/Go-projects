package main

import (
	"cacheV3/cache"
	"fmt"
	"log"
	"net/http"
)

// 缓存中没有的话，就从该db中查找
var db = map[string]string{
	"tom":  "100",
	"jack": "200",
	"sam":  "444",
}

func main() {
	//传函数入参
	cache.NewGroup("scores", 2<<10, cache.GetterFunc(funcCbGet))
	//传结构体入参，也可以
	// cbGet := &search{}
	// cache.NewGroup("scores", 2<<10, cbGet)

	addr := "localhost:10000"
	peers := cache.NewHTTPPool(addr, cache.DefaultBasePath)
	log.Fatal(http.ListenAndServe(addr, peers))
}

// 函数的
func funcCbGet(key string) ([]byte, error) {
	fmt.Println("callback search key: ", key)
	if v, ok := db[key]; ok {

		return []byte(v), nil
	}
	return nil, fmt.Errorf("%s not exit", key)
}

// 结构体，实现了Getter接口的Get方法，
type search struct {
}

func (s *search) Get(key string) ([]byte, error) {
	fmt.Println("struct callback search key: ", key)
	if v, ok := db[key]; ok {
		return []byte(v), nil
	}
	return nil, fmt.Errorf("%s not exit", key)
}
