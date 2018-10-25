package main

import (
	"fmt"
	"ppp2"
)

var alipay *ppp2.AliPay

func main1() {

	config := ppp2.LoadConfig("/Users/panjjo/work/go/src/ppp2/config.yml")
	ppp2.NewLogger(config.Sys.LogLevel)
	ppp2.NewDBPool(&config.DB)

	if !config.AliPay.Use {
		return
	}
	alipay = ppp2.NewAliPay(config.AliPay)
	// auth()
	// authSigner()
	// barPay()
	// tradeInfo()
	// cancel()
	// refund()
	// payParams(ppp2.APPPAY)
	// payParams(ppp2.WEBPAY)
}

func payParams(t ppp2.TradeType) {
	fmt.Println(alipay.PayParams(&ppp2.TradeParams{
		OutTradeID: "123456",
		TradeName:  "测试",
		Amount:     7520,
		ItemDes:    "trade for test",
		ShopID:     "abcd",
		Type:       t,
	}))
}

func refund() {
	fmt.Println(alipay.Refund(&ppp2.Refund{
		OutRefundID: "123456",
		SourceID:    "test12345",
		Amount:      101,
		Memo:        "可怜你的",
		MchID:       "2088102169330843",
	}))
}

func cancel() {
	fmt.Println(alipay.Cancel(&ppp2.Trade{
		MchID:      "2088102169330843",
		OutTradeID: "test12345",
	}))
}

func tradeInfo() {
	fmt.Println(alipay.TradeInfo(&ppp2.Trade{
		MchID:      "2088102169330843",
		OutTradeID: "test12345",
	}, true))
}

func barPay() {
	fmt.Println(alipay.BarPay(&ppp2.BarPay{
		OutTradeID: "test12345",
		Amount:     7520,
		AuthCode:   "284087546000708768",
		MchID:      "2088102169330843",
		ItemDes:    "测试一下",
		TradeName:  "测试",
	}))
}

func auth() {
	fmt.Println(alipay.Auth("8efe0e815f6c464d96fdb47ef1c37X84"))
}

func authSigner() {
	fmt.Println(alipay.AuthSigned(&ppp2.Auth{MchID: "2088102169330843", Status: 1, Account: "admin@shengyun.org"}))
}
