package main

import (
	"fmt"

	"github.com/hprose/hprose-golang/rpc"
	"github.com/panjjo/ppp"
)

func main() {
	client := rpc.NewTCPClient("tcp://127.0.0.1:1234")
	service := &ppp.Services{}
	client.UseService(&service)
	// fmt.Println(service.AliPayAuth("d70f66b37f5e40259b56277b29edbX49"))
	fmt.Println(service.WXPayPayParams(&ppp.TradeParams{
		MchID: "1490825832",
	}))
}

// type Student struct {
// 	Name string
// 	b    bb
// }
// type bb struct {
// 	tag int
// }

// func (s *Student) copy() {
// 	var t *Student = new(Student)

// 	s.Name = "jack"
// 	s.b.tag = 1

// 	*t = *s

// 	fmt.Println("t=", t, "s=", s)

// 	s.Name = "rose"
// 	s.b.tag = 2

// 	fmt.Println("t=", t, "s=", s)
// 	t.b.tag = 3
// 	fmt.Println("t=", t, "s=", s)
// }

// func main() {

// 	s := &Student{}
// 	s.copy()

// 	// var s *Student = new(Student)

// }
