package main

import (
	"log"
	"os"
	"sulog"
)

func main() {
	//使用默认全局变量
	sulog.Info("std info log")
	sulog.SetOptions(sulog.WithLevel(sulog.ErrorLevel))
	sulog.Info("std can not info") //设置了ErrorLevel等级，那InfoLevle就输出不了

	sulog.SetOptions(sulog.WithFormatter(&sulog.JsonFormatter{}))
	sulog.Error("bad error", "ERRORLEVLE OK")
	sulog.Errorf("%s %d", "myhome", 111) //用户自定义message输出格式

	//输出到文件
	file, err := os.OpenFile("./test.log", os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("create file test.log failed")
	}
	defer file.Close()

	//自定义logger变量,New函数中设置选项
	l := sulog.New(sulog.WithLevel(sulog.InfoLevel),
		sulog.WithOutput(file))
	l.SetOptions(sulog.WithFormatter(&sulog.JsonFormatter{DisableTimestamp: true}))
	l.Info("log with json")
	l.Infof("%s %d", "lihi", 111)
}
