package main

import (
	"fmt"

	"github.com/panjjo/ppp"
)

var alipay *ppp.AliPay
var ctx *ppp.Context

func main() {

	config := ppp.LoadConfig("/Users/panjjo/work/go/src/ppp/config.yml")
	ppp.NewLogger(config.Sys.LogLevel)
	ppp.NewDB(config.DB)

	if !config.AliPay.Use {
		return
	}
	alipay = ppp.NewAliPay(config.AliPay)
	ctx = ppp.NewContextWithCfg(ppp.ALIPAY, "2017020405513208")
	// auth()
	authSigner()
	// barPay()
	// tradeInfo()
	// cancel()
	// refund()
	// payParams(ppp.APPPAY)
	// payParams(ppp.WEBPAY)
}

func payParams(t ppp.TradeType) {
	// fmt.Println(alipay.PayParams(&ppp.TradeParams{
	// 	OutTradeID: "123456",
	// 	TradeName:  "测试",
	// 	Amount:     7520,
	// 	ItemDes:    "trade for test",
	// 	ShopID:     "abcd",
	// 	Type:       t,
	// }))
}

//
// func refund() {
// 	fmt.Println(alipay.Refund(&ppp.Refund{
// 		OutRefundID: "123456",
// 		SourceID:    "test12345",
// 		Amount:      101,
// 		Memo:        "可怜你的",
// 		MchID:       "2088102169330843",
// 	}))
// }
//
// func cancel() {
// 	fmt.Println(alipay.Cancel(&ppp.Trade{
// 		MchID:      "2088102169330843",
// 		OutTradeID: "test12345",
// 	}))
// }
//
// func tradeInfo() {
// 	fmt.Println(alipay.TradeInfo(&ppp.Trade{
// 		MchID:      "2088102169330843",
// 		OutTradeID: "test12345",
// 	}, true))
// }
//
// func barPay() {
// 	fmt.Println(alipay.BarPay(&ppp.BarPay{
// 		OutTradeID: "test12345",
// 		Amount:     7520,
// 		AuthCode:   "284087546000708768",
// 		MchID:      "2088102169330843",
// 		ItemDes:    "测试一下",
// 		TradeName:  "测试",
// 	}))
// }

func auth() {
	fmt.Println(alipay.Auth(ctx, "07e95dd9a3d647b1838b0880f33e3X49"))
}

func authSigner() {
	fmt.Println(alipay.AuthSigned(ctx, &ppp.Auth{MchID: "2088202800798491", Status: 1, Account: "admin@shengyun.org"}))
}
