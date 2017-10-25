package main

import (
	"fmt"
	"log"
	"net/rpc"

	"github.com/panjjo/ppp"
)

var client *rpc.Client
var err error

func main() {
	client, err = rpc.DialHTTP("tcp", ":1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	/*alipay()*/
	wxpay()
	/*wxsandbox()*/
}
func wxsandbox() {
	tradePayResult := ppp.Response{}
	err = client.Call(ppp.FC_WXPAY_SBKEY,
		"1489045402", &tradePayResult)
	if err != nil {
		fmt.Println("ERROR:", err)
	}
	fmt.Println(tradePayResult)
}
func wxpay() {
	tradePayResult := ppp.TradeResult{}
	err = client.Call(ppp.FC_WXPAY_BARCODEPAY, ppp.BarCodePayRequest{
		OutTradeId: "20774752413168",
		TradeName:  "Craig",
		ItemDes:    "快速收银,临时商品",
		AuthCode:   "135215036461711497",
		Amount:     1,
		UserId:     "120acd7a9d12"}, &tradePayResult)
	if err != nil {
		fmt.Println("ERROR:", err)
	}
	fmt.Println(tradePayResult)
}
func alipay() {
	tradePayResult := ppp.TradeResult{}
	err = client.Call(ppp.FC_ALIPAY_BARCODEPAY, ppp.BarCodePayRequest{
		OutTradeId: "20774752413167",
		TradeName:  "Craig",
		ItemDes:    "快速收银,临时商品",
		AuthCode:   "282878191055587849",
		Amount:     1000300,
		UserId:     "1234"}, &tradePayResult)
	if err != nil {
		fmt.Println("ERROR:", err)
	}
	fmt.Println(tradePayResult)
}
