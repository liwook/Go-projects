package geerpc

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"
)

func TestClient_dialTimeout(t *testing.T) {
	t.Parallel() //表示该测试将与（并且仅与）其他并行测试并行运行。

	l, _ := net.Listen("tcp", "localhost:10000")

	f := func(conn net.Conn, opt *Option) (*Client, error) {
		conn.Close()
		time.Sleep(time.Second * 2)
		return nil, nil
	}
	//命令行执行 go test -run TestClient_dialTimeout/timeout 测试
	t.Run("timeout", func(t *testing.T) {
		_, err := dialTimeout(f, "tcp", l.Addr().String(), &Option{ConnectTimeout: time.Second})
		_assert(err != nil && strings.Contains(err.Error(), "connect timeout"), "expect a timeout error")
	})
	//命令行执行 go test -run TestClient_dialTimeout/0 测试
	t.Run("0", func(t *testing.T) {
		_, err := dialTimeout(f, "tcp", l.Addr().String(), &Option{ConnectTimeout: 0})
		_assert(err == nil, "0 means no limit")
	})
}

// 测试用例2
type Bar int

func (b *Bar) Timeout(argv int, reply *int) error {
	time.Sleep(time.Second * 3) // 模拟3s的工作
	return nil
}

func startServer(addr chan string) {
	var b Bar
	_ = Register(&b)
	l, _ := net.Listen("tcp", "localhost:10000")
	addr <- l.Addr().String()
	Accept(l)
}

func TestClient_Call(t *testing.T) {
	t.Parallel()
	addrCh := make(chan string)
	go startServer(addrCh)
	addr := <-addrCh

	time.Sleep(time.Second)

	t.Run("client_timeout", func(t *testing.T) {
		client, _ := Dail("tcp", addr)
		ctx, _ := context.WithTimeout(context.Background(), time.Second*1)
		var reply int
		err := client.Call(ctx, "Bar.Timeout", 1, &reply)
		_assert(err != nil && strings.Contains(err.Error(), ctx.Err().Error()), "expect a timeout error")
	})

	t.Run("server_hander_timeout", func(t *testing.T) {
		client, _ := Dail("tcp", addr, &Option{
			HandleTimeout: time.Second,
		})
		var reply int
		err := client.Call(context.Background(), "Bar.Timeout", 1, &reply)
		_assert(err != nil && strings.Contains(err.Error(), "handle timeout"), "expect a timeout error")
	})
}
