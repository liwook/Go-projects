package main

import (
	"geeV7/gee"
	"net/http"
)

func main() {
	r := gee.Default()
	r.GET("/", func(c *gee.Context) {
		c.String(http.StatusOK, "Hello Gee\n")
	})
	// index out of range for testing Recovery()
	r.GET("/panic", func(c *gee.Context) {
		names := []string{"gee"}
		c.String(http.StatusOK, names[100])
	})
	r.Run("localhost:10000")
}

//开启新协程
// func main() {
// 	r := gee.Default() //Default是使用了recovery中间件的
// 	// index out of range for testing Recovery()
// 	r.GET("/panic", func(c *gee.Context) {
// 		names := []string{"geek"}
// 		c.String(http.StatusOK, names[100])
// 	})

// 	//开启子线程
// 	go func() {
// 		arr := []int{1, 2}
// 		fmt.Println(arr[4])
// 	}()
// 	r.Run("localhost:10000")
// }

//在路由Handler协程中去开启新线程
// func main() {
// 	r := gee.Default() //Default是使用了recovery中间件的
// 	r.GET("/bad", func(c *gee.Context) {
// 		//在路由Handler协程中去开启新线程
// 		go func() {
// 			arr := []int{1, 2}
// 			fmt.Println(arr[4])
// 		}()
// 		c.String(http.StatusOK, "hello")
// 	})

// 	r.Run("localhost:10000")
// }
