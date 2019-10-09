package main

import (
	"fmt"
	"github.com/hprose/hprose-golang/rpc"
	"github.com/panjjo/ppp"
)

func main() {
	client := rpc.NewTCPClient("tcp://127.0.0.1:1233")
	service := &ppp.Services{}
	client.UseService(&service)
	fmt.Println(service.WXPayTradeInfo(&ppp.Trade{MchID:"1490825832",OutTradeID:"201910081734490"},true,"wxfa9c9911c1beae14"))
	// fmt.Println(service.WXPayBarPay(&ppp.BarPay{
	// 	OutTradeID:"201901311100450",
	// 	TradeName:"abc",
	// 	Amount:1,
	// 	ItemDes:"abc",
	// 	UserID:"a56a767946",
	// 	AuthCode:"1354564",
	// 	ShopID:"92e8dba4466b2a1daaa5b4c8134aea24",
	// 	IPAddr:"154.124.15.12",
	// },""))
	// fmt.Println(service.AliPayAuth("d70f66b37f5e40259b56277b29edbX49"))
	// fmt.Println(service.AliPayBindUser(&ppp.User{UserID:"201901080932523",MchID:"2088021225710398"},"2019010262749621"))
	// fmt.Println(service.AliPayTradeInfo(&ppp.Trade{OutTradeID: "123456", MchID: "2088021225710398"}, true, "2019010262749621"))
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
