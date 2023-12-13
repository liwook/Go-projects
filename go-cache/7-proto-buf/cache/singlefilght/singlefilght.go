package singlefilght

import (
	"sync"
)

type call struct {
	wg  sync.WaitGroup // 控制线程是否等待
	val any            // 请求返回结果
	err error          // 返回的错误信息
}

type Group struct {
	mutex   sync.Mutex
	callMap map[string]*call
}

func (g *Group) Do(key string, fn func() (any, error)) (any, error) {
	g.mutex.Lock()
	if g.callMap == nil {
		g.callMap = make(map[string]*call)
	}
	if c, ok := g.callMap[key]; ok {
		g.mutex.Unlock()
		c.wg.Wait()         //如果有请求正在进行中，那就等待
		return c.val, c.err //请求结束，返回结果
	}

	c := new(call)
	c.wg.Add(1)        //发起请求前加锁
	g.callMap[key] = c //添加到map中，表明该key已有对应的请求在处理
	g.mutex.Unlock()

	c.val, c.err = fn() //调用fn,即是访问key的函数
	c.wg.Done()         //请求结束

	g.mutex.Lock()
	delete(g.callMap, key) //删除该key,表示该key当前没有请求在处理
	g.mutex.Unlock()

	return c.val, c.err
}

func (g *Group) notMutexDo(key string, fn func() (any, error)) (any, error) {
	if c, ok := g.callMap[key]; ok {
		c.wg.Wait()         //如果有请求正在进行中，那就等待
		return c.val, c.err //请求结束，返回结果
	}

	c := new(call)
	c.wg.Add(1)        //发起请求前加锁
	g.callMap[key] = c //添加到map中，表明该key已有对应的请求在处理

	c.val, c.err = fn() //调用fn,即是访问key的函数
	c.wg.Done()         //请求结束

	delete(g.callMap, key) //删除该key,表示该key当前没有请求在处理
	return c.val, c.err
}
