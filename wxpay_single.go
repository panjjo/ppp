package ppp

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"
)

var wxPaySGConfigMap map[string]wxPaySGConfig

type wxPaySGConfig struct {
	Url        string //微信支付请求地址
	AppId      string //微信公众号ID
	MchId      string //微信支付商户号
	NotifyUrl  string //异步通知地址
	ConfigPath string
	Type       string
	ApiKey     string
	Cert       *tls.Config
}

const (
	FC_WXPAYSG_TRADEINFO   string = "WXPaySG.TradeInfo" //订单详情
	FC_WXPAYSG_TRADEPARAMS string = "WXPaySG.PayParams" //网站支付参数组装
)

type WXPaySGInit struct {
	AppId      string
	Url        string
	MchId      string
	ApiKey     string
	ConfigPath string
	NotifyUrl  string
	Type       string
}

func (w *WXPaySGInit) Init() {
	if wxPaySGConfigMap == nil {
		wxPaySGConfigMap = map[string]wxPaySGConfig{}
	}
	wxPaySGConfigMap[w.Type] = wxPaySGConfig{
		AppId:      w.AppId,
		Url:        w.Url,
		MchId:      w.MchId,
		ApiKey:     w.ApiKey,
		ConfigPath: w.ConfigPath,
		NotifyUrl:  w.NotifyUrl,
		Type:       w.Type,
		Cert:       loadWXPaySGCertKey(w.ConfigPath, w.Type, w.MchId),
	}
	fmt.Printf("%+v,%v\n", wxPaySGConfigMap, w.Type)
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

	// q := bson.M{"source": PAYTYPE_WXPAYSG}
	q := bson.M{}
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
	var config wxPaySGConfig
	switch trade.Source {
	case PAYTYPE_WXPAYSG + APPPAYPARAMS:
		config = wxPaySGConfigMap["app"]
	default:
		config = wxPaySGConfigMap["other"]

	}
	params := wxTradeInfoRequest{
		AppId:      config.AppId,
		MchId:      config.MchId,
		NonceStr:   randomString(32),
		OutTradeId: request.OutTradeId,
		TradeId:    request.TradeId,
	}
	// params.SubMchId = request.r.mchid
	// params.SubAppId = auth.AppId
	params.Sign = WXPaySGSigner(structToMap(params, "xml"), config.ApiKey)
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
	var config wxPaySGConfig
	switch request.TradeType {
	case APPPAYPARAMS:
		tradeType = "APP"
		config = wxPaySGConfigMap["app"]
	case WAPPAYPARAMS:
		tradeType = "NATIVE"
		config = wxPaySGConfigMap["other"]
	}
	if request.r.time == 0 {
		request.r.time = getNowSec()
	}

	params := wxWapPayRequest{
		AppId:      config.AppId,
		MchId:      config.MchId,
		NonceStr:   randomString(32),
		Body:       request.ItemDes,
		OutTradeId: request.OutTradeId,
		Amount:     fmt.Sprintf("%d", request.Amount),
		IPAddr:     request.IPAddr,
		NotifyUrl:  config.NotifyUrl,
		TradeType:  tradeType,
		OpenId:     request.OpenId,
		SceneInfo: string(jsonEncode(map[string]interface{}{
			"h5_info": map[string]interface{}{"type": "Wap", "wap_url": request.Scene.Url, "wap_name": request.Scene.Name},
		})),
	}
	// if params.TradeType == JSPAYPARAMS {
	// 	if params.OpenId == "" && params.SubOpenId == "" {
	// 		resp.Code = SysErrParams
	// 		resp.SourceData = fmt.Sprintf("trade type:%s,openid or sub_openid must have one", params.TradeType)

	// 	}
	// }
	params.Sign = WXPaySGSigner(structToMap(params, "xml"), config.ApiKey)
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
		result, next, err = W.request(config.Url+"/pay/unifiedorder", postBody)
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
			switch request.TradeType {
			case APPPAYPARAMS:
				appParams := map[string]string{
					"appid":     config.AppId,
					"partnerid": config.MchId,
					"prepayid":  tmpresult.PrePayId,
					"package":   "Sign=WXPay",
					"noncestr":  randomString(32),
					"timestamp": fmt.Sprintf("%d", getNowSec()),
				}
				appParams["sign"] = WXPaySGSigner(appParams, config.ApiKey)
				resp.SourceData = string(jsonEncode(appParams))
			default:
				resp.SourceData = string(jsonEncode(map[string]string{
					"perpay_id": tmpresult.PrePayId,
					"mweb_url":  tmpresult.MWEBURL,
					"code_url":  tmpresult.CodeUrl,
				}))
			}

			//save tradeinfo
			saveTrade(Trade{
				OutTradeId: request.OutTradeId,
				Status:     0,
				Type:       1,
				Amount:     request.Amount,
				Source:     PAYTYPE_WXPAYSG + request.TradeType,
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
