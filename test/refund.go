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
	//alipay()
	wxpay()
}
func wxpay() {
	refundResult := ppp.TradeResult{}
	err = client.Call(ppp.FC_WXPAY_REFUND, ppp.RefundRequest{
		OutTradeId:  "20774752413168",
		OutRefundId: "1232asdf2398af23",
		Amount:      1,
		Memo:        "test1234",
		UserId:      "120acd7a9d12"}, &refundResult)
	fmt.Println(refundResult)
	if err != nil {
		fmt.Println(err)
	}
}

func alipay() {
	refundResult := ppp.TradeResult{}
	err = client.Call(ppp.FC_ALIPAY_REFUND, ppp.RefundRequest{
		OutTradeId: "20178792413436",
		Amount:     100,
		Memo:       "test1234",
		UserId:     "ebbcb0f8c999c2b"}, &refundResult)
	fmt.Println(refundResult)
	if err != nil {
		fmt.Println(err)
	}
}
