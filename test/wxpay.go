package main

import (
	"fmt"
	"ppp"
)

var wxpay *ppp.WXPay

func main() {

	config := ppp.LoadConfig("/Users/panjjo/work/go/src/ppp/config.yml")
	ppp.NewLogger(config.Sys.LogLevel)
	ppp.NewDBPool(&config.DB)

	if !config.WXPay.Use {
		return
	}
	wxpay = ppp.NewWXPay(config.WXPay)
	// barPay()
	authSigner1()
	// tradeInfo()
	// cancel()
	// refund()
	// payParams(ppp.APPPAY)
	// payParams(ppp.WEBPAY)

}

func authSigner1() {
	fmt.Println(wxpay.AuthSigned(&ppp.Auth{MchID: "1490825832", Status: 1, Account: "闪收网络"}))
}
