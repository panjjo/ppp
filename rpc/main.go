package main

import (
	"flag"
	"path/filepath"

	"github.com/panjjo/ppp"
)

var alipay *ppp.AliPay
var wxpay *ppp.WXPay
var wxpaySingle *ppp.WXPaySingle

var configPath = flag.String("path", "", "配置文件地址")
var scheme = flag.String("scheme", "rpc", "启动方式")

var config *ppp.Configs

func main() {
	flag.Parse()

	config = ppp.LoadConfig(filepath.Join(*configPath, "./config.yml"))
	ppp.Log.DEBUG.Println("config path", filepath.Join(*configPath, "./config.yml"))
	ppp.NewLogger(config.Sys.LogLevel)
	ppp.NewDBPool(&config.DB)
	// 初始化 alipay 和 默认收款账号
	if config.AliPay.Use {
		alipay = ppp.NewAliPay(config.AliPay)
		ppp.Log.DEBUG.Println("alipay init succ")
	}
	if config.WXPay.Use{
		wxpay = ppp.NewWXPay(config.WXPay)
		ppp.Log.DEBUG.Println("wxpay init succ")
	}
	if config.WXSingle.Use{
		wxpaySingle = ppp.NewWXPaySingle(config.WXSingle)
		ppp.Log.DEBUG.Println("wxpaySingle init succ")
	}
	hproseRPC()
}
