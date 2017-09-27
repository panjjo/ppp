package ppp

import (
	"fmt"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"
)

var (
	aliPayDefaultFormat     string = "JSON"  //默认格式
	aliPayDefaultCharset    string = "utf-8" //默认编码
	aliPayDefaultSignType   string = "RSA2"  //默认加密方式
	aliPayUrl               string           //alipay的地址
	aliPayServiceProviderId string           //收佣商户号
	aliPayAppId             string           //应用ID
)

const (
	FC_ALIPAY_BARCODEPAY   string = "AliPay.BarCodePay"
	FC_ALIPAY_CANCEL       string = "AliPay.Cancel"
	FC_ALIPAY_REFRESHTOKEN string = "AliPay.RefreshToken"
	FC_ALIPAY_REFUND       string = "AliPay.Refund"
	FC_ALIPAY_TRADEINFO    string = "AliPay.TradeInfo"
)

type AliPayInit struct {
	AppId             string
	Url               string
	ServiceProviderId string
	ConfigPath        string
}

func (a *AliPayInit) Init() {
	aliPayUrl = a.Url
	aliPayAppId = a.AppId
	aliPayServiceProviderId = a.ServiceProviderId
	loadAliPayCertKey(a.ConfigPath)
}

//支付宝接口主体
type AliPay struct {
}

// 统一收单支付接口
// DOC:https://docs.open.alipay.com/api_1/alipay.trade.pay
// 传入参数为 BarCodePayRequest格式
// 返回参数为 TradeResult
// userid 为收款方自定义id,应存在签约授权成功后保存的对应关系,传空表示收款到开发者支付宝帐号
func (A *AliPay) BarCodePay(request *BarCodePayRequest, resp *TradeResult) error {
	if request.r.time == 0 {
		request.r.time = getNowSec()
	}
	params := map[string]interface{}{
		"out_trade_no": request.OutTradeId,
		"scene":        "bar_code",
		"auth_code":    request.AuthCode,
		"subject":      request.TradeName,
		"total_amount": float64(request.Amount) / 100.0,
		"body":         request.ItemDes,
		"store_id":     request.ShopId,
	}
	//设置反佣系统商编号
	if aliPayServiceProviderId != "" {
		params["extend_params"] = map[string]interface{}{"sys_service_provider_id": aliPayServiceProviderId}
	}
	//组装系统参数
	sysParams := A.sysParams()
	sysParams["method"] = "alipay.trade.pay"
	sysParams["biz_content"] = string(jsonEncode(params))
	//设置子商户数据
	user := getUser(request.UserId, "alipay")
	if user.UserId == "" {
		resp.Code = AuthErr
		return nil
	}
	/*sysParams["app_auth_token"] = user.Token*/
	sysParams["sign"] = base64Encode(AliPaySigner(sysParams))
	//请求并除错
	requestParams := aliPayUrl + "?" + httpBuildQuery(sysParams)
	var result interface{}
	var next int
	var err error
	var needCancel, paySucc bool
	var trade TradeResult
	for getNowSec()-request.r.time < 30 {
		result, next, err = A.request(requestParams, "alipay_trade_pay_response")
		resp.SourceData = string(jsonEncode(result))
		if err != nil {
			if v, ok := aliErrMap[err.Error()]; ok {
				resp.Code = v
			} else {
				resp.Code = PayErr
			}
			if next == 1 {
				//重新授权后重试
				//重刷Token
				tokenResp := UserResult{}
				if A.RefreshToken(&RefreshToken{Type: "refresh_token", UserId: request.UserId, r: request.r}, &tokenResp); tokenResp.Code == 1000 {
					//重刷token后需要重新组装请求数据
					return A.BarCodePay(request, resp)
				} else {
					//重刷失败返回
					resp.Code = AuthErr
					break
				}
			} else if next == 2 {
				//系统支付异常
				A.TradeInfo(&TradeRequest{OutTradeId: request.OutTradeId, r: request.r}, &trade)
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
					A.TradeInfo(&TradeRequest{OutTradeId: request.OutTradeId, r: request.r}, &trade)
					if trade.Code == 1000 && trade.Data.Status == 1 {
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
		tmpresult := result.(map[string]interface{})
		amount, _ := strconv.ParseFloat(tmpresult["total_amount"].(string), 64)
		resp.Data = Trade{
			Id:         randomTimeString(),
			Amount:     int64(amount * 100),
			OutTradeId: request.OutTradeId,
			Source:     PAYTYPE_ALIPAY,
			PayTime:    request.r.time,
			UpTime:     request.r.time,
			Type:       1,
			Status:     TradeStatusSucc,
			TradeId:    tmpresult["trade_no"].(string),
		}
		saveTrade(resp.Data)
	}
	//撤销
	if needCancel {
		response := Response{}
		A.Cancel(&TradeRequest{OutTradeId: request.OutTradeId, r: request.r}, &response)
	}
	return nil
}

// 交易退款
// DOC:https://docs.open.alipay.com/api_1/alipay.trade.refund
func (A *AliPay) Refund(request *RefundRequest, resp *TradeResult) error {
	if request.r.time == 0 {
		request.r.time = getNowSec()
	}
	user := getUser(request.UserId, "alipay")
	if user.UserId == "" {
		resp.Code = AuthErr
		return nil
	}
	params := map[string]interface{}{
		"out_trade_no": request.OutTradeId,
		"trade_no":     request.TradeId,
		/*"out_request_no": request.RefundId,*/
		"refund_reason": request.Memo,
		"refund_amount": float64(request.Amount) / 100.0,
	}
	sysParams := A.sysParams()
	sysParams["method"] = "alipay.trade.refund"
	sysParams["biz_content"] = string(jsonEncode(params))
	sysParams["sign"] = base64Encode(AliPaySigner(sysParams))
	//请求并除错
	requestParams := aliPayUrl + "?" + httpBuildQuery(sysParams)
	var result interface{}
	var next int
	var err error
	for getNowSec()-request.r.time < 30 {
		result, next, err = A.request(requestParams, "alipay_trade_refund_response")
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
			tmpresult := result.(map[string]interface{})
			amount, _ := strconv.ParseFloat(tmpresult["refund_fee"].(string), 64)
			resp.Data = Trade{
				Id:         randomTimeString(),
				Amount:     int64(amount * 100),
				OutTradeId: request.OutTradeId,
				Source:     PAYTYPE_ALIPAY,
				Type:       -1,
				PayTime:    request.r.time,
				UpTime:     request.r.time,
				Memo:       request.Memo,
				Status:     TradeStatusSucc,
				TradeId:    tmpresult["trade_no"].(string),
			}
			saveTrade(resp.Data)
			return nil
		}
	}
	return nil
}

// 交易撤销
// DOC:https://docs.open.alipay.com/api_1/alipay.trade.cancel
// 入参 TradeRequest
// 出参 Response
func (A *AliPay) Cancel(request *TradeRequest, resp *Response) error {
	if request.r.time == 0 {
		request.r.time = getNowSec()
	}
	user := getUser(request.UserId, "alipay")
	if user.UserId == "" {
		resp.Code = AuthErr
		return nil
	}
	params := map[string]interface{}{
		"out_trade_no": request.OutTradeId,
		"trade_no":     request.TradeId,
	}
	sysParams := A.sysParams()
	sysParams["method"] = "alipay.trade.cancel"
	sysParams["biz_content"] = string(jsonEncode(params))
	sysParams["sign"] = base64Encode(AliPaySigner(sysParams))
	//请求并除错
	requestParams := aliPayUrl + "?" + httpBuildQuery(sysParams)
	var result interface{}
	var next int
	var err error
	for getNowSec()-request.r.time < 30 {
		result, next, err = A.request(requestParams, "alipay_trade_cancel_response")
		resp.SourceData = string(jsonEncode(result))
		if err != nil {
			if v, ok := aliErrMap[err.Error()]; ok {
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
	}
	return nil
}

// 获取支付单详情
// DOC:https://docs.open.alipay.com/api_1/alipay.trade.query/
// 传入参数TradeRequest
// 返回参数TradeResult
func (A *AliPay) TradeInfo(request *TradeRequest, resp *TradeResult) error {
	if request.r.time == 0 {
		request.r.time = getNowSec()
	}
	user := getUser(request.UserId, "alipay")
	if user.UserId == "" {
		resp.Code = AuthErr
		return nil
	}
	params := map[string]interface{}{
		"out_trade_no": request.OutTradeId,
		"trade_no":     request.TradeId,
	}
	sysParams := A.sysParams()
	sysParams["method"] = "alipay.trade.query"
	sysParams["biz_content"] = string(jsonEncode(params))
	sysParams["sign"] = base64Encode(AliPaySigner(sysParams))
	//请求并除错
	requestParams := aliPayUrl + "?" + httpBuildQuery(sysParams)
	var result interface{}
	var next int
	var err error
	for getNowSec()-request.r.time < 30 {
		result, next, err = A.request(requestParams, "alipay_trade_query_response")
		resp.SourceData = string(jsonEncode(result))
		if err != nil {
			if v, ok := aliErrMap[err.Error()]; ok {
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
			tmpresult := result.(map[string]interface{})
			amount, _ := strconv.ParseFloat(tmpresult["total_amount"].(string), 64)
			resp.Data = Trade{
				OutTradeId: tmpresult["out_trade_id"].(string),
				TradeId:    tmpresult["trade_id"].(string),
				Status:     Status(aliTradeStatusMap[tmpresult["status"].(string)]),
				Amount:     int64(amount * 100),
			}
			return nil
		}
	}
	return nil
}

// 刷新/获取授权token
// DOC:https://docs.open.alipay.com/api_9/alipay.open.auth.token.app
// 传入参数为RefreshToken格式
// 返回为 UserResult
func (A *AliPay) RefreshToken(request *RefreshToken, resp *UserResult) error {
	if request.r.time == 0 {
		request.r.time = getNowSec()
	}
	var user User
	params := map[string]interface{}{}
	if request.Type == "refresh" {
		user = getUser(request.UserId, "alipay")
		if user.UserId == "" {
			resp.Code = AuthErr
			return nil
		}
		params["grant_type"] = "refresh_token"
		params["refresh_token"] = user.ReToken
	} else {
		params["grant_type"] = "authorization_code"
		params["code"] = request.Code
	}
	sysParams := A.sysParams()
	sysParams["method"] = "alipay.open.auth.token.app"
	sysParams["biz_content"] = string(jsonEncode(params))
	sysParams["sign"] = base64Encode(AliPaySigner(sysParams))
	//请求并除错
	requestParams := aliPayUrl + "?" + httpBuildQuery(sysParams)
	var result interface{}
	var next int
	var err error
	for getNowSec()-request.r.time < 30 {
		result, next, err = A.request(requestParams, "alipay_open_auth_token_app_response")
		resp.SourceData = string(jsonEncode(result))
		if err != nil {
			if v, ok := aliErrMap[err.Error()]; ok {
				resp.Code = v
			} else {
				resp.Code = AuthErr
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
			tmpresult := result.(map[string]interface{})
			if user.UserId == "" {
				user = getUser(tmpresult["user_id"].(string), "alipay")
			}
			user.Token = tmpresult["app_auth_token"].(string)
			user.ReToken = tmpresult["app_refresh_token"].(string)
			user.ExAt = request.r.time + int64(tmpresult["expires_in"].(float64))
			resp.Data = user
			//保存用户授权
			if user.UserId != "" {
				updateUser(user.UserId, user.Source, bson.M{"$set": user})
			} else {
				user.UserId = tmpresult["user_id"].(string)
				user.Source = "alipay"
				saveUser(user)
			}
			return nil
		}
	}
	return nil
}

func (A *AliPay) request(url string, okey string) (interface{}, int, error) {
	fmt.Println(url)
	body, err := getRequest(url)
	if err != nil {
		//网络发起请求失败
		//需重试
		return nil, -1, err
	}
	fmt.Println(string(body))
	result := map[string]interface{}{}
	if err := jsonDecode(body, &result); err != nil {
		return nil, 0, err
	}
	data, ok := result[okey]
	if !ok {
		return nil, 0, newError("支付宝返回数据中不存在" + okey)
	}
	datamap, ok := data.(map[string]interface{})
	if !ok {
		return nil, 0, newError("支付宝返回数据中" + okey + "参数格式错误")
	}
	next, err := A.errorCheck(datamap)
	return datamap, next, err
}

func (A *AliPay) errorCheck(data map[string]interface{}) (int, error) {
	code := data["code"].(string)
	subCode := ""
	if v, ok := data["sub_code"]; ok {
		subCode = v.(string)
	}
	switch code {
	case "10000":
		//成功
		return 0, nil
	case "20001":
		//重新授权，后在重试
		return 1, newError(code + subCode)
	case "20000":
		//立即重试
		return 2, newError(code + subCode)
	case "10003":
		//循环重试
		return 3, newError(code + subCode)
	default:
		return 0, newError(code + subCode)
	}
}

/*组装系统级请求参数*/
func (a *AliPay) sysParams() map[string]string {
	return map[string]string{
		"app_id":    aliPayAppId,
		"format":    aliPayDefaultFormat,
		"charset":   aliPayDefaultCharset,
		"sign_type": aliPayDefaultSignType,
		"version":   "1.0",
		"timestamp": sec2Str("2006-01-02 15:04:05", getNowSec()),
	}
}

var aliErrMap = map[string]int{
	"40004ACQ.PAYMENT_AUTH_CODE_INVALID":  PayErrCode,
	"40004ACQ.TRADE_HAS_SUCCESS":          PayErrPayed,
	"40004ACQ.TRADE_NOT_EXIST":            TradeErrNotFound,
	"40004ACQ.TRADE_STATUS_ERROR":         TradeErrStatus,
	"40004ACQ.SELLER_BALANCE_NOT_ENOUGH":  RefundErrBalance,
	"40004ACQ.REFUND_AMT_NOT_EQUAL_TOTAL": RefundErrAmount,
}
var aliTradeStatusMap = map[string]Status{
	"WAIT_BUYER_PAY": TradeStatusWaitPay,
	"TRADE_CLOSED":   TradeStatusClose,
	"TRADE_SUCCESS":  TradeStatusSucc,
	"TRADE_FINISHED": TradeStatusSucc,
}
