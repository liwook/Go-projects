package main

import (
	"encoding/json"
	"geerpcV1/codec"
	"geerpcV1/geerpc"
	"log"
	"net"
)

func startServer(addr chan string) {
	l, err := net.Listen("tcp", "localhost:10000")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	geerpc.Accept(l)
}

func main() {
	addr := make(chan string)
	go startServer(addr)

	// in fact, following code is like a simple geerpc client
	conn, _ := net.Dial("tcp", <-addr)
	defer conn.Close()

	// send options
	_ = json.NewEncoder(conn).Encode(geerpc.DefaultOption)
	cc := codec.NewGobCodec(conn)
	// send request & receive response
	for i := 0; i < 3; i++ {
		h := &codec.Header{
			ServiceMethod: "Foo.Sum",
			Seq:           uint64(i),
		}

		cc.WriteResponse(h, h.Seq) //把h.Seq替换成其他类型运行时gob解码会出错，例如：string，int等等
		cc.ReadHeader(h)
		var reply string
		cc.ReadBody(&reply)
		log.Println("reply:", reply)
	}
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
