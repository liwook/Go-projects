package main

import (
	"context"
	"fmt"
	"geerpcV5/geerpc"
	"log"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

type My int

type Args struct{ Num1, Num2 int }

func (m *My) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	// time.Sleep(time.Second * 3)
	return nil
}

func startServer(addrCh chan string) {
	var myServie My
	//这里一定要用&myServie，因为前面Sum方法的接受者是*My;若接受者是My,myServie或者&myServie都可以
	if err := geerpc.Register(&myServie); err != nil {
		slog.Error("register error:", err) //slog是Go官方的日志库
		os.Exit(1)
	}
	geerpc.HandleHTTP()
	addrCh <- "127.0.0.1:10000"
	log.Fatal(http.ListenAndServe("127.0.0.1:10000", nil))

	//之前的写法
	// l, err := net.Listen("tcp", "localhost:10000")
	// geerpc.Accept(l)
}

func clientCall(addrCh chan string) {
	addr := <-addrCh
	fmt.Println(addr)
	client, err := geerpc.DialHTTP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	num := 5
	var wg sync.WaitGroup
	wg.Add(num)

	for i := 0; i < num; i++ {
		go func(i int) {
			defer wg.Done()
			args := &Args{Num1: i, Num2: i * i}
			var reply int = 1324
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			if err := client.Call(ctx, "My.Sum", args, &reply); err != nil {
				log.Println("call Foo.Sum error:", err)
			}
			fmt.Println("reply: ", reply)
		}(i)
	}
	wg.Wait()
}

func main() {
	ch := make(chan string)
	go clientCall(ch)
	startServer(ch)
}
