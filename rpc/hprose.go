package main

import (
	"flag"
	"net"
	"path/filepath"
	"ppp"

	"github.com/hprose/hprose-golang/rpc"
)

var alipay *ppp.AliPay
var wxpay *ppp.WXPay
var wxpaySingle *ppp.WXPaySingle
var wxpaySingleForAPP *ppp.WXPaySingleForAPP

var configPath = flag.String("path", "", "配置文件地址")

func main() {
	flag.Parse()

	config := ppp.LoadConfig(filepath.Join(*configPath, "./config.yml"))
	ppp.Log.DEBUG.Println("config path", filepath.Join(*configPath, "./config.yml"))
	ppp.NewLogger(config.Sys.LogLevel)
	ppp.NewDBPool(&config.DB)

	service := rpc.NewTCPService()

	if config.AliPay.Use {
		alipay = ppp.NewAliPay(config.AliPay)
		service.AddInstanceMethods(alipay)
		ppp.Log.DEBUG.Println("alipay init succ")
	}
	if config.WXPay.Use {
		wxpay = ppp.NewWXPay(config.WXPay)
		service.AddInstanceMethods(wxpay)
		ppp.Log.DEBUG.Println("wxpay init succ")
	}
	if config.WXSingle.Other.Use {
		wxpaySingle = ppp.NewWXPaySingle(config.WXSingle.Other)
		service.AddInstanceMethods(wxpaySingle)
		ppp.Log.DEBUG.Println("wxpay_single init succ")
	}
	if config.WXSingle.APP.Use {
		wxpaySingleForAPP = ppp.NewWXPaySingleForAPP(config.WXSingle.APP)
		service.AddAllMethods(wxpaySingleForAPP)
		ppp.Log.DEBUG.Println("wxpay_app init succ")
	}
	l, e := net.Listen("tcp", config.Sys.ADDR)
	if e != nil {
		ppp.Log.ERROR.Panicf("listen tcp %s error:%v", config.Sys.ADDR, e)
	}
	service.Serve(l)
}
