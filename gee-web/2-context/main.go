package main

import (
	"geeV2/gee"
	"net/http"
)

func main() {
	engine := gee.New()
	engine.GET("/", func(c *gee.Context) {
		c.String(http.StatusOK, "ok..")
	})

	engine.POST("/hello", func(c *gee.Context) {
		c.JSON(http.StatusOK, gee.H{
			"name": "abc",
			"age":  32,
		})
	})

	engine.Run("localhost:10000")
}
