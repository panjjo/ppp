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
	fmt.Println(service.AliPayBindUser(&ppp.User{UserID:"201901080932523",MchID:"2088021225710398"}))
	// fmt.Println(service.WXPaySingleBarPay(&ppp.BarPay{
	// 	MchID:      "1490825832",
	// 	Amount:     1,
	// 	AuthCode:   "",
	// 	OutTradeID: "64174267ew23a82",
	// 	TradeName:  "testa",
	// 	ItemDes:    "124354",
	// 	ShopID:     "1234",
	// }))
	// fmt.Println(service.WXPaySingleTradeInfo(&ppp.Trade{
	// 	OutTradeID:"18112411094200007",
	// },true))
}

