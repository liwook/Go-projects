package geerpc

import (
	"fmt"
	"reflect"
	"testing"
)

type Hello int //服务结构体类型

type Args struct{ Num1, Num2 int } //方法的参数类型Args

func (h Hello) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

// 不能导出的，方法sum的首字母是小写
func (h Hello) sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func _assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+msg, v...))
	}
}

// 测试 newService 方法。
func TestNewService(t *testing.T) {
	var h Hello
	s := newService(&h)
	//判断s的方法的个数，是可导出的方法
	_assert(len(s.method) == 1, "wrong service Method, expect 1, but got %d", len(s.method))

	mType := s.method["Sum"]
	//判断是否有这个方法
	_assert(mType != nil, "wrong Method, Sum shouldn't nil")
}

// 测试 call 方法。
func TestMethodType_Call(t *testing.T) {
	var h Hello
	s := newService(h)
	mType := s.method["Sum"]

	//new 方法的参数和返回结果
	argv := mType.newArgv()
	reply := mType.newReplyv()
	//设置参数值
	argv.Set(reflect.ValueOf(Args{Num1: 3, Num2: 47}))
	err := s.call(mType, argv, reply)
	_assert(err == nil && *reply.Interface().(*int) == 50, "failed to call Foo.Sum")
}
