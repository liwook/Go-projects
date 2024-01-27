package geerpc

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"geerpcV7/codec"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// 定义一个请求
type Call struct {
	ServiceMethod string      // The name of the service and method to call.
	Args          interface{} // The argument to the function (*struct).
	Reply         interface{} // The reply from the function (*struct).
	Error         error       // After completion, the error status.
	Done          chan *Call  // Receives *Call when Go is complete.
	Seq           uint64
}

func (call *Call) done() {
	call.Done <- call
}

type Client struct {
	code     codec.Codec
	opt      *Option
	sending  sync.Mutex
	header   codec.Header
	mutex    sync.Mutex //保护下面的变量
	seq      uint64
	pending  map[uint64]*Call //存储未完成的请求，键是编号，值是 Call 实例
	closing  bool             //user has called Close
	shutdown bool             // server has told us to stop
}

var ErrShutdown = errors.New("connection is shut down")

func (client *Client) Close() error {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	if client.closing {
		return ErrShutdown
	}
	client.closing = true
	return client.code.Close()
}

func (client *Client) IsAvailable() bool {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	return !client.closing && !client.shutdown
}

func (client *Client) RegisterCall(call *Call) (uint64, error) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if client.closing || client.shutdown {
		return 0, ErrShutdown
	}

	call.Seq = client.seq //设置Call的序号
	client.pending[call.Seq] = call
	client.seq++
	return call.Seq, nil
}

func (client *Client) removeCall(seq uint64) *Call {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	call := client.pending[seq]
	delete(client.pending, seq)
	return call
}

func (client *Client) terminateCalls(err error) {
	client.sending.Lock()
	defer client.sending.Unlock()
	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.shutdown = true
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
}

func NewClient(conn net.Conn, opt *Option) (*Client, error) {
	// send options with server
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client: options error: ", err)
		conn.Close()
		return nil, err
	}
	f := codec.NewCodeFuncMap[opt.CodecType]
	if f == nil { //没有符合条件的编解码器
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		log.Println("rpc client: codec error:", err)
		return nil, err
	}

	return newClientCodec(f(conn), opt), nil
}

func newClientCodec(code codec.Codec, opt *Option) *Client {
	client := &Client{
		seq:     1,
		code:    code,
		opt:     opt,
		pending: make(map[uint64]*Call),
	}
	go client.receive()
	return client
}

func parseOptions(opts ...*Option) (*Option, error) {
	if len(opts) == 0 || opts[0] == nil {
		return DefaultOption, nil
	}
	if len(opts) != 1 {
		return nil, errors.New("number of options is more than 1")
	}
	opt := opts[0]
	opt.MagicNumber = DefaultOption.MagicNumber

	if opt.CodecType == "" {
		opt.CodecType = DefaultOption.CodecType
	}
	if _, ok := codec.NewCodeFuncMap[opt.CodecType]; !ok {
		return nil, fmt.Errorf("invalid codec type %s", opt.CodecType)
	}
	return opt, nil
}

type newClientFunc func(conn net.Conn, opt *Option) (client *Client, err error)

type clientResult struct {
	client *Client
	err    error
}

func dialTimeout(f newClientFunc, network, address string, opt *Option) (client *Client, err error) {
	//超时连接检测
	conn, err := net.DialTimeout(network, address, opt.ConnectTimeout)
	if err != nil {
		return nil, err
	}

	//设置超时时间的情况
	ch := make(chan clientResult)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func(ctx context.Context) {
		client, err = f(conn, opt)

		select {
		case <-ctx.Done():
			return
		default:
			ch <- clientResult{client: client, err: err}
		}
	}(ctx)

	if opt.ConnectTimeout == 0 {
		result := <-ch
		return result.client, result.err
	}

	select {
	case <-time.After(opt.ConnectTimeout):
		cancel() //超时通知子协程结束退出
		return nil, fmt.Errorf("rpc client: connect timeout: expect within %s", opt.ConnectTimeout)
	case result := <-ch:
		fmt.Printf("result.client :%p\n", result.client)
		return result.client, result.err
	}
}

func Dail(network, address string, opts ...*Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}

	return dialTimeout(NewClient, network, address, opt)
}

// 使用例子 client, err := rpc.Dial("tcp", "localhost:1234")
// func Dail(network, address string, opts ...*Option) (client *Client, err error) {
// 	opt, err := parseOptions(opts...)
// 	if err != nil {
// 		return nil, err
// 	}
// 	conn, err := net.Dial(network, address)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return NewClient(conn, opt)
// }

func (client *Client) send(call *Call) {
	// make sure that the client will send a complete request
	client.sending.Lock()
	defer client.sending.Unlock()

	//注册，添加到pending中
	seq, err := client.RegisterCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	//复用同一个header
	client.header.ServiceMethod = call.ServiceMethod
	client.header.Seq = seq
	client.header.Error = ""

	// encode and send the request
	if err := client.code.WriteResponse(&client.header, call.Args); err != nil {
		call := client.removeCall(seq)
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

func (client *Client) Go(serviceMethod string, args, reply any, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10) //10或1或其他的也可以的，大于0即可
	} else if cap(done) == 0 {
		log.Panic("rpc client: done channel is unbuffered")
	}

	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	client.send(call)
	return call
}

func (client *Client) Call(ctx context.Context, serviceMethod string, args, reply any) error {
	call := client.Go(serviceMethod, args, reply, make(chan *Call, 1))
	select {
	case <-ctx.Done():
		client.removeCall(call.Seq)
		return errors.New("rpc client: call failed: " + ctx.Err().Error())
	case call := <-call.Done:
		return call.Error
	}
	//之前的写法
	// 	call := <-client.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
	// 	return call.Error
}

func (client *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		if err = client.code.ReadHeader(&h); err != nil {
			break
		}

		call := client.removeCall(h.Seq)
		switch {
		case call == nil:
			err = client.code.ReadBody(nil)
		case h.Error != "":
			call.Error = fmt.Errorf(h.Error)
			err = client.code.ReadBody(nil)
			call.done()
		default:
			err = client.code.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body " + err.Error())
			}
			call.done()
		}
	}
	client.terminateCalls(err)
}

// HTTP部分
func NewHTTPClient(conn net.Conn, opt *Option) (*Client, error) {
	io.WriteString(conn, fmt.Sprintf("CONNECT %s HTTP/1.0\n\n", defaultRPCPath))

	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err == nil && resp.Status == connected {
		return NewClient(conn, opt)
	}
	if err != nil {
		err = errors.New("unexpected HTTP response: " + resp.Status)
	}
	return nil, err
}

func DialHTTP(network, address string, opts ...*Option) (*Client, error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	return dialTimeout(NewHTTPClient, network, address, opt)
}

// 统一的建立rpc客户端的接口
// rpcAddr格式 http@10.0.0.1:34232,tpc@10.0.0.1:10000
func XDial(rpcAddr string, opts ...*Option) (*Client, error) {
	parts := strings.Split(rpcAddr, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("rpc client err: wrong format '%s', expect protocol@addr", rpcAddr)
	}
	protocol, addr := parts[0], parts[1]

	switch protocol {
	case "http":
		return DialHTTP("tcp", addr, opts...)
	default:
		// tcp, unix or other transport protocol
		return Dail(protocol, addr, opts...)
	}
}
