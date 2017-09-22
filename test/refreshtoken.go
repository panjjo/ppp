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
	err = client.Call("AliPay.RefreshToken", ppp.RefreshToken{
		Type:   "code",
		Code:   "8abf42b8bb14445aa3386b4399b46X84",
		UserId: "userid"}, &tradePayResult)
	fmt.Println(tradePayResult)
	if err != nil {
		fmt.Println(err)
	}
}
