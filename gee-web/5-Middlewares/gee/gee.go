package gee

import (
	"net/http"
	"strings"
)

// type HandlerFunc func(http.ResponseWriter, *http.Request)

type Engine struct {
	*RouterGroup
	router *router
	gorups []*RouterGroup //存储所有的路由组
}

type RouterGroup struct {
	prefix      string //前缀
	middlewares []HandlerFunc
	engine      *Engine //所有的路由组共享一个enginx实例
}

func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		engine: engine,
	}
	engine.gorups = append(engine.gorups, newGroup)
	return newGroup
}

func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

func (group *RouterGroup) addRoute(method string, pattern string, hander HandlerFunc) {
	path := group.prefix + pattern
	group.engine.router.addRoute(method, path, hander)
}

// GET defines the method to add GET request
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

// 创建enginx实例
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.gorups = []*RouterGroup{engine.RouterGroup}
	return engine
}

func (engine *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// c := newContext(w, req)
	// engine.router.handle(c)

	var middlewares []HandlerFunc
	for _, group := range engine.gorups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...) //添加该路由组的中间件
		}
	}
	c := newContext(w, req)
	c.midHandlers = middlewares
	engine.router.handle(c)
}
