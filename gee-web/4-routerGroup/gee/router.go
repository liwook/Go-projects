package gee

import (
	"fmt"
	"net/http"
	"strings"
)

type HandlerFunc func(*Context)

type router struct {
	handers map[string]HandlerFunc
	root    map[string]*node //key是GET，POST等请求方法
}

func newRouter() *router {
	return &router{
		handers: make(map[string]HandlerFunc),
		root:    make(map[string]*node),
	}
}

//	func parsePath(path string) (parts []string) {
//		par := strings.Split(path, "/")
//		for _, p := range par {
//			if p != "" {
//				parts = append(parts, p)
//				//如果p是以通配符*开头的
//				if p[0] == '*' {
//					break
//				}
//			}
//		}
//		return
//	}
//
// 在router.go文件中
func parsePath(path string) (parts []string) {
	par := strings.Split(path, "/")
	for _, p := range par {
		if p != "" {
			parts = append(parts, p)
			//如果p是以通配符*开头的
			if p[0] == '*' {
				break
			}
		}
	}
	return
}

// 添加路由
func (r *router) addRoute(method string, path string, handler HandlerFunc) {
	// key := method + "-" + pattern
	// r.handers[key] = handler

	if _, ok := r.root[method]; !ok {
		r.root[method] = &node{}
	}

	parts := parsePath(path)
	r.root[method].insert(path, parts)

	key := method + "-" + path
	r.handers[key] = handler
}

func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	root, ok := r.root[method]
	if !ok {
		return nil, nil
	}

	searchParts := parsePath(path)
	n := root.search(searchParts, 0)
	if n == nil {
		fmt.Println("nil,nil")
		return nil, nil
	}
	params := make(map[string]string)
	parts := parsePath(n.path)
	for i, part := range parts {
		if part[0] == ':' {
			params[part[1:]] = searchParts[i]
		}
		if part[0] == '*' && len(part) > 1 {
			params[part[1:]] = strings.Join(searchParts[i:], "/")
			break
		}
	}
	return n, params
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params
		// key := c.Method + "-" + c.Path
		key := c.Method + "-" + n.path
		fmt.Println(key)
		r.handers[key](c)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}
