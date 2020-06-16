package main

import (
	"flag"

	"github.com/panjjo/ppp"
	"github.com/sirupsen/logrus"
)

var alipay *ppp.AliPay
var wxpay *ppp.WXPay
var wxpaySingle *ppp.WXPaySingle

var configPath = flag.String("path", "./config.yml", "配置文件地址")
var scheme = flag.String("scheme", "rpc", "启动方式")

var config *ppp.Configs

func main() {
	flag.Parse()
	logrus.Infoln("config path", *configPath)
	config = ppp.LoadConfig(*configPath)
	ppp.NewLogger(config.Sys.LogLevel)
	ppp.NewDB(config.DB)
	// 初始化 alipay 和 默认收款账号
	if config.AliPay.Use {
		alipay = ppp.NewAliPay(config.AliPay)
		logrus.Infoln("alipay init succ")
	}
	if config.WXPay.Use {
		wxpay = ppp.NewWXPay(config.WXPay)
		logrus.Infoln("wxpay init succ")
	}
	if config.WXSingle.Use {
		wxpaySingle = ppp.NewWXPaySingle(config.WXSingle)
		logrus.Infoln("wxpaySingle init succ")
	}
	hproseRPC()
}
