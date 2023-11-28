package gee

import (
	"log"
	"net/http"
)

// type HandlerFunc func(http.ResponseWriter, *http.Request)

type Engine struct {
	router *router
}

// 创建enginx实例
func New() *Engine {
	return &Engine{router: newRouter()}
}

func (engine *Engine) addRoute(method string, pattern string, hander HandlerFunc) {
	engine.router.addRoute(method, pattern, hander)
	log.Printf("Route %4s - %s", method, pattern)
}

func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.addRoute("GET", pattern, handler)
}
func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRoute("POST", pattern, handler)
}

func (engine *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	engine.router.handle(c)
}
