package ppp

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"
)

var (
	wxPaySGUrl       string //微信支付请求地址
	wxPaySGAppId     string //微信公众号ID
	wxPaySGMchId     string //微信支付商户号
	wxPaySGNotifyUrl string //异步通知地址
)

const (
	FC_WXSGPAY_TRADEINFO   string = "WXSGPay.TradeInfo" //订单详情
	FC_WXSGPAY_TRADEPARAMS string = "WXSGPay.PayParams" //网站支付参数组装
)

type WXPaySGInit struct {
	AppId      string
	Url        string
	MchId      string
	ApiKey     string
	ConfigPath string
	NotifyUrl  string
}

func (w *WXPaySGInit) Init() {
	wxPaySGUrl = w.Url
	wxPaySGAppId = w.AppId
	wxPaySGMchId = w.MchId
	wxPaySGSecretKey = w.ApiKey
	wxPaySGNotifyUrl = w.NotifyUrl
	loadWXPaySGCertKey(w.ConfigPath)
}

//微信支付接口主体
type WXPaySG struct {
}

// 获取支付单详情
// DOC:https://pay.weixin.qq.com/wiki/doc/api/micropay_sl.php?chapter=9_2
// 传入参数TradeRequest
// 返回参数TradeResult
func (W *WXPaySG) TradeInfo(request *TradeRequest, resp *TradeResult) error {
	Log.DEBUG.Printf("WXSGPay api:TradeInfo,request:%+v", request)
	defer Log.DEBUG.Printf("WXSGPay api:TradeInfo,response:%+v", resp)
	if request.r.time == 0 {
		request.r.time = getNowSec()
	}

	q := bson.M{"source": PAYTYPE_WXPAY}
	if request.TradeId != "" {
		q["tradeid"] = request.TradeId
	}
	if request.OutTradeId != "" {
		q["outtradeid"] = request.OutTradeId
	}
	trade := getTrade(q)
	if request.DisSync {
		if trade.Id == "" {
			resp.Code = TradeErrNotFound
		}
		resp.Data = trade
		return nil
	}
	params := wxTradeInfoRequest{
		AppId:      wxPayAppId,
		MchId:      wxPayMchId,
		NonceStr:   randomString(32),
		OutTradeId: request.OutTradeId,
		TradeId:    request.TradeId,
	}
	// params.SubMchId = request.r.mchid
	// params.SubAppId = auth.AppId
	params.Sign = WXPaySGSigner(structToMap(params, "xml"))
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
			tmpresult := wxTradeResult{}
			xml.Unmarshal(result.([]byte), &tmpresult)
			resp.Data = Trade{
				OutTradeId: tmpresult.OutTradeId,
				TradeId:    tmpresult.TradeId,
				Status:     wxTradeStatusMap[tmpresult.Status],
				Amount:     tmpresult.Amount,
				PayTime:    str2Sec("20060102150405", tmpresult.TimeEnd),
				Id:         trade.Id,
			}
			if trade.Id != "" {
				updateTrade(bson.M{"id": trade.Id}, bson.M{"$set": bson.M{"status": resp.Data.Status, "uptime": getNowSec(), "paytime": resp.Data.PayTime}})
			}
			return nil
		}
	}
	return nil
}

//网页支付
//子商户模式
func (W *WXPaySG) PayParams(request *WapPayRequest, resp *Response) error {
	Log.DEBUG.Printf("WXSGPay api:WapPayParams,request:%+v", request)
	defer Log.DEBUG.Printf("WXSGPay api:WapPayParams,response:%+v", resp)
	var tradeType string
	switch request.TradeType {
	case APPPAYPARAMS:
		tradeType = "APP"
	}
	if request.r.time == 0 {
		request.r.time = getNowSec()
	}

	params := wxWapPayRequest{
		AppId: wxPayAppId,
		MchId: wxPayMchId,
		// SubMchId:   auth.MchId,
		// SubAppId:   auth.AppId,
		NonceStr:   randomString(32),
		Body:       request.ItemDes,
		OutTradeId: request.OutTradeId,
		Amount:     fmt.Sprintf("%d", request.Amount),
		IPAddr:     request.IPAddr,
		NotifyUrl:  wxPayNotifyUrl,
		TradeType:  tradeType,
		OpenId:     request.OpenId,
		SubOpenId:  request.SubOpenId,
		SceneInfo: string(jsonEncode(map[string]interface{}{
			"h5_info": map[string]interface{}{"type": "Wap", "wap_url": request.Scene.Url, "wap_name": request.Scene.Name},
		})),
	}
	if params.TradeType == JSPAYPARAMS {
		if params.OpenId == "" && params.SubOpenId == "" {
			resp.Code = SysErrParams
			resp.SourceData = fmt.Sprintf("trade type:%s,openid or sub_openid must have one", params.TradeType)

		}
	}
	params.Sign = WXPaySigner(structToMap(params, "xml"))
	Log.ERROR.Printf("wxpay params:%+v", params)
	postBody, err := xml.Marshal(params)
	if err != nil {
		resp.Code = SysErrParams
		resp.SourceData = err.Error()
		return nil
	}
	var result interface{}
	var next int
	for getNowSec()-request.r.time < 30 {
		result, next, err = W.request(wxPayUrl+"/pay/unifiedorder", postBody)
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
			tmpresult := wxWapPayResult{}
			xml.Unmarshal(result.([]byte), &tmpresult)
			if params.TradeType == JSPAYPARAMS {
				jsParams := map[string]string{
					"appId":     wxPayAppId,
					"timeStamp": fmt.Sprintf("%d", getNowSec()),
					"nonceStr":  randomString(32),
					"package":   "prepay_id=" + tmpresult.PrePayId,
					"signType":  "MD5",
				}
				jsParams["paySign"] = WXPaySigner(jsParams)
				resp.SourceData = string(jsonEncode(jsParams))
			} else {
				resp.SourceData = string(jsonEncode(map[string]string{
					"perpay_id": tmpresult.PrePayId,
					"mweb_url":  tmpresult.MWEBURL,
				}))
			}

			//save tradeinfo
			saveTrade(Trade{
				OutTradeId: request.OutTradeId,
				Status:     0,
				Type:       1,
				Amount:     request.Amount,
				Source:     PAYTYPE_WXPAY,
				UpTime:     getNowSec(),
				Ex:         request.Ex,
				Id:         randomTimeString(), // PPPID
			})
			return nil
		}
	}
	return nil
}

func (w *WXPaySG) requestTls(url string, data []byte) (interface{}, int, error) {
	body, err := postRequestTls(url, "text/xml", bytes.NewBuffer(data), wxPayCertTlsConfig)
	if err != nil {
		//网络发起请求失败
		//需重试
		return nil, -1, err
	}
	result := wxResult{}
	Log.DEBUG.Printf("WXSGPay request url:%s,body:%s", url, string(body))

	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, 0, err
	}
	if result.ReturnCode != "SUCCESS" {
		return nil, -1, newError(result.ReturnCode + result.ReturnMsg)
	}
	next, err := w.errorCheck(result)
	return body, next, err
}
func (w *WXPaySG) request(url string, data []byte) (interface{}, int, error) {
	body, err := postRequest(url, "text/xml", bytes.NewBuffer(data))
	if err != nil {
		//网络发起请求失败
		//需重试
		return nil, -1, err
	}
	result := wxResult{}
	Log.DEBUG.Printf("WXSGPay request url:%s,body:%s", url, string(body))
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, 0, err
	}
	if result.ReturnCode != "SUCCESS" {
		return nil, 0, newError("NOAUTH")
	}
	next, err := w.errorCheck(result)
	return body, next, err
}

func (w *WXPaySG) errorCheck(result wxResult) (int, error) {
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
