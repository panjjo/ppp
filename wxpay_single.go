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
	// WXPAYAPP 微信支付app支付标识
	WXPAYAPP string = "wxpay_app"
)

// WXPaySingle 微信支付单商户模式主体
// 微信支付服务商模式
// 服务商模式与单商户模式区别只是多了一个 子商户权限，其余接口结构返回完全一致
type WXPaySingle struct {
	appid     string
	tlsConfig *tls.Config // 应用私钥
	secret    string      // 支付密钥
	url       string
	serviceid string
	notify    string // 异步回调地址
	t         string
	rs        rs
}

// WXPaySingleForAPP 微信单商户模式 app支付实例
type WXPaySingleForAPP struct {
	WXPaySingle
}

// NewWXPaySingleForAPP 获取微信单商户模式 APP支付
func NewWXPaySingleForAPP(config Config) *WXPaySingleForAPP {
	return &WXPaySingleForAPP{*NewWXPaySingle(config)}
}

// NewWXPaySingle 获取微信实例-单商户
func NewWXPaySingle(config Config) *WXPaySingle {
	wx := &WXPaySingle{}
	if config.AppID != "" {
		wx.appid = config.AppID
	} else {
		Log.ERROR.Panicf("not found wxpay appid")
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
	wx.t = WXPAYSINGLE
	return wx
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
func (WS *WXPaySingle) BarPay(req *BarPay) (trade *Trade, e Error) {
	trade = getTrade(map[string]interface{}{"outtradeid": req.OutTradeID})
	if trade.ID != "" && trade.Status == TradeStatusSucc {
		// 如果订单已经存在并且支付，返回报错
		e.Code = PayErrPayed
		return
	}
	params := wxBarPayRequest{
		AppID:          WS.appid,
		MchID:          WS.serviceid,
		NonceStr:       randomString(32),
		Body:           req.TradeName,
		OutTradeID:     req.OutTradeID,
		Amount:         req.Amount,
		AuthCode:       req.AuthCode,
		SpbillCreateIP: req.IPAddr,
	}
	if WS.rs.auth != nil {
		params.SubMchID = WS.rs.auth.MchID
		params.SubAppID = WS.rs.auth.AppID
	}
	params.Sign = WS.Signer(structToMap(params, "xml"))
	postBody, err := xml.Marshal(params)
	// 订单是否需要撤销，支付是否成功
	var needCancel, paySucc bool
	rq := requestSimple{
		url:  WS.url + "/pay/micropay",
		body: postBody,
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
			trade, e = WS.TradeInfo(&Trade{OutTradeID: req.OutTradeID}, true)
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
			for getNowSec()-WS.rs.t < maxTimeout {
				trade, e = WS.TradeInfo(&Trade{OutTradeID: req.OutTradeID}, true)
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
		result.From = WS.t
		result.Type = BARPAY
		result.UserID = WS.rs.userid
		if WS.rs.auth != nil {
			result.MchID = WS.rs.auth.MchID
		}
		result.UpTime = WS.rs.t
		result.PayTime = WS.rs.t
		result.Status = TradeStatusSucc
		if result.ID == "" {
			result.OutTradeID = req.OutTradeID
			result.ID = randomTimeString()
			result.Create = WS.rs.t
			// 保存订单
			saveTrade(trade)
		} else {
			// 更新订单
			updateTrade(map[string]interface{}{"id": trade.ID}, trade)
		}

	}
	if needCancel {
		// 取消订单
		WS.Cancel(&Trade{OutTradeID: req.OutTradeID})
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
func (WS *WXPaySingle) Refund(req *Refund) (refund *Refund, e Error) {
	trade, e := WS.TradeInfo(&Trade{OutTradeID: req.SourceID}, true)
	if trade.TradeID == "" || e.Code == TradeErrNotFound {
		e.Code = TradeErrNotFound
		return
	}
	if trade.Status != TradeStatusSucc {
		e.Code = TradeErrStatus
		return
	}
	params := wxRefundRequest{
		AppID:        WS.appid,
		MchID:        WS.serviceid,
		NonceStr:     randomString(32),
		OutTradeID:   req.SourceID,
		OutRefundID:  req.OutRefundID,
		RefundAmount: req.Amount,
		RefundDesc:   req.Memo,
		TradeID:      trade.TradeID,
		Amount:       trade.Amount,
	}
	if WS.rs.auth != nil {
		params.SubMchID = WS.rs.auth.MchID
		params.SubAppID = WS.rs.auth.AppID
	}

	params.Sign = WS.Signer(structToMap(params, "xml"))
	postBody, err := xml.Marshal(params)
	rq := requestSimple{
		url:  WS.url + "/secapi/pay/refund",
		body: postBody,
		tls:  true,
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
			UserID:      WS.rs.userid,
			Amount:      req.Amount,
			SourceID:    req.SourceID,
			Status:      RefundStatusSucc,
			UpTime:      WS.rs.t,
			RefundTime:  WS.rs.t,
			Create:      WS.rs.t,
			Memo:        req.Memo,
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
func (WS *WXPaySingle) Cancel(req *Trade) (e Error) {
	params := wxCancelRequest{
		AppID:      WS.appid,
		MchID:      WS.serviceid,
		NonceStr:   randomString(32),
		OutTradeID: req.OutTradeID,
		TradeID:    req.TradeID,
	}

	if WS.rs.auth != nil {
		params.SubMchID = WS.rs.auth.MchID
		params.SubAppID = WS.rs.auth.AppID
	}
	params.Sign = WS.Signer(structToMap(params, "xml"))
	postBody, err := xml.Marshal(params)
	rq := requestSimple{
		url:  WS.url + "/secapi/pay/reverse",
		body: postBody,
		tls:  true,
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
func (WS *WXPaySingle) TradeInfo(req *Trade, sync bool) (trade *Trade, e Error) {
	q := bson.M{"from": WS.t}
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
		AppID:      WS.appid,
		MchID:      WS.serviceid,
		NonceStr:   randomString(32),
		OutTradeID: req.OutTradeID,
		TradeID:    req.TradeID,
	}
	if WS.rs.auth != nil {
		params.SubMchID = WS.rs.auth.MchID
		params.SubAppID = WS.rs.auth.AppID
	}
	params.Sign = WS.Signer(structToMap(params, "xml"))
	postBody, err := xml.Marshal(params)
	rq := requestSimple{
		url:  WS.url + "/pay/orderquery",
		body: postBody,
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
			From:       WS.t,
			PayTime:    str2Sec("20060102150405", tmpresult.TimeEnd),
		}
		trade.UserID = WS.rs.userid
		if WS.rs.auth != nil {
			trade.MchID = WS.rs.auth.MchID
		}
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
func (WS *WXPaySingle) PayParams(req *TradeParams) (data *PayParams, e Error) {
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
	params := wxPayParamsRequest{
		AppID:      WS.appid,
		MchID:      WS.serviceid,
		NonceStr:   randomString(32),
		Body:       req.ItemDes,
		OutTradeID: req.OutTradeID,
		Amount:     fmt.Sprintf("%d", req.Amount),
		IPAddr:     req.IPAddr,
		NotifyURL:  WS.notify,
		TradeType:  tradeType,
		OpenID:     req.OpenID,
		SubOpenID:  req.SubOpenID,
		SceneInfo: string(jsonEncode(map[string]interface{}{
			"h5_info": map[string]interface{}{"type": "Wap", "wap_url": req.Scene.URL, "wap_name": req.Scene.Name},
		})),
	}
	if WS.rs.auth != nil {
		params.SubMchID = WS.rs.auth.MchID
		params.SubAppID = WS.rs.auth.AppID
	}
	params.Sign = WS.Signer(structToMap(params, "xml"))
	postBody, err := xml.Marshal(params)
	rq := requestSimple{
		url:  WS.url + "/pay/unifiedorder",
		body: postBody,
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
				"appid":     WS.appid,
				"partnerid": WS.serviceid,
				"prepayid":  tmpresult.PrePayID,
				"package":   "Sign=WXPay",
				"noncestr":  randomString(32),
				"timestamp": fmt.Sprintf("%d", getNowSec()),
			}
			appparams["sign"] = WS.Signer(appparams)
			data.SourceData = string(jsonEncode(appparams))
			data.Params = httpBuildQuery(appparams)
		case MINIPAY:
			// 小程序支付返回接口组装好的请求参数
			params := map[string]string{
				"appid":     WS.appid,
				"timeStamp": fmt.Sprintf("%d", getNowSec()),
				"noncestr":  randomString(32),
				"package":   tmpresult.PrePayID,
				"signType":  "MD5",
			}
			params["paySign"] = WS.Signer(params)
			data.SourceData = string(jsonEncode(params))
			data.Params = httpBuildQuery(params)
		case JSPAY:
			// 公众号支付 返回预支付id
			data.SourceData = string(jsonEncode(map[string]string{
				"perpay_id": tmpresult.PrePayID,
			}))
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
			MchID:      WS.serviceid,
			UpTime:     getNowSec(),
			Create:     getNowSec(),
			From:       WS.t,
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
func (WS *WXPaySingle) Signer(data map[string]string) string {
	message := mapSortAndJoin(data, "=", "&", true)
	message += "&key=" + WS.secret
	return strings.ToUpper(makeMd5(message))
}

// Request 发送微信请求
func (WS *WXPaySingle) Request(d requestSimple) (result interface{}, err error) {
	var next Status
	if WS.rs.t == 0 {
		WS.rs.t = getNowSec()
	}
	if getNowSec()-WS.rs.t > maxTimeout {
		return nil, http.ErrHandlerTimeout
	}
	result, next, err = WS.request(d.url, d.body, d.tls)
	if err != nil {
		if d.fs != nil {
			return d.fs(result, next, err)
		}
	}
	return
}

func (WS *WXPaySingle) request(url string, data []byte, tls bool) (interface{}, Status, error) {
	var body []byte
	var err error
	if tls {
		body, err = postRequestTLS(url, "text/xml", bytes.NewBuffer(data), WS.tlsConfig)
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
