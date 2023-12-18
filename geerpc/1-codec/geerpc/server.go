package geerpc

import (
	"encoding/json"
	"fmt"
	"geerpcV1/codec"
	"io"
	"log"
	"net"
	"sync"
)

const MagicNumber = 0x3b3f5c

type Option struct {
	MagicNumber int            // MagicNumber marks this's a geerpc request
	CodecType   codec.CodeType // client may choose different Codec to encode body
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

type Server struct {
}

type request struct {
	h *codec.Header
	// argv, replyv reflect.Value
	requestData uint64
	replyData   string
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
	if opt.CodecType != codec.GobType {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	srv := codec.NewGobCodec(conn)
	server.servCode(srv)
}

var invalidRequest = struct{}{}

func (server *Server) servCode(cc codec.Codec) {
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
		go server.handleRequest(cc, req, sending, &wg)
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

	// req.requestData = reflect.New(reflect.TypeOf(uint64(1)))
	// err = cc.ReadBody(req.requestData.Interface())

	// TODO: now we don't know the type of request argv
	//这一章节，我们只能处理用户发送过来的uint64类型的数据
	if err = cc.ReadBody(&req.requestData); err != nil {
		log.Println("rpc server: read argv err:", err)
	}
	return req, nil
}
func (server *Server) sendResponse(cc codec.Codec, h *codec.Header, body any, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.WriteResponse(h, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("handleRequest ", req.h, req.requestData)
	// req.replyv = reflect.ValueOf(fmt.Sprintf("geerpc resp %d", req.h.Seq))
	// server.sendResponse(cc, req.h, req.replyv.Interface(), sending)

	req.replyData = fmt.Sprintf(" ok my resp %d", req.h.Seq)
	server.sendResponse(cc, req.h, &req.replyData, sending)
}
