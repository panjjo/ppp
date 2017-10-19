package main

import (
	"fmt"
	"log"
	"net/rpc"

	"github.com/panjjo/ppp"
)

func main() {
	client, err := rpc.DialHTTP("tcp", ":1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	result := ppp.AuthResult{}
	err = client.Call("AliPay.Auth", ppp.Token{
		Code: "dd4a1949fde143a3af17baa13b8adX84",
	}, &result)
	fmt.Println(result)
	if err != nil {
		fmt.Println(err)
	}
}
