package main

import (
	"geerpcV2/geerpc"
	"log"
	"net"
	"sync"
	"time"
)

func main() {
	addr := make(chan string)
	go startServer(addr)

	// in fact, following code is like a simple geerpc client
	client, _ := geerpc.Dail("tcp", <-addr) //上一节是使用net.Dail
	defer client.Close()
	time.Sleep(time.Second * 1)
	num := 3
	var wg sync.WaitGroup
	wg.Add(num)

	for i := 0; i < num; i++ {
		go func(i int) {
			defer wg.Done()
			args := uint64(i)
			var reply string
			if err := client.Call("foo.sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Println("reply: ", reply)
		}(i)
	}
	wg.Wait()
}

func startServer(addr chan string) {
	l, err := net.Listen("tcp", "localhost:10000")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	geerpc.Accept(l)
}

// 使用net/rpc库的简单例子
// type TestNetRpc struct{}

// func (t *TestNetRpc) Hello(request string, response *string) error {
// 	*response = "hello, " + request
// 	return nil
// }

// func startRPCServer() {
// 	rpc.Register(new(TestNetRpc)) // 注册服务

// 	lis, err := net.Listen("tcp", "localhost:9909")
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	fmt.Println(lis.Addr())
// 	rpc.Accept(lis) // 监听端口
// }

// func main() {
// 	go startRPCServer()
// 	time.Sleep(time.Second * 3)
// 	client, err := rpc.Dial("tcp", "localhost:9909")
// 	if err != nil {
// 		fmt.Println("dialing:", err)
// 	}
// 	var reply string
// 	err = client.Call("TestNetRpc.Hello", "hello", &reply)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	fmt.Println("reply:", reply)
// }
