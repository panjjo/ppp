package main

import (
	"fmt"
	"ppp"

	"github.com/hprose/hprose-golang/rpc"
)

func main() {
	client := rpc.NewTCPClient("tcp://127.0.0.1:1234")
	var alipay *ppp.AliPay
	client.UseService(&alipay)
	fmt.Println(alipay.Auth("123456"))
}
