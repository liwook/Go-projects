package main

func main() {

}

// type My int

// type Args struct{ Num1, Num2 int }

// func (m *My) Sum(args Args, reply *int) error {
// 	*reply = args.Num1 + args.Num2
// 	return nil
// }

// func startServer(addr chan string) {
// 	//注册服务
// 	var myServie My
// 	//这里一定要用&myServie，因为前面Sum方法的接受者是*My;若接受者是My,myServie或者&myServie都可以
// 	if err := geerpc.Register(&myServie); err != nil {
// 		slog.Error("register error:", err) //slog是Go官方建议的日志库
// 		os.Exit(1)
// 	}
// 	//启动服务端
// 	l, err := net.Listen("tcp", "localhost:10000")
// 	if err != nil {
// 		slog.Error("network error:", err)
// 		os.Exit(1)
// 	}

// 	slog.Info("start rpc server on " + l.Addr().String())
// 	addr <- l.Addr().String()
// 	geerpc.Accept(l)
// }

// func main() {
// 	addr := make(chan string)
// 	go startServer(addr)

// 	client, _ := geerpc.Dail("tcp", <-addr)
// 	defer client.Close()
// 	time.Sleep(time.Second * 1)
// 	num := 3
// 	var wg sync.WaitGroup
// 	wg.Add(num)

// 	for i := 0; i < num; i++ {
// 		go func(i int) {
// 			defer wg.Done()
// 			args := &Args{Num1: i, Num2: i * i}
// 			var reply int
// 			if err := client.Call("My.Sum", args, &reply); err != nil {
// 				log.Fatal("call Foo.Sum error:", err)
// 			}
// 			fmt.Println("reply: ", reply)
// 		}(i)
// 	}
// 	wg.Wait()
// }

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
