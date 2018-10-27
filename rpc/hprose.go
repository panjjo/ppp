package main

import (
	"flag"
	"net"
	"path/filepath"

	"github.com/hprose/hprose-golang/rpc"
	"github.com/panjjo/log4go"
	"github.com/panjjo/ppp"
)

var alipay *ppp.AliPay
var wxpay *ppp.WXPay
var wxpaySingle *ppp.WXPaySingle
var wxpaySingleForAPP *ppp.WXPaySingleForAPP

var configPath = flag.String("path", "", "配置文件地址")
var scheme = flag.String("scheme", "rpc", "启动方式")

func main() {
	flag.Parse()

	config := ppp.LoadConfig(filepath.Join(*configPath, "./config.yml"))
	ppp.Log.DEBUG.Println("config path", filepath.Join(*configPath, "./config.yml"))
	ppp.NewLogger(config.Sys.LogLevel)
	ppp.NewDBPool(&config.DB)

	service := rpc.NewTCPService()

	if config.AliPay.Use {
		alipay = ppp.NewAliPay(config.AliPay)
		service.AddInstanceMethods(alipay, rpc.Options{NameSpace: ppp.ALIPAY})
		ppp.Log.DEBUG.Println("alipay init succ")
	}
	if config.WXPay.Use {
		wxpay = ppp.NewWXPay(config.WXPay)
		service.AddInstanceMethods(wxpay, rpc.Options{NameSpace: ppp.WXPAY})
		ppp.Log.DEBUG.Println("wxpay init succ")
	}
	if config.WXSingle.Other.Use {
		wxpaySingle = ppp.NewWXPaySingle(config.WXSingle.Other)
		service.AddInstanceMethods(wxpaySingle, rpc.Options{NameSpace: ppp.WXPAYSINGLE})
		ppp.Log.DEBUG.Println("wxpay_single init succ")
	}
	if config.WXSingle.APP.Use {
		wxpaySingleForAPP = ppp.NewWXPaySingleForAPP(config.WXSingle.APP)
		service.AddAllMethods(wxpaySingleForAPP, rpc.Options{NameSpace: ppp.WXPAYAPP})
		ppp.Log.DEBUG.Println("wxpay_app init succ")
	}

	l, e := net.Listen("tcp", config.Sys.ADDR)
	if e != nil {
		ppp.Log.ERROR.Panicf("listen tcp %s error:%v", config.Sys.ADDR, e)
	}
	ppp.Log.INFO.Println("listen tcp at", config.Sys.ADDR)
	service.Serve(l)

}

type logFilter struct {
	log *log4go.Logger
}

func (lf logFilter) handler(
	request []byte,
	context rpc.Context,
	next rpc.NextFilterHandler) (response []byte, err error) {
	lf.log.INFO.Printf("request:%s", request)
	response, err = next(request, context)
	lf.log.INFO.Printf("response:%s", response)
	return
}
