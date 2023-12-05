package main

import (
	"fmt"
	"geeV6/gee"
	"net/http"
	"text/template"
	"time"
)

type student struct {
	Name string
	Age  int8
}

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
	//return fmt.Sprintf("%d-%02d-%d", year, month, day)
}

func main() {
	r := gee.New()
	r.Use(gee.Logger())
	//设置自定义渲染函数funcMap，custom_func.tmpl文件中的FormatAsDate格式就是FormatAsDate函数返回的格式
	r.SetFuncMap(template.FuncMap{
		"FormatAsDate": FormatAsDate,
	})
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./static")

	stu1 := &student{Name: "Geektutu", Age: 20}
	stu2 := &student{Name: "Jack", Age: 22}
	r.GET("/", func(c *gee.Context) {
		c.HTML(http.StatusOK, "css.tmpl", nil)
	})
	r.GET("/students", func(c *gee.Context) {
		c.HTML(http.StatusOK, "arr.tmpl", gee.H{
			"title":  "gee",
			"stuArr": [2]*student{stu1, stu2},
		})
	})

	r.GET("/date", func(c *gee.Context) {
		c.HTML(http.StatusOK, "custom_func.tmpl", gee.H{
			"title": "gee",
			"now":   time.Date(2023, 12, 5, 0, 0, 0, 0, time.UTC),
		})
	})

	r.Run("localhost:10000")
}
