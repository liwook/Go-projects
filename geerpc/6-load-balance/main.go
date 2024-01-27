package main

import (
	"context"
	"fmt"
	"geerpcV6/geerpc"
	"geerpcV6/xclient"
	"log"
	"log/slog"
	"net"
	"os"
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

func startServer(addrCh chan string) {
	var myServie My
	server := geerpc.NewServer()

	//这里一定要用&myServie，因为前面Sum方法的接受者是*My;若接受者是My,myServie或者&myServie都可以
	if err := server.Register(&myServie); err != nil {
		slog.Error("register error:", err) //slog是Go官方的日志库
		os.Exit(1)
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println("listen bad: ", err)
		os.Exit(1)
	}
	addrCh <- l.Addr().String()
	server.Accept(l)
}

// 调用单个服务实例
func clientCall(addr1, addr2 string) {
	d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
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

func broadcast(addr1, addr2 string) {
	d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
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

func main() {
	ch1 := make(chan string)
	ch2 := make(chan string)

	//start two servers
	go startServer(ch1)
	go startServer(ch2)

	addr1 := <-ch1
	addr2 := <-ch2
	time.Sleep(time.Second)
	clientCall(addr1, addr2)
	broadcast(addr1, addr2)
}
