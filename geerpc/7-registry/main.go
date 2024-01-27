package main

import (
	"context"
	"fmt"
	"geerpcV7/geerpc"
	"geerpcV7/registry"
	"geerpcV7/xclient"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type My int

type Args struct{ Num1, Num2 int }

func (m *My) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (m *My) Sleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.Num1))
	*reply = args.Num1 + args.Num2
	return nil
}

func main() {
	registryAddr := "http://localhost:9999/_rpc_/registry"

	var wg sync.WaitGroup
	wg.Add(1)
	go startRegistry(&wg) //开启注册中心服务
	wg.Wait()
	time.Sleep(time.Second)

	wg.Add(2)
	go startServer(registryAddr, &wg)
	go startServer(registryAddr, &wg)
	wg.Wait()

	time.Sleep(time.Second)
	clientCall(registryAddr)
	broadcast(registryAddr)
}

func startServer(registryAddr string, wg *sync.WaitGroup) {
	var myServie My

	l, _ := net.Listen("tcp", "localhost:0") //端口是0表示端口随机
	server := geerpc.NewServer()
	//这里一定要用&myServie，因为前面Sum方法的接受者是*My;若接受者是My,myServie或者&myServie都可以
	server.Register(&myServie)
	registry.Heartbeat(registryAddr, "tcp@"+l.Addr().String(), 0) //定时发送心跳
	wg.Done()
	server.Accept(l)
}

func startRegistry(wg *sync.WaitGroup) {
	l, _ := net.Listen("tcp", "localhost:9999")
	registry.HandleHTTP()
	wg.Done()
	http.Serve(l, nil)
}

// 调用单个服务实例
func clientCall(registryAddr string) {
	// d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	d := xclient.NewGeeRegistryDiscovery(registryAddr, 0)
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer xc.Close()

	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var reply int = 1324
			if err := xc.Call(context.Background(), "My.Sum", &Args{Num1: i, Num2: i * i}, &reply); err != nil {
				log.Println("call Foo.Sum error:", err)
			}
			fmt.Println("reply: ", reply)
		}(i)
	}
	wg.Wait()
}

func broadcast(registryAddr string) {
	// d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	d := xclient.NewGeeRegistryDiscovery(registryAddr, 0)
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer xc.Close()

	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var reply int = 1324
			if err := xc.Broadcast(context.Background(), "My.Sum", &Args{Num1: i, Num2: i * i}, &reply); err != nil {
				fmt.Println("Broadcast call Foo.Sum error:", err)
			}
			fmt.Println("Broadcast reply: ", reply)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()
			var replyTimeout int = 1324
			if err := xc.Broadcast(ctx, "My.Sleep", &Args{Num1: i, Num2: i * i}, &replyTimeout); err != nil {
				fmt.Println("Broadcast call Foo.Sum error:", err)
			}
			fmt.Println("timeout Broadcast reply: ", replyTimeout)

		}(i)
	}
	wg.Wait()
}
