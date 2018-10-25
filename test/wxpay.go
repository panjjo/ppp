package main

import (
	"fmt"
	"ppp2"
)

var wxpay *ppp2.WXPay

func main() {

	config := ppp2.LoadConfig("/Users/panjjo/work/go/src/ppp2/config.yml")
	ppp2.NewLogger(config.Sys.LogLevel)
	ppp2.NewDBPool(&config.DB)

	if !config.WXPay.Use {
		return
	}
	wxpay = ppp2.NewWXPay(config.WXPay)
	// barPay()
	authSigner1()
	// tradeInfo()
	// cancel()
	// refund()
	// payParams(ppp2.APPPAY)
	// payParams(ppp2.WEBPAY)

}

func authSigner1() {
	fmt.Println(wxpay.AuthSigned(&ppp2.Auth{MchID: "1490825832", Status: 1, Account: "闪收网络"}))
}
