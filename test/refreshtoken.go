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
	tradePayResult := ppp.UserResult{}
	err = client.Call("AliPay.Token", ppp.Token{
		Type: "code",
		Code: "b985e5b8c5474897b32ce7737741aX84",
	}, &tradePayResult)
	fmt.Println(tradePayResult)
	if err != nil {
		fmt.Println(err)
	}
}
