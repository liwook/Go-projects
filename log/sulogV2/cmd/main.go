package main

import (
	"fmt"
	"sulog"
)

type hook struct{}

func (h *hook) Levels() []sulog.Level {
	// return sulog.AllLevels
	return []sulog.Level{sulog.InfoLevel, sulog.DebugLevel}
}

func (h *hook) Fire(entry *sulog.Entry) error {
	fmt.Printf("this is a hook func:%v\n", entry.Data)
	return nil
}

func main() {
	// l := logrus.WithField("name", "21")
	// l.Info("hello")
	// l.Error("ok")
	// return
	//使用默认全局变量
	// fmt.Println(logrus.GetLevel())
	// return
	sulog.SetOptions(sulog.WithDisableCaller(true))
	sulog.SetOptions(sulog.WithFormatter(&sulog.TextFormatter{DisableTimestamp: true}))
	// sulog.AddHook(&hook{})
	l := sulog.WithField("name", "21")
	l.Info("hello")
	l.Error("ok")
	return
	sulog.WithField("name", "11").Info("ok withField")
	fmt.Println()
	sulog.WithField("country", "China").Error("ok withField")
}

// type hook struct{}

// func (h *hook) Levels() []sulog.Level {
// 	return sulog.AllLevels
// }
// func (h *hook) Fire(entry *sulog.Entry) error {
// 	fmt.Println("this is a hook func:", entry.Data)
// 	return nil
// }

// func main() {
// 	// slog.Info("sdf")

// 	// slog.With("source", "100").Info("hello info")
// 	// logrus.SetReportCaller(true)
// 	// // logrus.AddHook
// 	// logrus.WithField("level", "li").WithField("addr", 11).Info("ok info")
// 	// logrus.WithTime(time.Now()).Info("hello")

// 	// logrus.SetFormatter(&logrus.TextFormatter{})
// 	// //使用默认全局变量
// 	// sulog.Info("std info log")
// 	// sulog.SetOptions(sulog.WithLevel(sulog.ErrorLevel))
// 	// sulog.Info("std can not info") //设置了ErrorLevel等级，那InfoLevle就输出不了

// 	// sulog.SetOptions(sulog.WithFormatter(&sulog.TextFormatter{}))
// 	// sulog.Error("bad error", "ERRORLEVLE OK")
// 	// sulog.Errorf("%s %d", "myhome", 111) //用户自定义message输出格式

// 	// return
// 	// sulog.Info("hello info")
// 	// return

// 	sulog.AddHook(&hook{})
// 	sulog.WithField("name", "11").Info("ok withField")
// 	// sulog.WithFields(sulog.Fields{
// 	// 	"name": "li",
// 	// 	"age":  32,
// 	// }).Info("ok withFields")
// 	// sulog.WithField("level", "debug").Info("我们ok withField")
// 	// return
// 	// sulog.SetOptions(sulog.WithFormatter(&sulog.JsonFormatter{}))
// 	sulog.SetOptions(sulog.WithDisableCaller(true))
// 	sulog.SetOptions(sulog.WithFormatter(&sulog.JsonFormatter{DisableTimestamp: true}))
// 	sulog.WithFields(sulog.Fields{
// 		"level": "info",
// 		"name":  "lihai",
// 		"msg":   "this field message",
// 	}).Info("ok withField")
// 	return
// 	//输出到文件
// 	file, err := os.OpenFile("./test.log", os.O_CREATE|os.O_APPEND, 0666)
// 	if err != nil {
// 		log.Fatal("create file test.log failed")
// 	}
// 	defer file.Close()

// 	//自定义logger变量,New函数中设置选项
// 	l := sulog.New(sulog.WithLevel(sulog.InfoLevel),
// 		sulog.WithOutput(file))
// 	l.SetOptions(sulog.WithFormatter(&sulog.JsonFormatter{DisableTimestamp: false}))
// 	l.Info("log with json")
// 	fmt.Println("--------------------------------------")
// 	l.WithField("name", "li").Debug("hello")
// 	l.WithField("name", "li").WithField("age", "23").Info("hello")
// 	l.WithFields(sulog.Fields{"name": "abcv",
// 		"addr": "zhongguo"}).Info("ok go home")

// 	var wg sync.WaitGroup
// 	wg.Add(1)
// 	go func() {
// 		for i := 0; i < 10; i++ {
// 			l.WithField("test_num", strconv.Itoa(i+20000)).Info("go func Hello")
// 		}
// 		wg.Done()
// 	}()

// 	for i := 0; i < 10; i++ {
// 		l.WithField("num", strconv.Itoa(i)).Info("Hello")
// 	}
// 	wg.Wait()
// }
