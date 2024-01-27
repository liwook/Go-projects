package geerpc

import (
	"go/ast"
	"log"
	"reflect"
)

// 定义方法
type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	replyType reflect.Type
	numCalls  uint64
}

func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value
	if m.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

func (m *methodType) newReplyv() reflect.Value {
	// reply must be a pointer type
	replyv := reflect.New(m.replyType.Elem())
	switch m.replyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.replyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.replyType.Elem(), 0, 0))
	}
	return replyv
}

// 定义服务
type service struct {
	name   string
	typ    reflect.Type
	rcvr   reflect.Value
	method map[string]*methodType
}

// 语法 func Indirect(v Value) Value
// Indirect返回v持有的指针指向的值的Value封装。若v持有的值为nil，会返回Value零值。若v持有的变量不是指针，那么将返回原值v
func newService(rcvr any) *service {
	s := new(service)
	s.rcvr = reflect.ValueOf(rcvr)
	//通过reflect.Value.Type.Name()获取结构体名
	//,但是当reflect.Value是指针时候，Name()返回空字符串。所以要先通过Indirect取指针的值
	s.name = reflect.Indirect(s.rcvr).Type().Name()
	s.typ = reflect.TypeOf(rcvr)
	//该函数是判断s.name是否是以大写字母开头的
	if !ast.IsExported(s.name) {
		log.Fatalf("rpc server: %s is not a valid service name", s.name)
	}
	s.registerMethods() //判断方法是否符合条件的
	return s
}

func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)
	//reflect.Type.NumMethod()是获取该结构体的方法个数，只能获取能导出的(方法首字母大写的)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i) //reflect.Type.Method()获取对应类型对应的方法
		mType := method.Type      //获取方法类型
		//reflect.Type.NumIn()是获取参数个数，NumOut()是返回值个数
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		//reflect.Type.Out()是返回值类型， 判断返回值是不是error类型
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}

		argType, replyType := mType.In(1), mType.In(2) //In()是方法参数类型
		// 响应值必须为可导出或者内置类型
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}

		s.method[method.Name] = &methodType{
			method:    method,
			ArgType:   argType,
			replyType: replyType,
		}
		log.Printf("rpc server: register %s.%s\n", s.name, method.Name)
	}
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

// 调用服务方法
func (s *service) call(m *methodType, argv, replyv reflect.Value) error {
	f := m.method.Func //m.method是reflect.method类型
	returnValues := f.Call([]reflect.Value{s.rcvr, argv, replyv})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error) //从Value转为原始数据类型
	}
	return nil
}
