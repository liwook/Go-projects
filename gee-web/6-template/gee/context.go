package gee

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
)

const AbortIndex = math.MaxInt8 >> 1

type H map[string]any

type Context struct {
	Wrtier http.ResponseWriter
	Req    *http.Request

	Path   string
	Method string
	Params map[string]string
	//响应的状态码
	StatusCode int

	// middleware
	midHandlers []HandlerFunc //这节新添加的
	index       int8          //这节新添加的

	engine *Engine
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Wrtier: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		index:  -1, //这节新添加的
	}
}

func (c *Context) Next() {
	c.index++

	for c.index < int8(len(c.midHandlers)) {
		c.midHandlers[c.index](c) //执行中间件或者路由Handler
		c.index++
	}
}

func (c *Context) Abort() {
	c.index = AbortIndex
}

func (c *Context) Param(key string) string {
	value := c.Params[key]
	return value
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Wrtier.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Wrtier.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...any) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Wrtier.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj any) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Wrtier)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Wrtier, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Wrtier.Write(data)
}

func (c *Context) Fail(code int, err string) {
	c.index = AbortIndex
	c.JSON(code, H{"message": err})
}

// func (c *Context) HTML(code int, html string) {
func (c *Context) HTML(code int, name string, data any) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Wrtier, name, data); err != nil {
		c.Fail(500, err.Error())
	}

	// c.SetHeader("Content-Type", "text/html")
	// c.Status(code)
	// c.Wrtier.Write([]byte(html))
}
