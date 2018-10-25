package main

import (
	"net"
	"ppp2"

	"github.com/hprose/hprose-golang/rpc"
)

var alipay *ppp2.AliPay
var wxpay *ppp2.WXPay
var wxpaySingle *ppp2.WXPaySingle
var wxpaySingleForAPP *ppp2.WXPaySingleForAPP

func main() {

	config := ppp2.LoadConfig("/Users/panjjo/work/go/src/ppp2/config.yml")
	ppp2.NewLogger(config.Sys.LogLevel)
	ppp2.NewDBPool(&config.DB)

	service := rpc.NewTCPService()

	if config.AliPay.Use {
		alipay = ppp2.NewAliPay(config.AliPay)
		service.AddInstanceMethods(alipay)
		ppp2.Log.DEBUG.Println("alipay init succ")
	}
	if config.WXPay.Use {
		wxpay = ppp2.NewWXPay(config.WXPay)
		service.AddInstanceMethods(wxpay)
		ppp2.Log.DEBUG.Println("wxpay init succ")
	}
	if config.WXSingle.Other.Use {
		wxpaySingle = ppp2.NewWXPaySingle(config.WXSingle.Other)
		service.AddInstanceMethods(wxpaySingle)
		ppp2.Log.DEBUG.Println("wxpay_single init succ")
	}
	if config.WXSingle.APP.Use {
		wxpaySingleForAPP = ppp2.NewWXPaySingleForAPP(config.WXSingle.APP)
		service.AddAllMethods(wxpaySingleForAPP)
		ppp2.Log.DEBUG.Println("wxpay_app init succ")
	}
	l, e := net.Listen("tcp", config.Sys.ADDR)
	if e != nil {
		ppp2.Log.ERROR.Panicf("listen tcp %s error:%v", config.Sys.ADDR, e)
	}
	service.Serve(l)
}
