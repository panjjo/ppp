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
	tradePayResult := ppp.AccountResult{}
	err = client.Call("Account.Regist", ppp.User{
		Type:   ppp.PAYTYPE_ALIPAY,
		UserId: "1234",
	}, &tradePayResult)
	fmt.Println(tradePayResult)
	if err != nil {
		fmt.Println(err)
	}

	//token
	tokenres := ppp.AuthResult{}
	err = client.Call("AliPay.Auth", ppp.Token{
		Code: "821d8e3eee4640669d55102ae457aF84",
	}, &tokenres)
	fmt.Println(tokenres)
	if err != nil {
		fmt.Println(err)
	}

	//授权
	resp := &ppp.Response{}
	err = client.Call("Account.Auth", ppp.AccountAuth{Type: ppp.PAYTYPE_ALIPAY, UserId: "1234", MchId: tokenres.Data.MchId}, resp)
	fmt.Println(resp)
	if err != nil {
		fmt.Println(err)
	}
}
