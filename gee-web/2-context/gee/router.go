package gee

import "net/http"

type HandlerFunc func(*Context)

type router struct {
	handers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{handers: make(map[string]HandlerFunc)}
}

// 添加路由
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	key := method + "-" + pattern
	r.handers[key] = handler
}

func (r *router) handle(c *Context) {
	key := c.Method + "-" + c.Path
	if hander, ok := r.handers[key]; ok {
		hander(c)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}
