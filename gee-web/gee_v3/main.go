package main

import (
	"geeV3/gee"
	"net/http"
)

func main() {
	r := gee.New()
	r.GET("/", func(c *gee.Context) {
		c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
	})

	r.GET("/hello", func(c *gee.Context) {
		// expect /hello?name=geektutu
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
	})

	r.GET("/hello/:name", func(c *gee.Context) {
		// expect /hello/geektutu
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
	})

	r.GET("/assets/*filepath", func(c *gee.Context) {
		c.JSON(http.StatusOK, gee.H{"filepath": c.Param("filepath")})
	})

	r.GET("/:name", func(c *gee.Context) {
		name := c.Param("name")
		c.String(http.StatusOK, "name is %s", name)
	})

	r.GET("/16", func(c *gee.Context) {
		c.String(http.StatusOK, "16 is 16")
	})

	r.GET("/user/info/a", func(c *gee.Context) {
		c.String(http.StatusOK, "static is %s", "sdfsd")
	})

	r.GET("/user/:id/a", func(c *gee.Context) {
		name := c.Param("id")
		c.String(http.StatusOK, "id is %s", name)
	})

	r.Run("localhost:10000")
}
