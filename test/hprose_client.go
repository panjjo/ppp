package main

import (
	"context"
	"fmt"
	ppp "github.com/panjjo/ppp/proto"
	"google.golang.org/grpc"
	"log"
)

func main() {
	client,err:=grpc.Dial("127.0.0.1:1233",grpc.WithInsecure())
	if err!=nil{
		log.Fatal(err)
	}
	pppClient:=ppp.NewPPPClient(client)
	trade,err:=pppClient.AliBarPay(context.Background(),&ppp.Barpay{
		UserID:"a56a767946",
		AuthCode:"1354564",
		ShopID:"111",
		Amount: 1,
		TradeName:"test",
		OutTradeID:"123adskfjoeifja",
	})
	fmt.Println(trade,err)
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
