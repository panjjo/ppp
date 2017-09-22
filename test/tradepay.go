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
	tradePayResult := ppp.PayResult{}
	err = client.Call("AliPay.BarCodePay", ppp.BarCodePayRequest{
		OutTradeId: "20178702312435",
		TradeName:  "Craig",
		ItemDes:    "快速收银,临时商品",
		AuthCode:   "280121473379278075",
		Amount:     1,
		UserId:     "userid"}, &tradePayResult)
	fmt.Println(tradePayResult)
	if err != nil {
		fmt.Println(err)
	}
	user := ppp.User{Token: "1234"}
	userres := ppp.UserResult{}
	err = client.Call("Account.Add", user, &userres)
	fmt.Println(user)
	if err != nil {
		fmt.Println(err)
	}
}
