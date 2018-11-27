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
	if config.WXSingle.Other.Use {
		service.AddInstanceMethods(&WXPaySingleHproseRPC{}, rpc.Options{NameSpace: ppp.WXPAYSINGLE})
	}
	if config.WXSingle.APP.Use {
		service.AddAllMethods(&WXPayAPPHproseRPC{}, rpc.Options{NameSpace: ppp.WXPAYAPP})
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
func (A *AliPayHproseRPC) PayParams(req *ppp.TradeParams) (data *ppp.PayParams, e ppp.Error) {
	tmp := *alipay
	return tmp.PayParams(req)
}

// TradeInfo ...
func (A *AliPayHproseRPC) TradeInfo(req *ppp.Trade, sync bool) (trade *ppp.Trade, e ppp.Error) {
	tmp := *alipay
	return tmp.TradeInfo(req, sync)
}

// BarPay ...
func (A *AliPayHproseRPC) BarPay(req *ppp.BarPay) (trade *ppp.Trade, e ppp.Error) {
	tmp := *alipay
	return tmp.BarPay(req)
}

// Refund ...
func (A *AliPayHproseRPC) Refund(req *ppp.Refund) (refund *ppp.Refund, e ppp.Error) {
	tmp := *alipay
	return tmp.Refund(req)
}

// Cancel ...
func (A *AliPayHproseRPC) Cancel(req *ppp.Trade) (e ppp.Error) {
	tmp := *alipay
	return tmp.Cancel(req)
}

// AuthSigned ...
func (A *AliPayHproseRPC) AuthSigned(req *ppp.Auth) (auth *ppp.Auth, e ppp.Error) {
	tmp := *alipay
	return tmp.AuthSigned(req)
}

// BindUser ...
func (A *AliPayHproseRPC) BindUser(req *ppp.User) (user *ppp.User, e ppp.Error) {
	tmp := *alipay
	return tmp.BindUser(req)
}

// UnBindUser ...
func (A *AliPayHproseRPC) UnBindUser(req *ppp.User) (user *ppp.User, e ppp.Error) {
	tmp := *alipay
	return tmp.UnBindUser(req)
}

// Auth ...
func (A *AliPayHproseRPC) Auth(req string) (auth *ppp.Auth, e ppp.Error) {
	tmp := *alipay
	return tmp.Auth(req)
}

// WXPayHproseRPC ...
type WXPayHproseRPC struct {
}

// PayParams ...
func (A *WXPayHproseRPC) PayParams(req *ppp.TradeParams) (data *ppp.PayParams, e ppp.Error) {
	tmp := *wxpay
	return tmp.PayParams(req)
}

// TradeInfo ...
func (A *WXPayHproseRPC) TradeInfo(req *ppp.Trade, sync bool) (trade *ppp.Trade, e ppp.Error) {
	tmp := *wxpay
	return tmp.TradeInfo(req, sync)
}

// BarPay ...
func (A *WXPayHproseRPC) BarPay(req *ppp.BarPay) (trade *ppp.Trade, e ppp.Error) {
	tmp := *wxpay
	return tmp.BarPay(req)
}

// Refund ...
func (A *WXPayHproseRPC) Refund(req *ppp.Refund) (refund *ppp.Refund, e ppp.Error) {
	tmp := *wxpay
	return tmp.Refund(req)
}

// Cancel ...
func (A *WXPayHproseRPC) Cancel(req *ppp.Trade) (e ppp.Error) {
	tmp := *wxpay
	return tmp.Cancel(req)
}

// AuthSigned ...
func (A *WXPayHproseRPC) AuthSigned(req *ppp.Auth) (auth *ppp.Auth, e ppp.Error) {
	tmp := *wxpay
	return tmp.AuthSigned(req)
}

// BindUser ...
func (A *WXPayHproseRPC) BindUser(req *ppp.User) (user *ppp.User, e ppp.Error) {
	tmp := *wxpay
	return tmp.BindUser(req)
}

// UnBindUser ...
func (A *WXPayHproseRPC) UnBindUser(req *ppp.User) (user *ppp.User, e ppp.Error) {
	tmp := *wxpay
	return tmp.UnBindUser(req)
}

// WXPaySingleHproseRPC ...
type WXPaySingleHproseRPC struct {
}

// PayParams ...
func (A *WXPaySingleHproseRPC) PayParams(req *ppp.TradeParams) (data *ppp.PayParams, e ppp.Error) {
	tmp := *wxpaySingle
	return tmp.PayParams(req)
}

// TradeInfo ...
func (A *WXPaySingleHproseRPC) TradeInfo(req *ppp.Trade, sync bool) (trade *ppp.Trade, e ppp.Error) {
	tmp := *wxpaySingle
	return tmp.TradeInfo(req, sync)
}

// BarPay ...
func (A *WXPaySingleHproseRPC) BarPay(req *ppp.BarPay) (trade *ppp.Trade, e ppp.Error) {
	tmp := *wxpaySingle
	return tmp.BarPay(req)
}

// Refund ...
func (A *WXPaySingleHproseRPC) Refund(req *ppp.Refund) (refund *ppp.Refund, e ppp.Error) {
	tmp := *wxpaySingle
	return tmp.Refund(req)
}

// Cancel ...
func (A *WXPaySingleHproseRPC) Cancel(req *ppp.Trade) (e ppp.Error) {
	tmp := *wxpaySingle
	return tmp.Cancel(req)
}

// WXPayAPPHproseRPC ...
type WXPayAPPHproseRPC struct {
}

// PayParams ...
func (A *WXPayAPPHproseRPC) PayParams(req *ppp.TradeParams) (data *ppp.PayParams, e ppp.Error) {
	tmp := *wxpaySingleForAPP
	return tmp.PayParams(req)
}

// TradeInfo ...
func (A *WXPayAPPHproseRPC) TradeInfo(req *ppp.Trade, sync bool) (trade *ppp.Trade, e ppp.Error) {
	tmp := *wxpaySingleForAPP
	return tmp.TradeInfo(req, sync)
}

// BarPay ...
func (A *WXPayAPPHproseRPC) BarPay(req *ppp.BarPay) (trade *ppp.Trade, e ppp.Error) {
	tmp := *wxpaySingleForAPP
	return tmp.BarPay(req)
}

// Refund ...
func (A *WXPayAPPHproseRPC) Refund(req *ppp.Refund) (refund *ppp.Refund, e ppp.Error) {
	tmp := *wxpaySingleForAPP
	return tmp.Refund(req)
}

// Cancel ...
func (A *WXPayAPPHproseRPC) Cancel(req *ppp.Trade) (e ppp.Error) {
	tmp := *wxpaySingleForAPP
	return tmp.Cancel(req)
}

// WXPayMINIPHproseRPC ...
type WXPayMINIPHproseRPC struct {
}

// PayParams ...
func (A *WXPayMINIPHproseRPC) PayParams(req *ppp.TradeParams) (data *ppp.PayParams, e ppp.Error) {
	tmp := *wxpaySingleForMINIP
	return tmp.PayParams(req)
}
