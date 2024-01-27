package xclient

import (
	"context"
	"geerpcV7/geerpc"
	"reflect"
	"sync"
)

// xclient.go
type XClient struct {
	d       Discovery
	mode    SelectMode
	opt     *geerpc.Option
	mutex   sync.Mutex
	clients map[string]*geerpc.Client
}

func NewXClient(d Discovery, mode SelectMode, opt *geerpc.Option) *XClient {
	return &XClient{
		d:       d,
		mode:    mode,
		opt:     opt,
		clients: make(map[string]*geerpc.Client),
	}
}

func (xc *XClient) Close() error {
	xc.mutex.Lock()
	defer xc.mutex.Unlock()
	for key, client := range xc.clients {
		//只是关闭，没有其他的对错误的处理
		client.Close()
		delete(xc.clients, key)
	}
	return nil
}

func (xc *XClient) dial(rpcAddr string) (*geerpc.Client, error) {
	xc.mutex.Lock()
	defer xc.mutex.Unlock()

	client, ok := xc.clients[rpcAddr]
	if ok && !client.IsAvailable() {
		client.Close()
		delete(xc.clients, rpcAddr)
		client = nil
	}
	if client == nil {
		var err error
		client, err = geerpc.XDial(rpcAddr, xc.opt)
		if err != nil {
			return nil, err
		}

		xc.clients[rpcAddr] = client

	}

	return client, nil
}

func (xc *XClient) call(rpcAddr string, ctx context.Context, serviceMethod string, args, reply interface{}) error {
	//获取sokcet连接(复用)
	client, err := xc.dial(rpcAddr)
	if err != nil {
		return err
	}
	return client.Call(ctx, serviceMethod, args, reply)
}

// serviceMethod 例子："Foo.SUM"
func (xc *XClient) Call(ctx context.Context, serviceMethod string, args, reply any) error {
	//通过负载均衡策略得到服务实例
	rpcAddr, err := xc.d.Get(xc.mode)
	if err != nil {
		return err
	}

	return xc.call(rpcAddr, ctx, serviceMethod, args, reply)
}

func (xc *XClient) Broadcast(ctx context.Context, serviceMethod string, args, reply any) error {
	//获取所有的服务实例
	servers, err := xc.d.GetAll()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var mutex sync.Mutex //protect e and replyDone
	var e error

	replyDone := reply == nil
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, rpcAddr := range servers {
		wg.Add(1)
		//fmt.Printf("rpcAddrstring addr: %p\n", &rpcAddr) //其rpcAddr的地址都是一样的
		go func(rpcAddr string) {
			defer wg.Done()
			var clonedReply any
			if reply != nil {
				//reply是指针的，所以需要使用Elem()
				clonedReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			//xc.call方法中的参数clonedReply不能使用reply
			err := xc.call(rpcAddr, ctx, serviceMethod, args, clonedReply)
			mutex.Lock()
			defer mutex.Unlock()

			if err != nil && e == nil { //e==nil表明e还没有被赋值
				e = err
				cancel() // if any call failed, cancel unfinished calls
			}
			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(clonedReply).Elem())
				replyDone = true
			}
		}(rpcAddr)
	}
	wg.Wait()
	return e
}

// //另一种写法，go协程中没有参数
// func (xc *XClient) Broadcast(ctx context.Context, serviceMethod string, args, reply any) error {
//     .............
// 	for _, rpcAddr := range servers {
// 		wg.Add(1)
// 		//fmt.Printf("rpcAddrstring addr: %p\n", &rpcAddr) //其rpcAddr的地址都是一样的
//         addr:=rpcAddr
// 		go func() {
// 			defer wg.Done()
// 		err := xc.call(addr, ctx, serviceMethod, args, clonedReply)
//          ......
// 		}()
// 	}
//     ................
// }
