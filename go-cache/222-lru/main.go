package main

import (
	"container/list"
	"fmt"
)

func main() {
	ll := list.New()
	ll.PushFront(1)
	ll.PushFront("home")

	fmt.Println(ll.Front().Value.(string))
	fmt.Println(ll.Back().Value.(int))
	//下面的写法也成功打印出来
	// fmt.Println(ll.Front().Value)
	// fmt.Println(ll.Back().Value)
}
