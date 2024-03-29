package geerpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"geerpcV5/codec"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

const (
	MagicNumber = 0x3b3f5c

	connected        = "200 Connected to RPC"
	defaultRPCPath   = "/myrpc"
	defaultDebugPath = "/debug/rpc"
)

type Option struct {
	MagicNumber    int            // MagicNumber marks this's a geerpc request
	CodecType      codec.CodeType // client may choose different Codec to encode body
	ConnectTimeout time.Duration  //0 表示没有限制
	HandleTimeout  time.Duration
}

var DefaultOption = &Option{
	MagicNumber:    MagicNumber,
	CodecType:      codec.GobType,
	ConnectTimeout: time.Second * 10, //默认连接超时是10s
}

type Server struct {
	serviceMap sync.Map
}

type request struct {
	h *codec.Header
	// argv, replyv reflect.Value
	argv, replyv reflect.Value
	mtype        *methodType
	svc          *service

	//这是之前的
	// h *codec.Header
	// // argv, replyv reflect.Value
	// requestData uint64
	// replyData   string
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func (server *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		// 拿到客户端的连接, 开启新协程异步去处理.
		go server.ServeConn(conn)
	}
}

func Accept(lis net.Listener) { DefaultServer.Accept(lis) }

func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	defer conn.Close()

	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error: ", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}

	//目前只实现了gob编解码
	f := codec.NewCodeFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}

	server.servCode(f(conn), &opt)
}

var invalidRequest = struct{}{}

func (server *Server) servCode(cc codec.Codec, opt *Option) {
	sending := new(sync.Mutex)
	// wg := new(sync.WaitGroup)
	var wg sync.WaitGroup

	for {
		req, err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go server.handleRequest(cc, req, sending, &wg, opt.HandleTimeout)
	}
	wg.Wait()
	cc.Close()
}

func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := server.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}

	req := &request{h: h}
	req.svc, req.mtype, err = server.findService(h.ServiceMethod) //在此处使用findService

	//创建方法参数和返回值，new出来的
	req.argv = req.mtype.newArgv()
	req.replyv = req.mtype.newReplyv()

	// make sure that argvi is a pointer, ReadBody need a pointer as parameter
	argvi := req.argv.Interface() //使用Interface()方法是为了req.argv转回any类型， cc.ReadBody入参需要的
	if req.argv.Type().Kind() != reflect.Pointer {
		argvi = req.argv.Addr().Interface()
	}

	if err := cc.ReadBody(argvi); err != nil {
		log.Println("rpc server: read argv err:", err)
	}
	return req, err

	//之前的写法
	// req.requestData = reflect.New(reflect.TypeOf(uint64(1)))
	// err = cc.ReadBody(req.requestData.Interface())

	// TODO: now we don't know the type of request argv
	//这一章节，我们只能处理用户发送过来的uint64类型的数据
	// if err = cc.ReadBody(&req.requestData); err != nil {
	// 	log.Println("rpc server: read argv err:", err)
	// }
	// return req, nil
}

func (server *Server) sendResponse(cc codec.Codec, h *codec.Header, body any, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.WriteResponse(h, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	called := make(chan struct{})
	go func(ctx context.Context) {
		err := req.svc.call(req.mtype, req.argv, req.replyv)
		select {
		case <-ctx.Done():
			req.h.Error = fmt.Sprintf("rpc server: request handle timeout: expect within %s", timeout)
			server.sendResponse(cc, req.h, invalidRequest, sending)
		default:
			if err != nil {
				fmt.Println("call err:", err)
				req.h.Error = err.Error()
				server.sendResponse(cc, req.h, invalidRequest, sending)
			} else {
				server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
			}
			called <- struct{}{}
		}
	}(ctx)
	if timeout == 0 {
		<-called
		return
	}
	select {
	case <-time.After(timeout):
		cancel()
	case <-called:
		return
	}
}

// 注册服务
func (server *Server) Register(rcvr any) error {
	s := newService(rcvr)
	//如果获取的 key 存在，就返回 key 对应的元素，
	//若获取的 key 不存在，就返回我们设置的值，并且将我们设置的值，存入 map
	if _, dup := server.serviceMap.LoadOrStore(s.name, s); dup {
		return errors.New("rpc: service already defined: " + s.name)
	}
	return nil
}

// Register publishes the receiver's methods in the DefaultServer.
func Register(rcvr interface{}) error { return DefaultServer.Register(rcvr) }

func (server *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	//serviceMethod例子 "myservice.say"
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server: service/method request ill-formed: " + serviceMethod)
		return
	}
	//获取服务名字和方法名
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]

	//Load是sync.Map获取value的方法，返回值类型是any
	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can't find service " + serviceName)
		return
	}

	svc = svci.(*service)          //这个是any类型转成*service类型
	mtype = svc.method[methodName] //找到对应的 methodType
	if mtype == nil {
		err = errors.New("rpc server: can't find method " + methodName)
	}
	return
}

// HTTP

// server HTTP部分,server实现了ServeHTTP方法，就是http.Handler接口了
func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 must CONNECT\n")
		fmt.Println(req.Method)
		return
	}

	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Print("rpc hijacking ", req.RemoteAddr, " :", err.Error())
	}

	io.WriteString(conn, "HTTP/1.0 "+connected+"\n\n")
	//server.ServeConn(conn)就回到了之前的accept后的那部分
	server.ServeConn(conn)
}

func HandleHTTP() {
	DefaultServer.HandleHTTP()
}

func (server *Server) HandleHTTP() {
	//方法原型func Handle(pattern string, handler Handler)
	http.Handle(defaultRPCPath, server)
	http.Handle(defaultDebugPath, debugHTTP{Server: server}) //这个是处理debug的
}
