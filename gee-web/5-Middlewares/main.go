package main

import (
	"fmt"
	"geeV5/gee"
	"net/http"
)

func main() {
	r := gee.New()
	r.Use(gee.Logger())
	r.GET("/home", func(c *gee.Context) {
		c.HTML(http.StatusOK, "<h1>Index home</h1>")
	})

	v1 := r.Group("/admin")
	v1.Use(authMiddleWare())
	{
		v1.GET("/myname", func(c *gee.Context) {
			c.String(http.StatusOK, "Hello li")
		})
	}

	r.Run("localhost:10000")
}

func authMiddleWare() gee.HandlerFunc {
	return func(c *gee.Context) {
		fmt.Println("start 鉴权中间件")
		token := c.Req.Header.Get("token")
		if token == "" || token != "123" {
			c.JSON(http.StatusUnauthorized, gee.H{"message": "身份验证失败"})
			c.Abort()
			//return //不使用return的话，后面的fmt.Println("hello")会执行，使用return的话，后面的不会执行
		} else {
			fmt.Println("鉴权成功")
		}
		//使用Abort()后要是不使用return,打印hello还是会执行的
		//fmt.Println("hello")
	}
}
