package gee

import (
	"log"
	"time"
)

// 该函数返回的是函数指针
func Logger() HandlerFunc {
	return func(c *Context) {
		// Start timer
		t := time.Now()
		// Process request
		c.Next()
		// Calculate resolution time
		log.Printf("全局中间件  [%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
