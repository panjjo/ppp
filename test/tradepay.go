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
	tradePayResult := ppp.TradeResult{}
	err = client.Call(ppp.FC_ALIPAY_BARCODEPAY, ppp.BarCodePayRequest{
		OutTradeId: "20178792413436",
		TradeName:  "Craig",
		ItemDes:    "快速收银,临时商品",
		AuthCode:   "283133232310613353",
		Amount:     100,
		UserId:     "ebbcb0f8c999c2b"}, &tradePayResult)
	if err != nil {
		fmt.Println("ERROR:", err)
	}
	fmt.Println(tradePayResult)
}
