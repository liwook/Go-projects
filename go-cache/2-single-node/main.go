package main

import (
	"cacheV2/cache"
	"fmt"
)

// 缓存中没有的话，就从该db中查找
var db = map[string]string{
	"tom":  "100",
	"jack": "200",
	"sam":  "444",
}

// 统计某个键调用回调函数的次数
var loadCounts = make(map[string]int, len(db))

func main() {
	//传函数入参
	cache := cache.NewGroup("scores", 2<<10, cache.GetterFunc(funcCbGet))
	//传结构体入参，也可以
	// cbGet := &search{}
	// cache := cache.NewGroup("scores", 2<<10, cbGet)

	for k, v := range db {
		if view, err := cache.Get(k); err != nil || view.String() != v {
			fmt.Println("failed to get value of Tom")
		}

		if _, err := cache.Get(k); err != nil || loadCounts[k] > 1 {
			fmt.Printf("cache %s miss", k)
		}
	}

	if view, err := cache.Get("unknown"); err == nil {
		fmt.Printf("the value of unknow should be empty, but %s got", view)
	} else {
		fmt.Println(err)
	}
}

// 函数的
func funcCbGet(key string) ([]byte, error) {
	fmt.Println("callback search key: ", key)
	if v, ok := db[key]; ok {
		if _, ok := loadCounts[key]; !ok {
			loadCounts[key] = 0
		}
		loadCounts[key] += 1
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
		if _, ok := loadCounts[key]; !ok {
			loadCounts[key] = 0
		}
		loadCounts[key] += 1
		return []byte(v), nil
	}
	return nil, fmt.Errorf("%s not exit", key)
}
