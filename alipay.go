package ppp

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	proto "github.com/panjjo/ppp/proto"
	"github.com/sirupsen/logrus"
	"net/http"
	"path/filepath"
	"time"
)

const (
	aliPayDefaultFormat   = "JSON"
	aliPayDefaultCharset  = "utf-8"
	aliPayDefaultSignType = "RSA2"
)

var alipay *AliPay

type AliPay struct {
	cfgs map[string]config
	def  config
}

func aliConfig(config ConfigSingle) (a config) {
	if config.AppID != "" {
		a.appid = config.AppID
	} else {
		logrus.Fatalf("not found alipay appID")
	}
	if config.URL != "" {
		a.url = config.URL
	} else {
		logrus.Fatalf("not found alipay apiurl")
	}
	a.notify = config.Notify

	if config.ServiceID != "" {
		a.serviceid = config.ServiceID
	}
	// 加载应用私钥证书
	private, err := LoadPrivateKeyFromFile(filepath.Join(config.CertPath, "private.key"))
	if err != nil {
		logrus.Fatalf("load alipay privateCert fail,file:%s,err:%s", config.CertPath, err)
	}
	a.private = private
	// 加载支付宝公钥
	public, err := LoadPublicKeyFromFile(filepath.Join(config.CertPath, "public.key"))
	if err != nil {
		logrus.Fatalf("load alipay publicCert fail,file:%s,err:%s", config.CertPath, err)
	}
	a.public = public
	return a
}

// NewAliPay 获取支付宝实例
func NewAliPay(cfgs Config) *AliPay {
	alipay = &AliPay{cfgs: map[string]config{}}
	if cfgs.AppID != "" {
		cfgs.Apps = append([]ConfigSingle{cfgs.ConfigSingle}, cfgs.Apps...)
	}
	for _, cfg := range cfgs.Apps {
		c := aliConfig(cfg)
		if alipay.def.appid == "" {
			alipay.def = c
		}
		if cfg.Tag != "" {
			alipay.cfgs[cfg.Tag] = c
		} else {
			alipay.cfgs[cfg.AppID] = c
		}
	}
	logrus.Debugf("alipay cfgs:%+v,def:%+v", alipay.cfgs, alipay.def)
	return alipay
}

// MchPay 企业付款 到 个人支付宝账号
// 单商户模式调用
// UserName 不为空则强制检测真实姓名和支付宝账号是否一致
func (A *AliPay) MchPay(ctx *Context, req *proto.Transfer) (resp *proto.Transfer, e Error) {
	params := map[string]interface{}{
		"out_biz_no":      req.OutTradeID,
		"payee_account":   req.Account,
		"amount":          fmt.Sprintf("%.2f", float64(req.Amount)/100.0),
		"payee_real_name": req.UserName,
		"remark":          req.Desc,
	}
	switch req.AccountType {
	case proto.Accounttype_id:
		params["payee_type"] = "ALIPAY_USERID"
	case proto.Accounttype_login:
		params["payee_type"] = "ALIPAY_LOGONID"
	default:
		params["payee_type"] = "ALIPAY_LOGONID"
	}
	// 组装系统参数
	sysParams := A.sysParams(ctx)
	sysParams["method"] = "alipay.fund.trans.toaccount.transfer"
	sysParams["biz_content"] = string(jsonEncode(params))
	sysParams["sign"] = base64Encode(A.Signer(ctx, sysParams))
	rq := requestSimple{url: ctx.url(), get: httpBuildQuery(sysParams), relKey: "alipay_fund_trans_toaccount_transfer_response", ctx: ctx}
	rq.fs = func(result interface{}, next Status, err error) (interface{}, error) {
		switch next {
		case netConnErr, nextRetry:
			// 超时，异常立刻重试
			time.Sleep(1 * time.Second)
			return A.Request(rq)
		default:
			return result, err
		}
	}
	info, err := A.Request(rq)
	if err != nil {
		e.Msg = err.Error()
		if v, ok := aliErrMap[err.Error()]; ok {
			e.Code = v
		} else {
			e.Code = PayErr
		}
	} else {
		// 转账成功
		tmpresult := info.(map[string]interface{})
		resp = req
		resp.TradeID = tmpresult["order_Id"].(string)
	}
	return
}

// PayParams 获取支付参数
// 用于前段请求，不想暴露证书的私密信息的可用此方法组装请求参数，前端只负责请求
// 支持的有 网站支付，手机app支付，h5支付等
// 仅支持单商户模式，不支持服务商模式
// 默认使用的服务商对应账号开通的子商户收款（服务商模式下的分润获得的子商户）
func (A *AliPay) PayParams(ctx *Context, req *proto.Params) (data *proto.ParamsData, e Error) {
	trade := getTrade(map[string]interface{}{"OutTradeID": req.OutTradeID})
	if trade.ID != "" && trade.Status == proto.Tradestatus_succ {
		// 检测订单号是否存在 并且支付成功
		e.Code = TradeErrStatus
		e.Msg = "订单已支付"
		return
	}
	var productCode, method string
	switch req.Type {
	case proto.Tradetype_WEB:
		productCode = "FAST_INSTANT_TRADE_PAY"
		method = "alipay.trade.page.pay"
	case proto.Tradetype_APP:
		productCode = "QUICK_MSECURITY_PAY"
		method = "alipay.trade.app.pay"
	default:
		productCode = "FAST_INSTANT_TRADE_PAY"
		method = "alipay.trade.page.pay"
	}
	params := map[string]interface{}{
		"body":            req.ItemDes,
		"subject":         req.TradeName,
		"out_trade_no":    req.OutTradeID,
		"total_amount":    float64(req.Amount) / 100.0,
		"product_code":    productCode,
		"store_id":        req.ShopID,
		"passback_params": req.EX,
	}
	if req.NotifyURL == "" {
		req.NotifyURL = ctx.Notify()
	}
	sysParams := A.sysParams(ctx)
	sysParams["method"] = method
	sysParams["biz_content"] = string(jsonEncode(params))
	sysParams["return_url"] = req.ReturnURL
	sysParams["notify_url"] = req.NotifyURL
	sysParams["sign"] = base64Encode(A.Signer(ctx, sysParams))
	data = &proto.ParamsData{
		Params:     httpBuildQuery(sysParams),
		SourceData: string(jsonEncode(sysParams)),
	}
	logrus.Println(data.Params)
	logrus.Println(data.SourceData)
	newTrade := &proto.Trade{
		OutTradeID: req.OutTradeID,
		Amount:     req.Amount,
		ID:         randomTimeString(),
		Type:       req.Type,
		MchID:      ctx.serviceid(),
		UpTime:     getNowSec(),
		Create:     getNowSec(),
		AppID:      req.AppID,
		From:       proto.ALIPAY,
	}
	// save tradeinfo
	if trade.ID != "" {
		// 更新
		updateTrade(map[string]interface{}{"OutTradeID": trade.OutTradeID}, newTrade)
	} else {
		// 新增
		saveTrade(newTrade)
	}
	return
}

// BarPay 商户主动扫码支付
// 同一个OutTradeID 不能重复支付
// 支持服务商模式，单商户模式
func (A *AliPay) BarPay(ctx *Context, req *proto.Barpay) (trade *proto.Trade, e Error) {
	// 获取授权
	auth := ctx.getAuth(req.UserID, req.MchID)
	if auth.Status != proto.Accountstatus_ok {
		// 授权错误
		e.Code = AuthErr
		return
	}

	trade = getTrade(map[string]interface{}{"OutTradeID": req.OutTradeID})
	if trade.ID != "" && trade.Status == proto.Tradestatus_succ {
		// 如果订单已经存在并且支付，返回报错
		e.Code = PayErrPayed
		return
	}
	params := map[string]interface{}{
		"out_trade_no": req.OutTradeID,
		"scene":        "bar_code",
		"auth_code":    req.AuthCode,
		"subject":      req.TradeName,
		"total_amount": float64(req.Amount) / 100.0,
		"body":         req.ItemDes,
		"store_id":     req.ShopID,
	}
	// 设置反佣系统商编号
	if ctx.serviceid() != "" {
		params["extend_params"] = map[string]interface{}{"sys_service_provIder_Id": ctx.serviceid()}
	}
	// 组装系统参数
	sysParams := A.sysParams(ctx)
	sysParams["method"] = "alipay.trade.pay"
	sysParams["biz_content"] = string(jsonEncode(params))
	// 设置子商户数据
	if ctx.mchid() != ctx.serviceid() {
		sysParams["app_auth_token"] = ctx.token()
	}
	sysParams["sign"] = base64Encode(A.Signer(ctx, sysParams))
	// 订单是否需要撤销，支付是否成功
	var needCancel, paySucc bool
	rq := requestSimple{url: ctx.url(), get: httpBuildQuery(sysParams), relKey: "alipay_trade_pay_response", ctx: ctx}
	rq.fs = func(result interface{}, next Status, err error) (interface{}, error) {
		switch next {
		case netConnErr:
			// 网络错误
			time.Sleep(1 * time.Second)
			return A.Request(rq)
		case nextRetry:
			// 支付异常 https://docs.open.alipay.com/194/105322/
			// 查询订单，如果支付失败，则取消订单
			trade, e = A.TradeInfo(ctx, &proto.Trade{OutTradeID: req.OutTradeID})
			if e.Code == TradeErrNotFound {
				// 订单不存在 相同参数再次支付
				return A.Request(rq)
			} else if trade.Status == proto.Tradestatus_succ {
				// 订单支付成功
				paySucc = true
			} else {
				// 其他错误，取消订单
				needCancel = true
				e.Code = PayErr
				e.Msg = e.Msg + "订单支付失败"
			}
		case nextWaitRetry:
			needCancel = true
			// 等待用户输入密码
			// 每3秒获取一次订单信息，直至支付超时或支付成功
			for getNowSec()-ctx.gt() < maxTimeout {
				time.Sleep(3 * time.Second)
				trade, e = A.TradeInfo(ctx, &proto.Trade{OutTradeID: req.OutTradeID})
				if e.Code == 0 && trade.Status == proto.Tradestatus_succ {
					// 支付成功
					paySucc = true
					needCancel = false
					return trade, nil
				} else {
					paySucc = false
					needCancel = true
					e.Code = PayErr
					e.Msg = e.Msg + "订单未支付"
				}
			}
		default:
			needCancel = true
			e.Code = PayErr
			e.Msg = err.Error()
		}
		return trade, newErrorByE(e)
	}
	info, err := A.Request(rq)
	if err != nil {
		e.Msg = err.Error()
		if v, ok := aliErrMap[err.Error()]; ok {
			e.Code = v
		} else {
			e.Code = PayErr
		}
	} else {
		// 返回成功
		paySucc = true
		needCancel = false
	}
	if paySucc {
		result := trade
		switch info.(type) {
		case *proto.Trade:
			tmpresult := info.(*proto.Trade)
			result.TradeID = tmpresult.TradeID
		case map[string]interface{}:
			tmpresult := info.(map[string]interface{})
			result.TradeID = tmpresult["trade_no"].(string)
		}
		result.Amount = req.Amount
		result.From = proto.ALIPAY
		result.UserID = ctx.userid()
		result.MchID = ctx.mchid()
		result.UpTime = ctx.gt()
		result.PayTime = ctx.gt()
		result.AppID = ctx.appid()
		result.Status = proto.Tradestatus_succ
		result.Type = proto.Tradetype_BAR
		if result.ID == "" {
			result.OutTradeID = req.OutTradeID
			result.ID = randomTimeString()
			result.Create = ctx.gt()
			// 保存订单
			saveTrade(result)
		} else {
			// 更新订单
			updateTrade(map[string]interface{}{"Id": trade.ID}, result)
		}
	}
	if needCancel {
		// 取消订单
		ctx.t = getNowSec()
		A.Cancel(ctx, &proto.Trade{OutTradeID: req.OutTradeID})
	}
	return
}

// Refund 退款
// req sourceId使用交易里面对应的OutTradeID
// 仅支持使用ppp支付的订单退款
// 支持服务商模式，单商户模式
func (A *AliPay) Refund(ctx *Context, req *proto.Refund) (refund *proto.Refund, e Error) {
	// 获取授权
	auth := ctx.getAuth(req.UserID, req.MchID)
	if auth.Status != proto.Accountstatus_ok {
		// 授权错误
		e.Code = AuthErr
		return
	}
	trade, e := A.TradeInfo(ctx, &proto.Trade{OutTradeID: req.SourceID})
	if trade.ID == "" || e.Code == TradeErrNotFound {
		e.Code = TradeErrNotFound
		return
	}
	if trade.Status != proto.Tradestatus_succ && trade.Status != proto.Tradestatus_refunded {
		e.Code = TradeErrStatus
		return
	}
	params := map[string]interface{}{
		"out_trade_no":   req.SourceID,
		"out_request_no": req.OutRefundID,
		"refund_reason":  req.Memo,
		"refund_amount":  float64(req.Amount) / 100.0,
	}
	sysParams := A.sysParams(ctx)
	sysParams["method"] = "alipay.trade.refund"
	// 设置子商户数据
	if ctx.mchid() != ctx.serviceid() {
		sysParams["app_auth_token"] = ctx.token()
	}
	sysParams["biz_content"] = string(jsonEncode(params))
	sysParams["sign"] = base64Encode(A.Signer(ctx, sysParams))
	rq := requestSimple{url: ctx.url(), get: httpBuildQuery(sysParams), relKey: "alipay_trade_refund_response", ctx: ctx}
	rq.fs = func(result interface{}, next Status, err error) (interface{}, error) {
		switch next {
		case netConnErr, nextRetry:
			// 超时，异常立刻重试
			time.Sleep(1 * time.Second)
			return A.Request(rq)
		default:
			return result, err
		}
	}
	info, err := A.Request(rq)
	if err != nil {
		e.Msg = err.Error()
		if v, ok := aliErrMap[err.Error()]; ok {
			e.Code = v
		} else {
			e.Code = RefundErr
		}
	} else {
		// 退款成功
		result := info.(map[string]interface{})
		refund = &proto.Refund{
			RefundID:    result["trade_no"].(string),
			ID:          randomTimeString(),
			OutRefundID: req.OutRefundID,
			MchID:       ctx.mchid(),
			UserID:      ctx.userid(),
			Amount:      req.Amount,
			SourceID:    req.SourceID,
			Status:      proto.Tradestatus_refunded,
			UpTime:      ctx.gt(),
			RefundTime:  ctx.gt(),
			Create:      ctx.gt(),
			From:        proto.ALIPAY,
			AppID:       ctx.appid(),
			Memo:        req.Memo,
		}
		saveRefund(refund)
		// 退款成功更新订单状态
		trade.Status = proto.Tradestatus_refunded
		updateTrade(map[string]string{"Id": trade.ID}, trade)
	}
	return
}

// Cancel 取消订单
// 可用参数 req:tradeId OutTradeID MchID UserID
// 如果订单已支付会取消失败
// 支持服务商模式，单商户模式
func (A *AliPay) Cancel(ctx *Context, req *proto.Trade) (trade *proto.Trade, e Error) {
	trade, e = A.TradeInfo(ctx, &proto.Trade{OutTradeID: req.OutTradeID})
	// 获取授权
	auth := ctx.getAuth(req.UserID, req.MchID)
	if auth.Status != proto.Accountstatus_ok {
		// 授权错误
		e.Code = AuthErr
		return
	}
	params := map[string]interface{}{
		"out_trade_no": req.OutTradeID,
		"trade_no":     req.TradeID,
	}
	sysParams := A.sysParams(ctx)
	sysParams["method"] = "alipay.trade.cancel"
	sysParams["biz_content"] = string(jsonEncode(params))
	// 设置子商户数据
	if ctx.mchid() != ctx.serviceid() {
		sysParams["app_auth_token"] = ctx.token()
	}
	sysParams["sign"] = base64Encode(A.Signer(ctx, sysParams))
	rq := requestSimple{url: ctx.url(), get: httpBuildQuery(sysParams), relKey: "alipay_trade_cancel_response", ctx: ctx}
	rq.fs = func(result interface{}, next Status, err error) (interface{}, error) {
		switch next {
		case netConnErr, nextRetry:
			// 超时，异常立刻重试
			time.Sleep(1 * time.Second)
			return A.Request(rq)
		default:
			return result, err
		}
	}
	_, err := A.Request(rq)
	if err != nil {
		e.Msg = err.Error()
		if v, ok := aliErrMap[err.Error()]; ok {
			e.Code = v
		} else {
			e.Code = TradeErr
		}
	} else {
		// 撤销成功
		if trade.ID != "" {
			trade.Status = proto.Tradestatus_cancel
			updateTrade(map[string]string{"Id": trade.ID}, trade)
		}
	}
	return trade, e
}

// TradeInfo 获取订单详情
// 可用参数 req: tradeId OutTradeID MchID UserID
// 支持服务商模式，单商户模式
func (A *AliPay) TradeInfo(ctx *Context, req *proto.Trade) (trade *proto.Trade, e Error) {
	// 获取授权
	auth := ctx.getAuth(req.UserID, req.MchID)
	if auth.Status != proto.Accountstatus_ok {
		// 授权错误
		e.Code = AuthErr
		return
	}
	q := map[string]interface{}{"from": proto.ALIPAY}
	if req.OutTradeID != "" {
		q["OutTradeID"] = req.OutTradeID
	}
	if req.TradeID != "" {
		q["TradeID"] = req.TradeID
	}
	trade = getTrade(q)
	// 同步第三方数据
	params := map[string]interface{}{
		"out_trade_no": req.OutTradeID,
		"trade_no":     req.TradeID,
	}
	sysParams := A.sysParams(ctx)
	sysParams["method"] = "alipay.trade.query"
	sysParams["biz_content"] = string(jsonEncode(params))
	// 设置子商户数据
	if ctx.mchid() != ctx.serviceid() {
		sysParams["app_auth_token"] = ctx.token()
	}
	sysParams["sign"] = base64Encode(A.Signer(ctx, sysParams))
	rq := requestSimple{url: ctx.url(), get: httpBuildQuery(sysParams), relKey: "alipay_trade_query_response", ctx: ctx}
	rq.fs = func(result interface{}, next Status, err error) (interface{}, error) {
		switch next {
		case netConnErr, nextRetry:
			// 超时，异常立刻重试
			time.Sleep(1 * time.Second)
			return A.Request(rq)
		default:
			return result, err
		}
	}
	info, err := A.Request(rq)
	if err != nil {
		e.Msg = err.Error()
		if v, ok := aliErrMap[err.Error()]; ok {
			e.Code = v
		} else {
			e.Code = TradeErr
		}
	} else {
		// 成功返回
		tmpresult := info.(map[string]interface{})
		// 数据返回后以第三方返回数据为准
		trade = &proto.Trade{
			Amount:     round(parseFloat(tmpresult["total_amount"].(string)) * 100),
			Status:     aliTradeStatusMap[tmpresult["trade_status"].(string)],
			ID:         trade.ID,
			UpTime:     getNowSec(),
			OutTradeID: req.OutTradeID,
			TradeID:    tmpresult["trade_no"].(string),
			Create:     trade.Create,
			Type:       trade.Type,
			From:       proto.ALIPAY,
			AppID:      ctx.appid(),
		}
		trade.MchID = ctx.mchid()
		trade.UserID = ctx.userid()
		if paytime, ok := tmpresult["send_pay_date"]; ok {
			trade.PayTime = str2Sec("2006-01-02 15:04:05", paytime.(string))
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
			err := updateTrade(map[string]interface{}{"Id": trade.ID}, trade)
			if err != nil {
				e.Code = SysErrDB
				e.Msg = err.Error()
			}
		}
	}
	return
}

// AuthSigned 支付宝授权签约
// 支付宝签约完成后调用，可用参数 status account MchID
func (A *AliPay) AuthSigned(ctx *Context, req *proto.Account) (auth *proto.Account, e Error) {
	auth = ctx.getAuth("authsigned", req.MchID)
	if auth.ID == "" {
		e.Code = AuthErr
		return
	}
	// 更新状态
	if req.Status != auth.Status {
		auth.Status = req.Status
		// 检测权限是否真实开通
		if auth.Status == proto.Accountstatus_unauth {
			// 临时指定auth状态为proto.Tradestatus_succ 为了后面通过权限验证
			ctx.auth.Status = proto.Accountstatus_ok
			if _, err := A.TradeInfo(ctx, &proto.Trade{MchID: ctx.mchid(), OutTradeID: "tradeforAuthSignedCheck"}); err.Code == AuthErr {
				// 查询订单返回权限错误，说明授权存在问题
				e.Code = AuthErr
				e.Msg = err.Msg
				return
			}
		}
	}
	// 更新信息
	if req.Account != auth.Account {
		auth.Account = req.Account
	}
	// 更新authinfo
	updateToken(ctx.mchid(), ctx.appid(), auth)

	// 更新所有绑定过此auth的用户数据
	updateUserMulti(map[string]interface{}{"MchID": ctx.mchid(), "type": proto.ALIPAY}, map[string]interface{}{"status": auth.Status})
	return
}

// Auth 支付宝授权token
// token 调用后只是授权接口调用权限，还需要支付宝签约完成后调用AuthSigned
// token 每次授权都会变化，新的生效，老的过期
// MchID 商户Id存在时 刷新token，不存在时创建一个新的Auth对象
func (A *AliPay) Auth(ctx *Context, code string) (auth *proto.Account, e Error) {
	params := map[string]interface{}{}
	params["grant_type"] = "authorization_code"
	params["code"] = code
	sysParams := A.sysParams(ctx)
	sysParams["method"] = "alipay.open.auth.token.app"
	sysParams["biz_content"] = string(jsonEncode(params))
	sysParams["sign"] = base64Encode(A.Signer(ctx, sysParams))
	info, err := A.Request(requestSimple{url: ctx.url(), get: httpBuildQuery(sysParams), relKey: "alipay_open_auth_token_app_response", ctx: ctx})
	if err != nil {
		e.Msg = string(jsonEncode(info))
		if v, ok := aliErrMap[err.Error()]; ok {
			e.Code = v
		} else {
			e.Code = AuthErr
		}
	} else {
		// 成功返回
		tmpresult := info.(map[string]interface{})
		MchID := tmpresult["user_id"].(string)
		auth = ctx.getAuth("alipay.auth", MchID)
		auth.Token = tmpresult["app_auth_token"].(string)
		// 保存用户授权
		if auth.ID != "" {
			err = updateToken(ctx.mchid(), ctx.appid(), auth)
		} else {
			auth.ID = randomTimeString()
			auth.MchID = MchID
			auth.From = proto.ALIPAY
			auth.AppID = ctx.appid()
			err = saveToken(auth)
		}
	}
	if err != nil {
		e.Code = SysErrDB
		e.Msg = err.Error()
	}
	return
}

// BindUser 用户绑定
// 将Auth授权绑定到User上去
// 多个用户可使用同一个Auth，可有效防止重复授权导致多个Auth争取token问题
// 绑定了之后 调用其他接口可传UserID查找对应Auth
// 如果业务逻辑不需要绑定，就不要绑定，调用其他接口传MchID即可
func (A *AliPay) BindUser(ctx *Context, req *proto.User) (user *proto.User, e Error) {
	if req.UserID == "" || req.MchID == "" {
		e.Code = SysErrParams
		e.Msg = "UserID MchID 必传"
		return
	}
	auth := ctx.getAuth(req.UserID, req.MchID)
	if auth.ID == "" {
		// 授权不存在
		e.Code = AuthErr
		return
	}
	user = getUser(req.UserID, proto.ALIPAY)
	if user.ID != "" {
		// 存在更新授权
		user.MchID = ctx.mchid()
		user.Status = auth.Status
		updateUser(map[string]interface{}{"UserID": user.UserID}, user)
	} else {
		// 保存授权
		user = &proto.User{
			UserID: req.UserID,
			MchID:  req.MchID,
			Status: auth.Status,
			ID:     randomTimeString(),
			From:   proto.ALIPAY,
		}
		saveUser(user)
	}
	return
}

// Request 发送支付宝请求
func (A *AliPay) Request(d requestSimple) (result interface{}, err error) {
	var next Status
	if getNowSec()-d.ctx.gt() > maxTimeout {
		return nil, http.ErrHandlerTimeout
	}
	result, next, err = A.request(d.url+"?"+d.get, d.relKey)
	if err != nil {
		if d.fs != nil {
			return d.fs(result, next, err)
		}
	}
	return
}

func (A *AliPay) request(url string, relKey string) (interface{}, Status, error) {
	body, err := getRequest(url)
	if err != nil {
		// 网络发起请求失败
		// 需重试
		return nil, netConnErr, err
	}
	result := map[string]interface{}{}
	logrus.Debugf("alipayresult:%+v", string(body))
	if err := jsonDecode(body, &result); err != nil {
		return nil, nextStop, err
	}
	data, ok := result[relKey]
	if !ok {
		return nil, nextStop, newError("支付宝返回数据中不存在" + relKey)
	}
	datamap, ok := data.(map[string]interface{})
	if !ok {
		return nil, nextStop, newError("支付宝返回数据中" + relKey + "参数格式错误")
	}
	next, err := A.errorCheck(datamap)
	return datamap, next, err
}

// Signer 支付宝请求做验签
// 使用应用私钥
func (A *AliPay) Signer(ctx *Context, data map[string]string) (signer []byte) {
	message := mapSortAndJoin(data, "=", "&", true)
	logrus.Infoln(message)
	rng := rand.Reader
	hashed := sha256.Sum256([]byte(message))
	var e error
	signer, e = rsa.SignPKCS1v15(rng, ctx.privateKey(), crypto.SHA256, hashed[:])
	fmt.Println(e)
	return
}

func (A *AliPay) sysParams(ctx *Context) map[string]string {
	return map[string]string{
		"app_id":    ctx.appid(),
		"format":    aliPayDefaultFormat,
		"charset":   aliPayDefaultCharset,
		"sign_type": aliPayDefaultSignType,
		"version":   "1.0",
		"timestamp": sec2Str("2006-01-02 15:04:05", getNowSec()),
	}
}

func (A *AliPay) errorCheck(data map[string]interface{}) (Status, error) {
	code := data["code"].(string)
	subCode := ""
	if v, ok := data["sub_code"]; ok {
		subCode = v.(string)
	}
	switch code {
	case "10000":
		// 成功
		return nextStop, nil
	case "20001":
		// 重新授权后在重试
		return nextWaitAuth, newError(code + subCode)
	case "20000":
		// 立即重试
		return nextRetry, newError(code + subCode)
	case "10003":
		// 循环重试
		return nextWaitRetry, newError(code + subCode)
	default:
		return nextStop, newError(code + subCode)
	}
}

var aliErrMap = map[string]int32{
	"40004ACQ.PAYMENT_AUTH_CODE_INVALId":  PayErrCode,
	"40004ACQ.TRADE_HAS_SUCCESS":          PayErrPayed,
	"40004ACQ.TRADE_NOT_EXIST":            TradeErrNotFound,
	"40004ACQ.TRADE_STATUS_ERROR":         TradeErrStatus,
	"40004ACQ.SELLER_BALANCE_NOT_ENOUGH":  UserErrBalance,
	"40004ACQ.REFUND_AMT_NOT_EQUAL_TOTAL": RefundErrAmount,
	"40004ACQ.ACCESS_FORBIdDEN":           AuthErr,
}
var aliTradeStatusMap = map[string]proto.Tradestatus{
	"WAIT_BUYER_PAY": proto.Tradestatus_waitepay,
	"TRADE_CLOSED":   proto.Tradestatus_closed,
	"TRADE_SUCCESS":  proto.Tradestatus_succ,
	"TRADE_FINISHED": proto.Tradestatus_succ,
}
