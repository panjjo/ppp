package ppp

import (
	"bytes"
	"encoding/xml"
	"time"
)

var (
	wxPayUrl   string //微信支付请求地址
	wxPayAppId string //微信公众号ID
	wxPayMchId string //微信支付商户号
)

type WXPayInit struct {
	AppId      string
	Url        string
	MchId      string
	ConfigPath string
}

func (w *WXPayInit) Init() {
	wxPayUrl = w.Url
	wxPayAppId = w.AppId
	wxPayMchId = w.MchId
	loadWXPayCertKey(w.ConfigPath)
}

type wxRequest struct {
	XMLName xml.Name `xml:"xml"`

	// required
	AppId          string `xml:"appid"`            // 公众账号ID
	MchId          string `xml:"mch_id"`           // 商户号
	SubMchId       string `xml:"sub_mch_id"`       // 子商户ID
	NonceStr       string `xml:"nonce_str"`        // 随机字符串
	Body           string `xml:"body"`             // 商品描述
	OutTradeNo     string `xml:"out_trade_no"`     // 商户订单号
	TotalFee       int64  `xml:"total_fee"`        // 订单金额
	AuthCode       string `xml:"auth_code"`        // 授权码
	SpbillCreateIp string `xml:"spbill_create_ip"` // 终端IP
	Sign           string `xml:"sign"`             // 签名

	// optional
	DeviceInfo string `xml:"device_info"` // 设备号
	SignType   string `xml:"sign_type"`   // 签名类型
	Detail     string `xml:"detail"`      // 商品详情
	Attach     string `xml:"attach"`      // 附加数据
	FeeType    string `xml:"fee_type"`    // 货币类型
	GoodsTag   string `xml:"goods_tag"`   // 商品标记

	//refund
	RefundId     string `xml:"out_refund_no"`  //商户退款单号
	RefundAmount int64  `xml:"refund_fee"`     // 退款金额
	RefundDesc   string `xml:"refund_desc"`    //退款备注
	TradeId      string `xml:"transaction_id"` //微信订单号
}

//微信支付接口主体
type WXPay struct {
}

// 统一收单支付接口
// DOC:https://pay.weixin.qq.com/wiki/doc/api/micropay_sl.php?chapter=9_10&index=1
// 传入参数为 BarCodePayRequest格式
// 返回参数为 TradeResult
// userid 为收款方自定义id,应存在签约授权成功后保存的对应关系,传空表示收款到开发者支付宝帐号
func (W *WXPay) BarCodePay(request *BarCodePayRequest, resp *TradeResult) error {
	if request.r.time == 0 {
		request.r.time = getNowSec()
	}
	params := wxRequest{
		AppId:          wxPayAppId,
		MchId:          wxPayMchId,
		NonceStr:       randomString(32),
		Body:           request.TradeName,
		OutTradeNo:     request.OutTradeId,
		TotalFee:       request.Amount,
		AuthCode:       request.AuthCode,
		SpbillCreateIp: "",
	}
	user := getUser(request.UserId, PAYTYPE_WXPAY)
	if user.UserId == "" {
		resp.Code = AuthErr
		return nil
	}
	params.SubMchId = user.UserId
	params.Sign = WXPaySigner(structToMap(params, "xml"))
	postBody, err := xml.Marshal(params)
	if err != nil {
		resp.Code = SysErrParams
		resp.SourceData = err.Error()
		return nil
	}
	//请求并除错
	var result interface{}
	var next int
	var needCancel, paySucc bool
	var trade TradeResult
	for getNowSec()-request.r.time < 30 {
		result, next, err = W.request(wxPayUrl+"/pay/micropay", postBody)
		resp.SourceData = string(jsonEncode(result))
		if err != nil {
			if v, ok := wxErrMap[err.Error()]; ok {
				resp.Code = v
			} else {
				resp.Code = PayErr
			}
			if next == 2 {
				//系统支付异常
				W.TradeInfo(&TradeRequest{OutTradeId: request.OutTradeId, r: request.r}, &trade)
				if trade.Code == TradeErrNotFound {
					//订单不存在，重试
					time.Sleep(1 * time.Second)
				} else if trade.Data.Status == 1 {
					paySucc = true
					//存在支付成功
					break
				} else {
					//其他情况撤销
					needCancel = true
				}
			} else if next == 3 {
				needCancel = true
				//等待用户输入密码
				//循环
				//获取一次一直到成功
				for getNowSec()-request.r.time < 30 {
					W.TradeInfo(&TradeRequest{OutTradeId: request.OutTradeId, r: request.r}, &trade)
					if trade.Code == 0 && trade.Data.Status == 1 {
						//订单存在且支付
						//取消撤销
						needCancel = false
						paySucc = true
						break
					}
					time.Sleep(5 * time.Second)
				}
				break
			} else if next == 0 {
				//其他错误，直接返回
				break
			}
			// -1
			//网络异常 1s后重试
			time.Sleep(1 * time.Second)
		} else {
			//支付成功返回
			paySucc = true
			break
		}
	}
	//支付成功
	if paySucc {
		tmpresult := map[string]interface{}{}
		xml.Unmarshal(result.([]byte), &tmpresult)
		resp.Data = Trade{
			Id:         randomTimeString(),
			Amount:     int64(tmpresult["total_fee"].(float64)),
			OutTradeId: request.OutTradeId,
			Source:     PAYTYPE_WXPAY,
			PayTime:    request.r.time,
			UpTime:     request.r.time,
			Type:       1,
			Status:     1,
			TradeId:    tmpresult["transaction_id"].(string),
		}
		saveTrade(resp.Data)
	}
	//撤销
	if needCancel {
		response := Response{}
		W.Cancel(&TradeRequest{OutTradeId: request.OutTradeId, r: request.r}, &response)
	}
	return nil
}

// 交易退款
// DOC:https://pay.weixin.qq.com/wiki/doc/api/micropay_sl.php?chapter=9_4
func (W *WXPay) Refund(request *RefundRequest, resp *TradeResult) error {
	if request.r.time == 0 {
		request.r.time = getNowSec()
	}
	params := wxRequest{
		AppId:          wxPayAppId,
		MchId:          wxPayMchId,
		NonceStr:       randomString(32),
		OutTradeNo:     request.OutTradeId,
		RefundId:       request.RefundId,
		RefundAmount:   request.Amount,
		RefundDesc:     request.Memo,
		SpbillCreateIp: "",
	}
	trade := TradeResult{}
	W.TradeInfo(&TradeRequest{r: request.r, UserId: request.UserId, OutTradeId: request.OutTradeId}, &trade)
	if trade.Code != 0 {
		resp.Code = trade.Code
		resp.SourceData = trade.SourceData
		return nil
	}
	params.TradeId = trade.Data.TradeId
	params.TotalFee = trade.Data.Amount

	user := getUser(request.UserId, PAYTYPE_WXPAY)
	if user.UserId == "" {
		resp.Code = AuthErr
		return nil
	}
	params.SubMchId = user.UserId
	params.Sign = WXPaySigner(structToMap(params, "xml"))
	postBody, err := xml.Marshal(params)
	if err != nil {
		resp.Code = SysErrParams
		resp.SourceData = err.Error()
		return nil
	}
	var result interface{}
	var next int
	for getNowSec()-request.r.time < 30 {
		result, next, err = W.request(wxPayUrl+"secapi/pay/refund", postBody)
		resp.SourceData = string(jsonEncode(result))
		if err != nil {
			if v, ok := aliErrMap[err.Error()]; ok {
				resp.Code = v
			} else {
				resp.Code = RefundErr
			}
			switch next {
			case 2, -1:
				//网络异常 1s后重试
				time.Sleep(1 * time.Second)
			default:
				//其他错误，直接返回
				return nil
			}
		} else {
			//成功返回
			tmpresult := map[string]interface{}{}
			xml.Unmarshal(result.([]byte), &tmpresult)
			resp.Data = Trade{
				Id:         randomTimeString(),
				Amount:     int64(tmpresult["refund_fee"].(float64)),
				OutTradeId: request.RefundId,
				Source:     PAYTYPE_WXPAY,
				Type:       -1,
				PayTime:    request.r.time,
				UpTime:     request.r.time,
				Memo:       request.Memo,
				Status:     1,
				TradeId:    tmpresult["refund_id"].(string),
			}
			saveTrade(resp.Data)
			return nil
		}
	}
	return nil
}

// 交易撤销
// DOC:https://pay.weixin.qq.com/wiki/doc/api/micropay_sl.php?chapter=9_11&index=3
// 入参 TradeRequest
// 出参 Response
func (W *WXPay) Cancel(request *TradeRequest, resp *Response) error {
	if request.r.time == 0 {
		request.r.time = getNowSec()
	}
	params := wxRequest{
		AppId:          wxPayAppId,
		MchId:          wxPayMchId,
		NonceStr:       randomString(32),
		OutTradeNo:     request.OutTradeId,
		SpbillCreateIp: "",
	}
	user := getUser(request.UserId, PAYTYPE_WXPAY)
	if user.UserId == "" {
		resp.Code = AuthErr
		return nil
	}
	params.SubMchId = user.UserId
	params.Sign = WXPaySigner(structToMap(params, "xml"))
	postBody, err := xml.Marshal(params)
	if err != nil {
		resp.Code = SysErrParams
		resp.SourceData = err.Error()
		return nil
	}
	var result interface{}
	var next int
	result, next, err = W.requestTls(wxPayUrl+"/secapi/pay/reverse", postBody)
	resp.SourceData = string(jsonEncode(result))
	if err != nil {
		if v, ok := wxErrMap[err.Error()]; ok {
			resp.Code = v
		} else {
			resp.Code = TradeErr
		}
		switch next {
		case 2, -1:
			//网络异常 1s后重试
			time.Sleep(1 * time.Second)
		default:
			//其他错误，直接返回
			return nil
		}
	} else {
		//成功返回
		return nil
	}
	return nil
}

// 获取支付单详情
// DOC:https://pay.weixin.qq.com/wiki/doc/api/micropay_sl.php?chapter=9_2
// 传入参数TradeRequest
// 返回参数TradeResult
func (W *WXPay) TradeInfo(request *TradeRequest, resp *TradeResult) error {
	if request.r.time == 0 {
		request.r.time = getNowSec()
	}
	user := getUser(request.UserId, PAYTYPE_WXPAY)
	if user.UserId == "" {
		resp.Code = AuthErr
		return nil
	}
	params := wxRequest{
		AppId:          wxPayAppId,
		MchId:          wxPayMchId,
		NonceStr:       randomString(32),
		OutTradeNo:     request.OutTradeId,
		SpbillCreateIp: "",
	}
	params.SubMchId = user.UserId
	params.Sign = WXPaySigner(structToMap(params, "xml"))
	postBody, err := xml.Marshal(params)
	if err != nil {
		resp.Code = SysErrParams
		resp.SourceData = err.Error()
		return nil
	}
	var result interface{}
	var next int
	for getNowSec()-request.r.time < 30 {
		result, next, err = W.request(wxPayUrl+"/pay/orderquery", postBody)
		resp.SourceData = string(jsonEncode(result))
		if err != nil {
			if v, ok := wxErrMap[err.Error()]; ok {
				resp.Code = v
			} else {
				resp.Code = TradeErr
			}
			switch next {
			case 2, -1:
				//网络异常 1s后重试
				time.Sleep(1 * time.Second)
			default:
				//其他错误，直接返回
				return nil
			}
		} else {
			//成功返回
			trade := map[string]interface{}{}
			xml.Unmarshal(result.([]byte), &trade)
			resp.Data = Trade{
				OutTradeId: trade["out_trade_no"].(string),
				TradeId:    trade["transaction_id"].(string),
				Status:     wxTradeStatusMap[trade["trade_state"].(string)],
				Amount:     int64(trade["total_fee"].(float64)),
			}
			return nil
		}
	}
	return nil
}

type wxResult struct {
	XMLName xml.Name `xml:"xml"`

	ReturnCode string `xml:"return_code"` // 返回状态码
	ReturnMsg  string `xml:"return_msg"`  // 返回信息

	// when return_code == SUCCESS
	AppId      string `xml:"appid"`        // 公众账号ID
	MchId      string `xml:"mch_id"`       // 商户号
	DeviceInfo string `xml:"device_info"`  // 设备号
	NonceStr   string `xml:"nonce_str"`    // 随机字符串
	Sign       string `xml:"sign"`         // 签名
	ResultCode string `xml:"result_code"`  // 业务结果
	ErrCode    string `xml:"err_code"`     // 错误代码
	ErrCodeDes string `xml:"err_code_des"` // 错误代码描述

}

func (w *WXPay) requestTls(url string, data []byte) (interface{}, int, error) {
	body, err := postRequestTls(url, "text/xml", bytes.NewBuffer(data), wxPayCertTlsConfig)
	if err != nil {
		//网络发起请求失败
		//需重试
		return nil, -1, err
	}
	result := wxResult{}
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, 0, err
	}
	if result.ReturnCode != "SUCCESS" {
		return nil, -1, newError(result.ReturnCode + result.ReturnMsg)
	}
	next, err := w.errorCheck(result)
	return body, next, err
}
func (w *WXPay) request(url string, data []byte) (interface{}, int, error) {
	body, err := postRequest(url, "text/xml", bytes.NewBuffer(data))
	if err != nil {
		//网络发起请求失败
		//需重试
		return nil, -1, err
	}
	result := wxResult{}
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, 0, err
	}
	if result.ReturnCode != "SUCCESS" {
		return nil, -1, newError(result.ReturnCode + result.ReturnMsg)
	}
	next, err := w.errorCheck(result)
	return body, next, err
}

func (w *WXPay) errorCheck(result wxResult) (int, error) {
	if result.ResultCode == "SUCCESS" {
		//成功
		return 0, nil
	}
	var code int
	switch result.ErrCode {
	case "SYSTEMERROR", "BANKERROR":
		//需确认
		code = 2
	case "USERPAYING":
		//需循环确认
		code = 3
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
	"USERPAYING": TradeStatusWaitPay,
}
