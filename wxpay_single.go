package ppp

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
)

const (
	// WXPAYSINGLE 微信支付单商户的标识
	WXPAYSINGLE string = "wxpay_single"
)

var wxpaySingle *WXPaySingle

// WXPaySingle 微信支付单商户模式主体
// 微信支付服务商模式
// 服务商模式与单商户模式区别只是多了一个 子商户权限，其余接口结构返回完全一致
type WXPaySingle struct {
	cfgs map[string]config
	def  config
}

func wxConfig(config ConfigSingle) (wx config) {
	if config.AppID != "" {
		wx.appid = config.AppID
	} else {
		if len(config.AppIDS) == 0 {
			Log.ERROR.Panicf("not found wxpay appid")
		} else {
			wx.appid = config.AppIDS[0]
		}
	}
	if config.Secret != "" {
		wx.secret = config.Secret
	} else {
		Log.ERROR.Panicf("not found wxpay secret")
	}
	if config.ServiceID != "" {
		wx.serviceid = config.ServiceID
	} else {
		Log.ERROR.Panicf("not found wxpay serviceid")
	}
	if config.URL != "" {
		wx.url = config.URL
	} else {
		Log.ERROR.Panicf("not found wxpay apiurl")
	}
	wx.notify = config.Notify
	// 加载证书
	cert, err := LoadCertFromP12(filepath.Join(config.CertPath, "cert.p12"), wx.serviceid)
	if err != nil {
		Log.ERROR.Panicf("oad wxpay cert fail,file:%s,err:%v", config.CertPath, err)
	} else {
		wx.tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}
	return wx
}

// NewWXPaySingle 获取微信实例-单商户
func NewWXPaySingle(cfgs Config) *WXPaySingle {
	wxpaySingle = &WXPaySingle{cfgs: map[string]config{}}
	if cfgs.AppID != "" || len(cfgs.AppIDS) > 0 {
		cfgs.Apps = append([]ConfigSingle{cfgs.ConfigSingle}, cfgs.Apps...)
	}
	for _, cfg := range cfgs.Apps {
		c := wxConfig(cfg)
		if wxpaySingle.def.appid == "" {
			if cfg.AppID != "" {
				wxpaySingle.def = c
			}
		}
		cfg.AppIDS = append(cfg.AppIDS, cfg.AppID)
		for _, appid := range cfg.AppIDS {
			if wxpaySingle.def.appid == "" {
				wxpaySingle.def = c
			}
			c.appid = appid
			wxpaySingle.cfgs[appid] = c
		}
	}
	return wxpaySingle
}

// wxResult 微信返回最外层系统参数
type wxResult struct {
	XMLName xml.Name `xml:"xml"`

	ReturnCode string `xml:"return_code"` // 返回状态码
	ReturnMsg  string `xml:"return_msg"`  // 返回信息

	// when return_code == SUCCESS
	AppID      string `xml:"appid"`        // 公众账号ID
	MchID      string `xml:"mch_id"`       // 商户号
	DeviceInfo string `xml:"device_info"`  // 设备号
	NonceStr   string `xml:"nonce_str"`    // 随机字符串
	Sign       string `xml:"sign"`         // 签名
	ResultCode string `xml:"result_code"`  // 业务结果
	ErrCode    string `xml:"err_code"`     // 错误代码
	ErrCodeDes string `xml:"err_code_des"` // 错误代码描述
}

type wxMchPayRequest struct {
	XMLName xml.Name `xml:"xml"`
	// required
	AppID string `xml:"mch_appid"` // 公众账号ID
	MchID string `xml:"mchid"`     // 商户号
	// SubMchID       string `xml:"sub_mch_id"`       // 子商户ID
	// SubAppID       string `xml:"sub_appid"`        // 子商户公众号ID
	NonceStr       string `xml:"nonce_str"`        // 随机字符串
	OutTradeID     string `xml:"partner_trade_no"` // 商户订单号
	OpenID         string `xml:"openid"`           // appid下对应的用户openid
	CheckName      string `xml:"check_name"`       // NO_CHECK：不校验真实姓名 FORCE_CHECK：强校验真实姓名
	UserName       string `xml:"re_user_name"`     // 收款用户真实姓名
	Amount         int64  `xml:"amount"`           // 订单金额
	Desc           string `xml:"desc"`             // 付款备注
	SpbillCreateIP string `xml:"spbill_create_ip"` // 终端IP
	Sign           string `xml:"sign"`             // 签名
}

// wxMchPayResult 微信企业付款返回结构
type wxMchPayResult struct {
	XMLName xml.Name `xml:"xml"`

	ReturnCode string `xml:"return_code"` // 返回状态码
	ReturnMsg  string `xml:"return_msg"`  // 返回信息

	// when return_code == SUCCESS
	AppID      string `xml:"mch_appid"`    // 公众账号ID
	MchID      string `xml:"mchid"`        // 商户号
	DeviceInfo string `xml:"device_info"`  // 设备号
	NonceStr   string `xml:"nonce_str"`    // 随机字符串
	Sign       string `xml:"sign"`         // 签名
	ResultCode string `xml:"result_code"`  // 业务结果
	ErrCode    string `xml:"err_code"`     // 错误代码
	ErrCodeDes string `xml:"err_code_des"` // 错误代码描述

	OutTradeID string `xml:"partner_trade_no"` // 商户订单号
	TradeID    string `xml:"payment_no"`       // 微信付款单号
}

// MchPay 企业付款 到 微信零钱包
// 单商户模式调用
// 默认不开启真实姓名强验证，传入姓名则开启
func (WS *WXPaySingle) MchPay(ctx *Context, req *MchPay) (tid string, e Error) {
	params := wxMchPayRequest{
		AppID:          ctx.appid(),
		MchID:          ctx.serviceid(),
		NonceStr:       randomString(32),
		OutTradeID:     req.OutTradeID,
		OpenID:         req.Account,
		UserName:       req.UserName,
		Amount:         req.Amount,
		Desc:           req.Desc,
		SpbillCreateIP: req.IPAddr,
	}
	if params.UserName != "" {
		params.CheckName = "FORCE_CHECK"
	} else {
		params.CheckName = "NO_CHECK"
	}
	params.Sign = WS.Signer(ctx, structToMap(params, "xml"))
	postBody, err := xml.Marshal(params)
	rq := requestSimple{
		url:  ctx.url() + "/mmpaymkttransfers/promotion/transfers",
		body: postBody,
		tls:  true,
		ctx:  ctx,
	}
	rq.fs = func(result interface{}, next Status, err error) (interface{}, error) {
		switch next {
		case netConnErr, nextRetry:
			// 超时，异常立刻重试
			time.Sleep(1 * time.Second)
			return WS.Request(rq)
		default:
			return result, err
		}
	}
	info, err := WS.Request(rq)
	if err != nil {
		e.Msg = string(jsonEncode(info))
		if v, ok := wxErrMap[err.Error()]; ok {
			e.Code = v
		} else {
			e.Code = PayErr
		}
	} else {
		// 转账成功
		tmpresult := wxRefundResult{}
		xml.Unmarshal(info.([]byte), &tmpresult)
		tid = tmpresult.TradeID
	}
	return
}

// wxBarPayRequest 微信条码支付请求结构
type wxBarPayRequest struct {
	XMLName xml.Name `xml:"xml"`
	// required
	AppID          string `xml:"appid"`            // 公众账号ID
	MchID          string `xml:"mch_id"`           // 商户号
	SubMchID       string `xml:"sub_mch_id"`       // 子商户ID
	SubAppID       string `xml:"sub_appid"`        // 子商户公众号ID
	NonceStr       string `xml:"nonce_str"`        // 随机字符串
	Body           string `xml:"body"`             // 商品描述
	OutTradeID     string `xml:"out_trade_no"`     // 商户订单号
	Amount         int64  `xml:"total_fee"`        // 订单金额
	AuthCode       string `xml:"auth_code"`        // 授权码
	SpbillCreateIP string `xml:"spbill_create_ip"` // 终端IP
	Sign           string `xml:"sign"`             // 签名
}

// BarPay 商户主动扫码支付
// 单商户模式调用
// 服务商模式请调用 WXPay.BarPay
func (WS *WXPaySingle) BarPay(ctx *Context, req *BarPay) (trade *Trade, e Error) {
	trade = getTrade(map[string]interface{}{"outtradeid": req.OutTradeID})
	if trade.ID != "" && trade.Status == TradeStatusSucc {
		// 如果订单已经存在并且支付，返回报错
		e.Code = PayErrPayed
		return
	}
	params := wxBarPayRequest{
		AppID:          ctx.appid(),
		MchID:          ctx.serviceid(),
		NonceStr:       randomString(32),
		Body:           req.TradeName,
		OutTradeID:     req.OutTradeID,
		Amount:         req.Amount,
		AuthCode:       req.AuthCode,
		SpbillCreateIP: req.IPAddr,
	}
	params.SubMchID = ctx.mchid()
	params.SubAppID = ctx.subappid()
	params.Sign = WS.Signer(ctx, structToMap(params, "xml"))
	postBody, err := xml.Marshal(params)
	// 订单是否需要撤销，支付是否成功
	var needCancel, paySucc bool
	rq := requestSimple{
		url:  ctx.url() + "/pay/micropay",
		body: postBody,
		ctx:  ctx,
	}
	rq.fs = func(result interface{}, next Status, err error) (interface{}, error) {
		switch next {
		case netConnErr:
			// 网络错误
			time.Sleep(1 * time.Second)
			return WS.Request(rq)
		case nextRetry:
			// 支付异常 https://pay.weixin.qq.com/wiki/doc/api/micropay_sl.php?chapter=9_10&index=1
			// 查询订单，如果支付失败，则取消订单
			trade, e = WS.TradeInfo(ctx, &Trade{OutTradeID: req.OutTradeID}, true)
			if e.Code == TradeErrNotFound {
				// 订单不存在 相同参数再次支付
				return WS.Request(rq)
			} else if trade.Status == TradeStatusSucc {
				// 订单支付成功
				paySucc = true
			} else {
				// 其他错误，取消订单
				needCancel = true
			}
		case nextWaitRetry:
			needCancel = true
			// 等待用户输入密码
			// 每3秒获取一次订单信息，直至支付超时或支付成功
			for getNowSec()-ctx.gt() < maxTimeout {
				trade, e = WS.TradeInfo(ctx, &Trade{OutTradeID: req.OutTradeID}, true)
				if e.Code == 0 && trade.Status == TradeStatusSucc {
					// 支付成功
					paySucc = true
					needCancel = false
					return trade, nil
				} else if trade.Status == TradeStatusWaitPay {
					// 订单取消支付
					paySucc = false
					needCancel = true
					return trade, newError("用户取消支付")
				}
				time.Sleep(3 * time.Second)
			}
		default:
			needCancel = true
		}
		return trade, newError(e.Msg)
	}
	info, err := WS.Request(rq)
	if err != nil {
		e.Msg = string(jsonEncode(info))
		if v, ok := wxErrMap[err.Error()]; ok {
			e.Code = v
		} else {
			e.Code = TradeErr
		}
	} else {
		// 请求成功
		// 返回成功
		paySucc = true
		needCancel = false
	}
	if paySucc {
		result := trade
		switch info.(type) {
		case *Trade:
			tmpresult := info.(*Trade)
			result.TradeID = tmpresult.TradeID
			result.Amount = req.Amount
		case []uint8:
			tmpresult := wxTradeResult{}
			xml.Unmarshal(info.([]byte), &tmpresult)
			result.TradeID = tmpresult.TradeID
			result.Amount = req.Amount
		}
		result.From = WXPAY
		result.Type = BARPAY
		result.UserID = ctx.userid()
		result.MchID = ctx.serviceid()
		result.UpTime = ctx.gt()
		result.AppID = ctx.appid()
		result.PayTime = ctx.gt()
		result.Status = TradeStatusSucc
		if result.ID == "" {
			result.OutTradeID = req.OutTradeID
			result.ID = randomTimeString()
			result.Create = ctx.gt()
			// 保存订单
			saveTrade(result)
		} else {
			// 更新订单
			updateTrade(map[string]interface{}{"id": trade.ID}, result)
		}

	}
	if needCancel {
		// 取消订单
		WS.Cancel(ctx, &Trade{OutTradeID: req.OutTradeID})
	}
	return
}

// wxRefundRequest 微信退款请求结构
type wxRefundRequest struct {
	XMLName xml.Name `xml:"xml"`

	AppID    string `xml:"appid"`      // 公众账号ID
	MchID    string `xml:"mch_id"`     // 商户号
	SubMchID string `xml:"sub_mch_id"` // 子商户ID
	SubAppID string `xml:"sub_appid"`  // 子商户公众账号ID
	NonceStr string `xml:"nonce_str"`  // 随机字符串

	OutTradeID string `xml:"out_trade_no"` // 商户订单号
	Amount     int64  `xml:"total_fee"`    // 订单金额
	Sign       string `xml:"sign"`         // 签名

	RefundID     string `xml:"refund_id"`      // 商户退款单号
	RefundAmount int64  `xml:"refund_fee"`     // 退款金额
	RefundDesc   string `xml:"refund_desc"`    // 退款备注
	TradeID      string `xml:"transaction_id"` // 微信订单号
	OutRefundID  string `xml:"out_refund_no"`
}

// wxRefundResult 微信退款返回结构
type wxRefundResult struct {
	XMLName xml.Name `xml:"xml"`

	ReturnCode string `xml:"return_code"` // 返回状态码
	ReturnMsg  string `xml:"return_msg"`  // 返回信息

	// when return_code == SUCCESS
	AppID      string `xml:"appid"`        // 公众账号ID
	MchID      string `xml:"mch_id"`       // 商户号
	DeviceInfo string `xml:"device_info"`  // 设备号
	NonceStr   string `xml:"nonce_str"`    // 随机字符串
	Sign       string `xml:"sign"`         // 签名
	ResultCode string `xml:"result_code"`  // 业务结果
	ErrCode    string `xml:"err_code"`     // 错误代码
	ErrCodeDes string `xml:"err_code_des"` // 错误代码描述

	TradeID      string `xml:"transaction_id"`
	OutTradeID   string `xml:"out_trade_no"`
	OutRefundID  string `xml:"out_refund_no"`
	RefundID     string `xml:"refund_id"`
	RefundAmount int64  `xml:"refund_fee"`
}

// Refund 订单退款
// 单商户模式调用
// 服务商模式请调用 WXPay.Refund
func (WS *WXPaySingle) Refund(ctx *Context, req *Refund) (refund *Refund, e Error) {
	trade, e := WS.TradeInfo(ctx, &Trade{OutTradeID: req.SourceID}, true)
	if trade.TradeID == "" || e.Code == TradeErrNotFound {
		e.Code = TradeErrNotFound
		return
	}
	// 订单存在多次退款情况
	if trade.Status != TradeStatusSucc && trade.Status != TradeStatusRefund {
		e.Code = TradeErrStatus
		return
	}
	params := wxRefundRequest{
		AppID:        ctx.appid(),
		MchID:        ctx.serviceid(),
		NonceStr:     randomString(32),
		OutTradeID:   req.SourceID,
		OutRefundID:  req.OutRefundID,
		RefundAmount: req.Amount,
		RefundDesc:   req.Memo,
		TradeID:      trade.TradeID,
		Amount:       trade.Amount,
	}
	params.SubMchID = ctx.mchid()
	params.SubAppID = ctx.subappid()

	params.Sign = WS.Signer(ctx, structToMap(params, "xml"))
	postBody, err := xml.Marshal(params)
	rq := requestSimple{
		url:  ctx.url() + "/secapi/pay/refund",
		body: postBody,
		tls:  true,
		ctx:  ctx,
	}
	rq.fs = func(result interface{}, next Status, err error) (interface{}, error) {
		switch next {
		case netConnErr, nextRetry:
			// 超时，异常立刻重试
			time.Sleep(1 * time.Second)
			return WS.Request(rq)
		default:
			return result, err
		}
	}
	info, err := WS.Request(rq)
	if err != nil {
		e.Msg = string(jsonEncode(info))
		if v, ok := wxErrMap[err.Error()]; ok {
			e.Code = v
		} else {
			e.Code = RefundErr
		}
	} else {
		// 退款成功
		tmpresult := wxRefundResult{}
		xml.Unmarshal(info.([]byte), &tmpresult)
		refund = &Refund{
			RefundID:    tmpresult.RefundID,
			ID:          randomTimeString(),
			OutRefundID: req.OutRefundID,
			MchID:       params.MchID,
			UserID:      ctx.userid(),
			Amount:      req.Amount,
			SourceID:    req.SourceID,
			Status:      RefundStatusSucc,
			UpTime:      ctx.gt(),
			RefundTime:  ctx.gt(),
			Create:      ctx.gt(),
			Memo:        req.Memo,
			AppID:       ctx.appid(),
			From:        WXPAY,
		}
		saveRefund(refund)
	}
	return
}

// wxCancelRequest 微信撤销订单请求结构
type wxCancelRequest struct {
	XMLName xml.Name `xml:"xml"`

	// required
	AppID      string `xml:"appid"`      // 公众账号ID
	MchID      string `xml:"mch_id"`     // 商户号
	SubMchID   string `xml:"sub_mch_id"` // 子商户ID
	SubAppID   string `xml:"sub_appid"`
	NonceStr   string `xml:"nonce_str"`    // 随机字符串
	Body       string `xml:"body"`         // 商品描述
	OutTradeID string `xml:"out_trade_no"` // 商户订单号
	TradeID    string `xml:"transaction_id"`
	Sign       string `xml:"sign"` // 签名
}

// Cancel 撤销订单
// 单商户模式调用
// 服务商模式请调用 WXPay.Cancel
func (WS *WXPaySingle) Cancel(ctx *Context, req *Trade) (e Error) {
	params := wxCancelRequest{
		AppID:      ctx.appid(),
		MchID:      ctx.serviceid(),
		NonceStr:   randomString(32),
		OutTradeID: req.OutTradeID,
		TradeID:    req.TradeID,
	}

	params.SubMchID = ctx.mchid()
	params.SubAppID = ctx.subappid()
	params.Sign = WS.Signer(ctx, structToMap(params, "xml"))
	postBody, err := xml.Marshal(params)
	rq := requestSimple{
		url:  ctx.url() + "/secapi/pay/reverse",
		body: postBody,
		tls:  true,
		ctx:  ctx,
	}

	rq.fs = func(result interface{}, next Status, err error) (interface{}, error) {
		switch next {
		case netConnErr, nextRetry:
			// 超时，异常立刻重试
			time.Sleep(1 * time.Second)
			return WS.Request(rq)
		default:
			return result, err
		}
	}
	info, err := WS.Request(rq)
	if err != nil {
		e.Msg = string(jsonEncode(info))
		if v, ok := wxErrMap[err.Error()]; ok {
			e.Code = v
		} else {
			e.Code = TradeErr
		}
	} else {
		// 撤销成功
	}
	return e
}

// wxTradeInfoRequest 微信支付 获取订单详情接口请求参数结构
type wxTradeInfoRequest struct {
	XMLName xml.Name `xml:"xml"`

	// required
	AppID      string `xml:"appid"`      // 公众账号ID
	MchID      string `xml:"mch_id"`     // 商户号
	SubMchID   string `xml:"sub_mch_id"` // 子商户ID
	SubAppID   string `xml:"sub_appid"`
	NonceStr   string `xml:"nonce_str"`    // 随机字符串
	Body       string `xml:"body"`         // 商品描述
	OutTradeID string `xml:"out_trade_no"` // 商户订单号
	TradeID    string `xml:"transaction_id"`
	Sign       string `xml:"sign"` // 签名
}

// wxTradeResult 微信订单返回结构
type wxTradeResult struct {
	XMLName xml.Name `xml:"xml"`

	ReturnCode string `xml:"return_code"` // 返回状态码
	ReturnMsg  string `xml:"return_msg"`  // 返回信息

	// when return_code == SUCCESS
	AppID      string `xml:"appid"`        // 公众账号ID
	MchID      string `xml:"mch_id"`       // 商户号
	DeviceInfo string `xml:"device_info"`  // 设备号
	NonceStr   string `xml:"nonce_str"`    // 随机字符串
	Sign       string `xml:"sign"`         // 签名
	ResultCode string `xml:"result_code"`  // 业务结果
	ErrCode    string `xml:"err_code"`     // 错误代码
	ErrCodeDes string `xml:"err_code_des"` // 错误代码描述

	// when return_code == result_code == SUCCESS
	OpenID             string `xml:"openid"`     // 用户标识
	TradeType          string `xml:"trade_type"` // 交易类型
	Amount             int64  `xml:"total_fee"`  // 订单金额
	Status             string `xml:"trade_state"`
	SettlementTotalFee int64  `xml:"settlement_total_fee"` // 应结订单金额
	CouponFee          int64  `xml:"coupon_fee"`           // 代金券金额
	CashFeeType        string `xml:"cash_fee_type"`        // 现金支付货币类型
	CashFee            int64  `xml:"cash_fee"`             // 现金支付金额
	TradeID            string `xml:"transaction_id"`       // 微信支付订单号
	OutTradeID         string `xml:"out_trade_no"`         // 用户订单号
	Attach             string `xml:"attach"`               // 商家数据包
	TimeEnd            string `xml:"time_end"`             // 支付完成时间
}

// TradeInfo 获取订单详情
// 单商户模式调用
// 服务商模式请调用 WXPay.TradeInfo
func (WS *WXPaySingle) TradeInfo(ctx *Context, req *Trade, sync bool) (trade *Trade, e Error) {
	q := bson.M{"from": WXPAY}
	if req.TradeID != "" {
		q["tradeid"] = req.TradeID
	}
	if req.OutTradeID != "" {
		q["outtradeid"] = req.OutTradeID
	}
	trade = getTrade(q)
	if !sync {
		// 不同步的情况直接返回本地查询数据
		if trade.ID == "" {
			e.Code = TradeErrNotFound
		}
		return
	}
	// 同步第三方数据
	params := wxTradeInfoRequest{
		AppID:      ctx.appid(),
		MchID:      ctx.serviceid(),
		NonceStr:   randomString(32),
		OutTradeID: req.OutTradeID,
		TradeID:    req.TradeID,
	}
	params.SubMchID = ctx.mchid()
	params.SubAppID = ctx.subappid()
	params.Sign = WS.Signer(ctx, structToMap(params, "xml"))
	postBody, err := xml.Marshal(params)
	rq := requestSimple{
		url:  ctx.url() + "/pay/orderquery",
		body: postBody,
		ctx:  ctx,
	}
	rq.fs = func(result interface{}, next Status, err error) (interface{}, error) {
		switch next {
		case netConnErr, nextRetry:
			// 超时，异常立刻重试
			time.Sleep(1 * time.Second)
			return WS.Request(rq)
		default:
			return result, err
		}
	}
	info, err := WS.Request(rq)
	if err != nil {
		e.Msg = string(jsonEncode(info))
		if v, ok := wxErrMap[err.Error()]; ok {
			e.Code = v
		} else {
			e.Code = TradeErr
		}
	} else {
		// 请求成功
		tmpresult := wxTradeResult{}
		xml.Unmarshal(info.([]byte), &tmpresult)
		// 数据返回后以第三方返回数据为准
		trade = &Trade{
			Amount:     tmpresult.Amount,
			Status:     wxTradeStatusMap[tmpresult.Status],
			ID:         trade.ID,
			UpTime:     getNowSec(),
			OutTradeID: req.OutTradeID,
			TradeID:    tmpresult.TradeID,
			Create:     trade.Create,
			Type:       trade.Type,
			From:       WXPAY,
			AppID:      trade.AppID,
			PayTime:    str2Sec("20060102150405", tmpresult.TimeEnd),
		}
		trade.UserID = ctx.userid()
		trade.MchID = ctx.serviceid()
		if trade.ID == "" {
			// 本地不存在
			// trade.ID = randomTimeString()
			// trade.Create = getNowSec()
			// err := saveTrade(trade)
			// if err != nil {
			// 	e.Code = SysErrDB
			// 	e.Msg = err.Error()
			// }
		} else {
			// 更新
			err := updateTrade(map[string]interface{}{"id": trade.ID}, trade)
			if err != nil {
				e.Code = SysErrDB
				e.Msg = err.Error()
			}
		}
	}
	return
}

// wxPayParamsRequest 微信支付参数获取请求结构
type wxPayParamsRequest struct {
	XMLName xml.Name `xml:"xml"`

	// required
	AppID      string `xml:"appid"`      // 公众账号ID
	MchID      string `xml:"mch_id"`     // 商户号
	SubMchID   string `xml:"sub_mch_id"` // 子商户ID
	SubAppID   string `xml:"sub_appid"`
	NonceStr   string `xml:"nonce_str"`    // 随机字符串
	Body       string `xml:"body"`         // 商品描述
	OutTradeID string `xml:"out_trade_no"` // 商户订单号
	Amount     string `xml:"total_fee"`
	IPAddr     string `xml:"spbill_create_ip"`
	NotifyURL  string `xml:"notify_url"`
	TradeType  string `xml:"trade_type"`
	SceneInfo  string `xml:"scene_info"`
	Sign       string `xml:"sign"`        // 签名
	OpenID     string `xml:"openid"`      // 与sub_openid二选一 公众号支付必传，openid为在服务商公众号的id
	SubOpenID  string `xml:"sub_openid" ` // 与openid 二选一 公众号支付必传，sub_openid为在子商户公众号的id
}

// wxPayParamsResult 微信支付参数获取返回结构
type wxPayParamsResult struct {
	XMLName xml.Name `xml:"xml"`

	ReturnCode string `xml:"return_code"` // 返回状态码
	ReturnMsg  string `xml:"return_msg"`  // 返回信息

	// when return_code == SUCCESS
	AppID      string `xml:"appid"`        // 公众账号ID
	MchID      string `xml:"mch_id"`       // 商户号
	DeviceInfo string `xml:"device_info"`  // 设备号
	NonceStr   string `xml:"nonce_str"`    // 随机字符串
	Sign       string `xml:"sign"`         // 签名
	ResultCode string `xml:"result_code"`  // 业务结果
	ErrCode    string `xml:"err_code"`     // 错误代码
	ErrCodeDes string `xml:"err_code_des"` // 错误代码描述

	TradeType string `xml:"trade_type"`
	PrePayID  string `xml:"prepay_id"`
	MWEBURL   string `xml:"mweb_url"`
	CodeURL   string `xml:"code_url"`
}

// PayParams 获取支付参数
// 用于前段请求，不想暴露证书的私密信息的可用此方法组装请求参数，前端只负责请求
// 支持的有 JS支付，手机app支付，公众号支付
// APP支付紧支持单商户模式，公众号支付，扫码支付等支持服务商和单商户模式
func (WS *WXPaySingle) PayParams(ctx *Context, req *TradeParams) (data *PayParams, e Error) {
	trade := getTrade(map[string]interface{}{"outtradeid": req.OutTradeID})
	if trade.ID != "" && trade.Status == TradeStatusSucc {
		// 检测订单号是否存在 并且支付成功
		e.Code = TradeErrStatus
		e.Msg = "订单已支付"
		return
	}
	var tradeType string
	switch req.Type {
	case APPPAY:
		tradeType = "APP"
	case JSPAY, MINIPAY:
		tradeType = "JSAPI"
		// 公众号支付 openid subopenid 二者必传其一
		if req.OpenID == "" && req.SubOpenID == "" {
			e.Code = SysErrParams
			e.Msg = " openid subopenid 二者必传其一"
			return
		}
	case CBARPAY:
		tradeType = "NATIVE"
	default:
		tradeType = "NATIVE"
	}
	if req.NotifyURL == "" {
		req.NotifyURL = ctx.Notify()
	}
	params := wxPayParamsRequest{
		AppID:      ctx.appid(),
		MchID:      ctx.serviceid(),
		NonceStr:   randomString(32),
		Body:       req.ItemDes,
		OutTradeID: req.OutTradeID,
		Amount:     fmt.Sprintf("%d", req.Amount),
		IPAddr:     req.IPAddr,
		NotifyURL:  req.NotifyURL,
		TradeType:  tradeType,
		OpenID:     req.OpenID,
		SubOpenID:  req.SubOpenID,
		SceneInfo: string(jsonEncode(map[string]interface{}{
			"h5_info": map[string]interface{}{"type": "Wap", "wap_url": req.Scene.URL, "wap_name": req.Scene.Name},
		})),
	}
	params.SubMchID = ctx.mchid()
	params.SubAppID = ctx.subappid()
	params.Sign = WS.Signer(ctx, structToMap(params, "xml"))
	postBody, err := xml.Marshal(params)
	rq := requestSimple{
		url:  ctx.url() + "/pay/unifiedorder",
		body: postBody,
		ctx:  ctx,
	}
	rq.fs = func(result interface{}, next Status, err error) (interface{}, error) {
		switch next {
		case netConnErr, nextRetry:
			// 超时，异常立刻重试
			time.Sleep(1 * time.Second)
			return WS.Request(rq)
		default:
			return result, err
		}
	}
	info, err := WS.Request(rq)
	Log.DEBUG.Printf("%+v,%+v", info, err)
	if err != nil {
		e.Msg = string(jsonEncode(info))
		if v, ok := wxErrMap[err.Error()]; ok {
			e.Code = v
		} else {
			e.Code = TradeErr
		}
	} else {
		// 请求成功
		tmpresult := wxPayParamsResult{}
		xml.Unmarshal(info.([]byte), &tmpresult)
		data = &PayParams{}
		switch req.Type {
		case APPPAY:
			// app支付返回的是请求参数
			appparams := map[string]string{
				"appid":     ctx.appid(),
				"partnerid": ctx.serviceid(),
				"prepayid":  tmpresult.PrePayID,
				"package":   "Sign=WXPay",
				"noncestr":  randomString(32),
				"timestamp": fmt.Sprintf("%d", getNowSec()),
			}
			appparams["sign"] = WS.Signer(ctx, appparams)
			data.SourceData = string(jsonEncode(appparams))
			data.Params = httpBuildQuery(appparams)
		case MINIPAY, JSPAY:
			// 小程序和公众号支付返回接口组装好的请求参数
			params := map[string]string{
				"appId":     ctx.appid(),
				"timeStamp": fmt.Sprintf("%d", getNowSec()),
				"nonceStr":  randomString(32),
				"package":   fmt.Sprintf("prepay_id=%s", tmpresult.PrePayID),
				"signType":  "MD5",
			}
			params["paySign"] = WS.Signer(ctx, params)
			data.SourceData = string(jsonEncode(params))
			data.Params = httpBuildQuery(params)
		case CBARPAY, WEBPAY:
			// 顾客扫码支付 返回的是二维码地址
			data.SourceData = string(jsonEncode(map[string]string{
				"code_url": tmpresult.CodeURL,
			}))
		default:
			//
			data.SourceData = string(jsonEncode(map[string]string{
				"code_url": tmpresult.CodeURL,
			}))
		}
		newTrade := &Trade{
			OutTradeID: req.OutTradeID,
			Amount:     req.Amount,
			ID:         randomTimeString(),
			Type:       req.Type,
			MchID:      ctx.serviceid(),
			UpTime:     getNowSec(),
			Create:     getNowSec(),
			From:       WXPAY,
			AppID:      ctx.appid(),
		}
		// save tradeinfo
		if trade.ID != "" {
			// 更新
			updateTrade(map[string]interface{}{"outtradeid": trade.OutTradeID}, newTrade)
		} else {
			// 新增
			saveTrade(newTrade)
		}
	}
	return
}

// Signer 微信请求做验签
// 使用支付私钥
func (WS *WXPaySingle) Signer(ctx *Context, data map[string]string) string {
	message := mapSortAndJoin(data, "=", "&", true)
	message += "&key=" + ctx.secret()
	return strings.ToUpper(makeMd5(message))
}

// Request 发送微信请求
func (WS *WXPaySingle) Request(d requestSimple) (result interface{}, err error) {
	var next Status
	if getNowSec()-d.ctx.gt() > maxTimeout {
		return nil, http.ErrHandlerTimeout
	}
	result, next, err = WS.request(d.url, d.body, d.tls, d.ctx)
	if err != nil {
		if d.fs != nil {
			return d.fs(result, next, err)
		}
	}
	return
}

func (WS *WXPaySingle) request(url string, data []byte, tls bool, ctx *Context) (interface{}, Status, error) {
	var body []byte
	var err error
	if tls {
		body, err = postRequestTLS(url, "text/xml", bytes.NewBuffer(data), ctx.tlsConfig())
	} else {
		body, err = postRequest(url, "text/xml", bytes.NewBuffer(data))
	}
	Log.DEBUG.Printf("url:%s,data:%s,tls:%v,err:%v", url, string(data), tls, err)
	if err != nil {
		// 网络发起请求失败
		// 需重试
		return nil, netConnErr, err
	}
	result := wxResult{}
	Log.DEBUG.Printf("WXPay request url:%s,body:%s", url, string(body))
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, nextStop, err
	}
	if result.ReturnCode != "SUCCESS" {
		return nil, nextStop, newError("NOAUTH")
	}
	next, err := WS.errorCheck(result)
	return body, next, err
}

func (WS *WXPaySingle) errorCheck(result wxResult) (Status, error) {
	if result.ResultCode == "SUCCESS" {
		// 成功
		return nextStop, nil
	}
	var code Status
	switch result.ErrCode {
	case "SYSTEMERROR", "BANKERROR":
		// 需确认
		code = nextRetry
	case "USERPAYING":
		// 需循环确认
		code = nextWaitRetry
	default:
	}
	return code, newError(result.ErrCode)
}

var wxErrMap = map[string]int{
	"ORDERPAID":         PayErrPayed,
	"NOAUTH":            AuthErr,
	"AUTHCODEEXPIRE":    PayErrCode,
	"NOTENOUGH":         UserErrBalance,
	"ORDERCLODES":       TradeErrStatus,
	"ORDERREVERSED":     TradeErrStatus,
	"OUT_TRADE_NO_USED": TradeErrStatus,
	"AUTH_CODE_ERROR":   PayErrCode,
	"AUTH_CODE_INVALID": PayErrCode,
	"ORDERNOTEXIST":     TradeErrNotFound,
	"REVERSE_EXPIRE":    RefundErrExpire,
}
var wxTradeStatusMap = map[string]Status{
	"SUCCESS":    TradeStatusSucc,
	"REFUND":     TradeStatusRefund,
	"NOTPAY":     TradeStatusWaitPay,
	"CLOSED":     TradeStatusClose,
	"REVOKED":    TradeStatusClose,
	"USERPAYING": TradeStatusPaying,
}
