package main

import (
	"fmt"

	"github.com/hprose/hprose-golang/rpc"
	"github.com/panjjo/ppp"
)

func main() {
	client := rpc.NewHTTPClient("http://127.0.0.1:1234")
	service := &ppp.Services{}
	client.UseService(&service)
	// fmt.Println(service.AliPayAuth("d70f66b37f5e40259b56277b29edbX49"))
	fmt.Println(service.AliPayAuthSigned(&ppp.Auth{
		MchID:   "2088202800798491",
		Account: "panjjo",
	}))
}
