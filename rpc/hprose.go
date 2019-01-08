package main

import (
	"net"

	"github.com/hprose/hprose-golang/rpc"
	"github.com/panjjo/log4go"
	"github.com/panjjo/ppp"
)

func hproseRPC() {

	service := rpc.NewTCPService()
	if config.AliPay.Use {
		service.AddInstanceMethods(&AliPayHproseRPC{}, rpc.Options{NameSpace: ppp.ALIPAY})
	}
	if config.WXPay.Use {
		service.AddInstanceMethods(&WXPayHproseRPC{}, rpc.Options{NameSpace: ppp.WXPAY})
	}
	if config.WXSingle.Use {
		service.AddInstanceMethods(&WXPaySingleHproseRPC{}, rpc.Options{NameSpace: ppp.WXPAYSINGLE})
	}
	service.AddBeforeFilterHandler(logFilter{ppp.Log}.handler)
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
	lf.log.INFO.Println("request:", string(request))
	response, err = next(request, context)
	lf.log.INFO.Println("response:", string(response))
	return
}

// AliPayHproseRPC ...
type AliPayHproseRPC struct {
}

// PayParams ...
func (A *AliPayHproseRPC) PayParams(req *ppp.TradeParams, tag string) (data *ppp.PayParams, e ppp.Error) {
	return alipay.PayParams(ppp.NewContextWithCfg(ppp.ALIPAY, tag), req)
}

// TradeInfo ...
func (A *AliPayHproseRPC) TradeInfo(req *ppp.Trade, sync bool, tag string) (trade *ppp.Trade, e ppp.Error) {
	return alipay.TradeInfo(ppp.NewContextWithCfg(ppp.ALIPAY, tag), req, sync)
}

// BarPay ...
func (A *AliPayHproseRPC) BarPay(req *ppp.BarPay, tag string) (trade *ppp.Trade, e ppp.Error) {
	return alipay.BarPay(ppp.NewContextWithCfg(ppp.ALIPAY, tag), req)
}

// Refund ...
func (A *AliPayHproseRPC) Refund(req *ppp.Refund, tag string) (refund *ppp.Refund, e ppp.Error) {
	return alipay.Refund(ppp.NewContextWithCfg(ppp.ALIPAY, tag), req)
}

// Cancel ...
func (A *AliPayHproseRPC) Cancel(req *ppp.Trade, tag string) (e ppp.Error) {
	return alipay.Cancel(ppp.NewContextWithCfg(ppp.ALIPAY, tag), req)
}

// AuthSigned ...
func (A *AliPayHproseRPC) AuthSigned(req *ppp.Auth, tag string) (auth *ppp.Auth, e ppp.Error) {
	return alipay.AuthSigned(ppp.NewContextWithCfg(ppp.ALIPAY, tag), req)
}

// BindUser ...
func (A *AliPayHproseRPC) BindUser(req *ppp.User, tag string) (user *ppp.User, e ppp.Error) {
	return alipay.BindUser(ppp.NewContextWithCfg(ppp.ALIPAY, tag), req)
}

// UnBindUser ...
func (A *AliPayHproseRPC) UnBindUser(req *ppp.User, tag string) (user *ppp.User, e ppp.Error) {
	return alipay.UnBindUser(ppp.NewContextWithCfg(ppp.ALIPAY, tag), req)
}

// Auth ...
func (A *AliPayHproseRPC) Auth(req string, tag string) (auth *ppp.Auth, e ppp.Error) {
	return alipay.Auth(ppp.NewContextWithCfg(ppp.ALIPAY, tag), req)
}


// WXPayHproseRPC ...
type WXPayHproseRPC struct {
}

// PayParams ...
func (A *WXPayHproseRPC) PayParams(req *ppp.TradeParams,tag string) (data *ppp.PayParams, e ppp.Error) {
	return wxpay.PayParams(ppp.NewContextWithCfg(ppp.WXPAY, tag),req)
}

// TradeInfo ...
func (A *WXPayHproseRPC) TradeInfo(req *ppp.Trade, sync bool,tag string) (trade *ppp.Trade, e ppp.Error) {
	return wxpay.TradeInfo(ppp.NewContextWithCfg(ppp.WXPAY, tag),req, sync)
}

// BarPay ...
func (A *WXPayHproseRPC) BarPay(req *ppp.BarPay,tag string) (trade *ppp.Trade, e ppp.Error) {
	return wxpay.BarPay(ppp.NewContextWithCfg(ppp.WXPAY, tag),req)
}

// Refund ...
func (A *WXPayHproseRPC) Refund(req *ppp.Refund,tag string) (refund *ppp.Refund, e ppp.Error) {
	return wxpay.Refund(ppp.NewContextWithCfg(ppp.WXPAY, tag),req)
}

// Cancel ...
func (A *WXPayHproseRPC) Cancel(req *ppp.Trade,tag string) (e ppp.Error) {
	return wxpay.Cancel(ppp.NewContextWithCfg(ppp.WXPAY, tag),req)
}

// AuthSigned ...
func (A *WXPayHproseRPC) AuthSigned(req *ppp.Auth,tag string) (auth *ppp.Auth, e ppp.Error) {
	return wxpay.AuthSigned(ppp.NewContextWithCfg(ppp.WXPAY, tag),req)
}

// BindUser ...
func (A *WXPayHproseRPC) BindUser(req *ppp.User,tag string) (user *ppp.User, e ppp.Error) {
	return wxpay.BindUser(ppp.NewContextWithCfg(ppp.WXPAY, tag),req)
}

// UnBindUser ...
func (A *WXPayHproseRPC) UnBindUser(req *ppp.User,tag string) (user *ppp.User, e ppp.Error) {
	return wxpay.UnBindUser(ppp.NewContextWithCfg(ppp.WXPAY, tag),req)
}

// WXPaySingleHproseRPC ...
type WXPaySingleHproseRPC struct {
}

// PayParams ...
func (A *WXPaySingleHproseRPC) PayParams(req *ppp.TradeParams,tag string) (data *ppp.PayParams, e ppp.Error) {
	return wxpaySingle.PayParams(ppp.NewContextWithCfg(ppp.WXPAYSINGLE, tag),req)
}

// TradeInfo ...
func (A *WXPaySingleHproseRPC) TradeInfo(req *ppp.Trade, sync bool,tag string) (trade *ppp.Trade, e ppp.Error) {
	return wxpaySingle.TradeInfo(ppp.NewContextWithCfg(ppp.WXPAYSINGLE, tag),req, sync)
}

// BarPay ...
func (A *WXPaySingleHproseRPC) BarPay(req *ppp.BarPay,tag string) (trade *ppp.Trade, e ppp.Error) {
	return wxpaySingle.BarPay(ppp.NewContextWithCfg(ppp.WXPAYSINGLE, tag),req)
}

// Refund ...
func (A *WXPaySingleHproseRPC) Refund(req *ppp.Refund,tag string) (refund *ppp.Refund, e ppp.Error) {
	return wxpaySingle.Refund(ppp.NewContextWithCfg(ppp.WXPAYSINGLE, tag),req)
}

// Cancel ...
func (A *WXPaySingleHproseRPC) Cancel(req *ppp.Trade,tag string) (e ppp.Error) {
	return wxpaySingle.Cancel(ppp.NewContextWithCfg(ppp.WXPAYSINGLE, tag),req)
}

// MchPay ...
func (A *WXPaySingleHproseRPC) MchPay(req *ppp.MchPay,tag string) (tid string, e ppp.Error) {
	return wxpaySingle.MchPay(ppp.NewContextWithCfg(ppp.WXPAYSINGLE, tag),req)
}
