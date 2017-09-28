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
	refundResult := ppp.TradeResult{}
	err = client.Call("AliPay.Refund", ppp.RefundRequest{
		OutTradeId: "20178792413436",
		Amount:     100,
		Memo:       "test1234",
		UserId:     "ebbcb0f8c999c2b"}, &refundResult)
	fmt.Println(refundResult)
	if err != nil {
		fmt.Println(err)
	}
}
