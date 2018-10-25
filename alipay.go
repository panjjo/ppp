package ppp

import (
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
	aliPayNotifyUrl         string           //异步通知地址
)

const (
	FC_ALIPAY_BARCODEPAY     string = "AliPay.BarCodePay" //支付宝条码支付
	FC_ALIPAY_CANCEL         string = "AliPay.Cancel"     //支付宝取消交易
	FC_ALIPAY_AUTH           string = "AliPay.Auth"       //支付宝授权
	FC_ALIPAY_REFUND         string = "AliPay.Refund"     //支付宝退款
	FC_ALIPAY_TRADEINFO      string = "AliPay.TradeInfo"  //支付宝订单详情
	FC_ALIPAY_AUTHSIGNED     string = "AliPay.AuthSigned" //签约接口
	FC_ALIPAY_WAPTRADEPARAMS string = "AliPay.PayParams"  //支付参数组装
)

type AliPayInit struct {
	AppId             string
	Url               string
	ServiceProviderId string
	ConfigPath        string
	NotifyUrl         string
}

func (a *AliPayInit) Init() {
	aliPayUrl = a.Url
	aliPayAppId = a.AppId
	aliPayServiceProviderId = a.ServiceProviderId
	aliPayNotifyUrl = a.NotifyUrl
	loadAliPayCertKey(a.ConfigPath)
}

//支付宝接口主体
type AliPay struct {
}

//支付宝签约通过后调用
//支付宝做更新签约状态，签约支付宝账号
func (A *AliPay) AuthSigned(request *AuthRequest, resp *Response) error {
	Log.DEBUG.Printf("AliPay api:AuthSigned,request:%+v", request)
	defer Log.DEBUG.Printf("AliPay api:AuthSigned,response:%+v", resp)
	auth := getToken(request.MchId, PAYTYPE_ALIPAY)
	if auth.Id == "" {
		resp.Code = AuthErr
		return nil
	}
	updateToken(auth.MchId, PAYTYPE_ALIPAY, bson.M{"$set": bson.M{"status": AuthStatusSucc, "account": request.Account}})
	updateUserMulti(bson.M{"mchid": auth.MchId, "type": PAYTYPE_ALIPAY, "status": bson.M{"$ne": UserFreeze}}, bson.M{"$set": bson.M{"status": UserSucc}})
	//验证权限是否真实开通
	trade := TradeResult{}
	A.TradeInfo(&TradeRequest{r: rsys{mchid: auth.MchId}, TradeId: "test123"}, &trade)
	if trade.Code == AuthErr {
		//撤销
		updateToken(auth.MchId, PAYTYPE_ALIPAY, bson.M{"$set": bson.M{"status": AuthStatusWaitSigned, "account": request.Account}})
		updateUserMulti(bson.M{"mchid": auth.MchId, "type": PAYTYPE_ALIPAY, "status": bson.M{"$ne": UserFreeze}}, bson.M{"$set": bson.M{"status": UserWaitVerify}})
		resp.Code = AuthErr
	}
	return nil
}

// 统一收单支付接口
// DOC:https://docs.open.alipay.com/api_1/alipay.trade.pay
// 传入参数为 BarCodePayRequest格式
// 返回参数为 TradeResult
// userid 为收款方自定义id,应存在签约授权成功后保存的对应关系
func (A *AliPay) BarCodePay(request *BarCodePayRequest, resp *TradeResult) error {
	Log.DEBUG.Printf("AliPay api:BarCodePay,request:%+v", request)
	defer Log.DEBUG.Printf("AliPay api:BarCodePay,response:%+v", resp)
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
	auth := A.token(request.UserId, request.r.mchid)
	if auth.Status != AuthStatusSucc {
		resp.Code = AuthErrNotSigned
		return nil
	}
	request.r.mchid = auth.MchId
	sysParams["app_auth_token"] = auth.Token
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
				tokenResp := AuthResult{}
				if A.Auth(&Token{refresh: true, Code: auth.ReToken, r: request.r}, &tokenResp); tokenResp.Code == 1000 {
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
				} else if trade.Data.Status == TradeStatusSucc {
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
					if trade.Code == 0 && trade.Data.Status == TradeStatusSucc {
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
		resp.Code = Succ
		saveTrade(resp.Data)
	}
	//撤销
	if needCancel {
		response := Response{}
		A.Cancel(&TradeRequest{OutTradeId: request.OutTradeId}, &response)
	}
	return nil
}

// 交易退款
// DOC:https://docs.open.alipay.com/api_1/alipay.trade.refund
func (A *AliPay) Refund(request *RefundRequest, resp *TradeResult) error {
	Log.DEBUG.Printf("AliPay api:Refund,request:%+v", request)
	defer Log.DEBUG.Printf("AliPay api:Refund,response:%+v", resp)
	if request.r.time == 0 {
		request.r.time = getNowSec()
	}
	//获取授权
	auth := A.token(request.UserId, request.r.mchid)
	if auth.Status != AuthStatusSucc {
		resp.Code = AuthErrNotSigned
		return nil
	}
	request.r.mchid = auth.MchId
	trade := TradeResult{}
	A.TradeInfo(&TradeRequest{r: request.r, OutTradeId: request.OutTradeId}, &trade)
	if trade.Code != 0 {
		resp.Code = trade.Code
		resp.SourceData = trade.SourceData
		return nil
	}
	if trade.Data.Id == "" {
		resp.Code = TradeErrNotFound
		return nil
	}
	params := map[string]interface{}{
		"out_trade_no":   request.OutTradeId,
		"trade_no":       request.TradeId,
		"out_request_no": request.OutRefundId,
		"refund_reason":  request.Memo,
		"refund_amount":  float64(request.Amount) / 100.0,
	}
	sysParams := A.sysParams()
	sysParams["method"] = "alipay.trade.refund"
	sysParams["app_auth_token"] = auth.Token
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
				OutTradeId: request.RefundId,
				Source:     PAYTYPE_ALIPAY,
				Type:       -1,
				PayTime:    request.r.time,
				UpTime:     request.r.time,
				Memo:       request.Memo,
				Status:     TradeStatusSucc,
				TradeId:    tmpresult["trade_no"].(string),
				ParentId:   trade.Data.Id,
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
	Log.DEBUG.Printf("AliPay api:Cancel,request:%+v", request)
	defer Log.DEBUG.Printf("AliPay api:Cancel,response:%+v", resp)
	if request.r.time == 0 {
		request.r.time = getNowSec()
	}
	//获取授权
	auth := A.token(request.UserId, request.r.mchid)
	if auth.Status != AuthStatusSucc {
		resp.Code = AuthErrNotSigned
		return nil
	}
	request.r.mchid = auth.MchId
	params := map[string]interface{}{
		"out_trade_no": request.OutTradeId,
		"trade_no":     request.TradeId,
	}
	sysParams := A.sysParams()
	sysParams["method"] = "alipay.trade.cancel"
	sysParams["biz_content"] = string(jsonEncode(params))
	sysParams["app_auth_token"] = auth.Token
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
	Log.DEBUG.Printf("AliPay api:TradeInfo,request:%+v", request)
	defer Log.DEBUG.Printf("AliPay api:TradeInfo,response:%+v", resp)
	if request.r.time == 0 {
		request.r.time = getNowSec()
	}
	//获取授权
	auth := A.token(request.UserId, request.r.mchid)
	if auth.Status != AuthStatusSucc {
		resp.Code = AuthErrNotSigned
		return nil
	}
	request.r.mchid = auth.MchId
	q := bson.M{"source": PAYTYPE_ALIPAY}
	if request.OutTradeId != "" {
		q["outtradeid"] = request.OutTradeId
	}
	if request.TradeId != "" {
		q["tradeid"] = request.TradeId
	}
	trade := getTrade(q)
	if request.DisSync {
		if trade.Id == "" {
			resp.Code = TradeErrNotFound
		}
		resp.Data = trade
		return nil
	}
	params := map[string]interface{}{
		"out_trade_no": request.OutTradeId,
		"trade_no":     request.TradeId,
	}
	sysParams := A.sysParams()
	sysParams["method"] = "alipay.trade.query"
	sysParams["biz_content"] = string(jsonEncode(params))
	sysParams["app_auth_token"] = auth.Token
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
				OutTradeId: tmpresult["out_trade_no"].(string),
				TradeId:    tmpresult["trade_no"].(string),
				Status:     aliTradeStatusMap[tmpresult["trade_status"].(string)],
				Amount:     int64(amount * 100),
				Id:         trade.Id,
				UpTime:     getNowSec(),
			}
			if paytime, ok := tmpresult["send_pay_date"]; ok {
				resp.Data.PayTime = str2Sec("2006-01-02 15:04:05", paytime.(string))
			}
			//更新数据
			if trade.Id != "" {
				updateTrade(bson.M{"id": trade.Id}, bson.M{"$set": bson.M{"status": resp.Data.Status, "uptime": getNowSec(), "paytime": resp.Data.PayTime}})
			}
			return nil
		}
	}
	return nil
}

// 刷新/获取授权token
// DOC:https://docs.open.alipay.com/api_9/alipay.open.auth.token.app
// 传入参数为Token格式
// 返回为 AuthResult
// 如果刷新获取token后返回的第三方授权已经存在会更新，不存在新生授权
func (A *AliPay) Auth(request *Token, resp *AuthResult) error {
	Log.DEBUG.Printf("AliPay api:Auth,request:%+v", request)
	defer Log.DEBUG.Printf("AliPay api:Auth,response:%+v", resp)
	if request.r.time == 0 {
		request.r.time = getNowSec()
	}
	params := map[string]interface{}{}
	if request.refresh {
		params["grant_type"] = "authorization_code"
		params["refresh_token"] = request.Code
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
		if err != nil {
			resp.SourceData = string(jsonEncode(result))
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
			mchid := tmpresult["user_id"].(string)
			auth := getToken(mchid, PAYTYPE_ALIPAY)
			auth.Token = tmpresult["app_auth_token"].(string)
			auth.ReToken = tmpresult["app_refresh_token"].(string)
			auth.ExAt = request.r.time + int64(tmpresult["expires_in"].(float64))
			//保存用户授权
			if auth.MchId != "" {
				updateToken(auth.MchId, PAYTYPE_ALIPAY, bson.M{"$set": auth})
			} else {
				auth.Id = randomString(15)
				auth.MchId = mchid
				auth.Type = PAYTYPE_ALIPAY
				saveToken(auth)
			}
			resp.Data = Auth{
				MchId:  auth.MchId,
				Id:     auth.Id,
				Type:   auth.Type,
				Status: auth.Status,
			}
			return nil
		}
	}
	return nil
}

//使用应用自授权，非子商户模式
//DOC:https://docs.open.alipay.com/203/107090/
//本接口只负责数据组装，发起请求应由对应客户端发起
func (A *AliPay) PayParams(request *WapPayRequest, resp *Response) error {
	Log.DEBUG.Printf("AliPay api:WapPayParams,request:%+v", request)
	defer Log.DEBUG.Printf("AliPay api:WapPayParams,response:%+v", resp)
	var productCode, method string
	switch request.TradeType {
	case WAPPAYPARAMS:
		productCode = "FAST_INSTANT_TRADE_PAY"
		method = "alipay.trade.page.pay"
	case APPPAYPARAMS:
		productCode = "QUICK_MSECURITY_PAY"
		method = "alipay.trade.app.pay"
	default:
		productCode = "FAST_INSTANT_TRADE_PAY"
		method = "alipay.trade.page.pay"
	}
	params := map[string]interface{}{
		"body":            request.ItemDes,
		"subject":         request.TradeName,
		"out_trade_no":    request.OutTradeId,
		"total_amount":    float64(request.Amount) / 100.0,
		"product_code":    productCode,
		"store_id":        request.ShopId,
		"passback_params": request.Ex,
	}
	sysParams := A.sysParams()
	sysParams["method"] = method
	sysParams["biz_content"] = string(jsonEncode(params))
	sysParams["return_url"] = request.ReturnUrl
	sysParams["notify_url"] = aliPayNotifyUrl
	sysParams["sign"] = base64Encode(AliPaySigner(sysParams))
	resp.SourceData = httpBuildQuery(sysParams)
	// resp.SourceData = string(jsonEncode(sysParams))

	//save tradeinfo
	saveTrade(Trade{
		OutTradeId: request.OutTradeId,
		Status:     0,
		Type:       1,
		Amount:     request.Amount,
		Source:     PAYTYPE_ALIPAY,
		UpTime:     getNowSec(),
		Ex:         request.Ex,
		Id:         randomTimeString(), // PPPID
	})
	return nil
}

//异步回调
//wap支付的异步回调
//request 接收到的支付宝回调所有参数
func (A *AliPay) CallBack(request map[string]string, resp *Response) error {
	Log.DEBUG.Printf("AliPay api:CallBack,request:%+v", request)
	defer Log.DEBUG.Printf("AliPay api:CallBack,response:%+v", resp)
	sign, ok := request["sign"]
	if !ok {
		resp.Code = SysErrParams
		return nil
	}
	signType, ok := request["sign_type"]
	if !ok || signType != "RSA2" {
		resp.Code = SysErrParams
		return nil
	}
	delete(request, "sign")
	delete(request, "sign_type")
	err := AliPayRSAVerify(request, sign)
	if err != nil {
		resp.Code = SysErrVerify
		resp.SourceData = err.Error()
		return nil
	}
	//TODO:更新trade
	return nil
}

func (A *AliPay) request(url string, okey string) (interface{}, int, error) {
	body, err := getRequest(url)
	if err != nil {
		//网络发起请求失败
		//需重试
		return nil, -1, err
	}
	result := map[string]interface{}{}
	Log.DEBUG.Printf("alipay request url:%s,body:%s", url, string(body))
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

func (A *AliPay) token(userid, mchid string) authBase {
	auth := authBase{}
	if mchid == "" {
		user := getUser(userid, PAYTYPE_ALIPAY)
		if user.Status != UserSucc {
			return auth
		}
		mchid = user.MchId
	}
	return getToken(mchid, PAYTYPE_ALIPAY)
}

/*组装系统级请求参数*/
func (A *AliPay) sysParams() map[string]string {
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
	"40004ACQ.SELLER_BALANCE_NOT_ENOUGH":  UserErrBalance,
	"40004ACQ.REFUND_AMT_NOT_EQUAL_TOTAL": RefundErrAmount,
}
var aliTradeStatusMap = map[string]Status{
	"WAIT_BUYER_PAY": TradeStatusWaitPay,
	"TRADE_CLOSED":   TradeStatusClose,
	"TRADE_SUCCESS":  TradeStatusSucc,
	"TRADE_FINISHED": TradeStatusSucc,
}
