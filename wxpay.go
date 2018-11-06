package ppp

const (
	// WXPAY 微信支付服务商的标识
	WXPAY string = "wxpay"
)

// WXPay 微信支付服务商模式主体
// 微信支付服务商模式
// 服务商模式与单商户模式区别只是多了一个 子商户权限，其余接口结构返回完全一致
type WXPay struct {
	ws WXPaySingle
	t  string
	rs rs
}

// NewWXPay 获取微信实例
func NewWXPay(config Config) *WXPay {
	ws := NewWXPaySingle(config)
	wx := &WXPay{ws: *ws}
	wx.t = WXPAY
	return wx
}

// BarPay 商户主动扫码支付
// 服务商模式调用
// 真实处理为 WXPaySingle.BarPay
func (W *WXPay) BarPay(req *BarPay) (trade *Trade, e Error) {
	auth := W.token(req.UserID, req.MchID)
	if auth.Status != AuthStatusSucc {
		//授权错误
		e.Code = AuthErr
		return
	}
	return W.ws.BarPay(req)
}

// Refund 订单退款
// 服务商模式调用
// 真实处理为 WXPaySingle.Refund
func (W *WXPay) Refund(req *Refund) (refund *Refund, e Error) {
	auth := W.token(req.UserID, req.MchID)
	if auth.Status != AuthStatusSucc {
		//授权错误
		e.Code = AuthErr
		return
	}
	return W.ws.Refund(req)
}

// Cancel 撤销订单
// 服务商模式调用
// 真实处理为 WXPaySingle.Cancel
func (W *WXPay) Cancel(req *Trade) (e Error) {
	auth := W.token(req.UserID, req.MchID)
	if auth.Status != AuthStatusSucc {
		//授权错误
		e.Code = AuthErr
		return
	}
	return W.ws.Cancel(req)
}

// TradeInfo 获取订单详情
// 服务商模式调用
// 真实处理为 WXPaySingle.TradeInfo
func (W *WXPay) TradeInfo(req *Trade, sync bool) (trade *Trade, e Error) {
	auth := W.token(req.UserID, req.MchID)
	if auth.Status != AuthStatusSucc {
		//授权错误
		e.Code = AuthErr
		return
	}
	return W.ws.TradeInfo(req, sync)
}

// PayParams 获取支付参数
// 用于前段请求，不想暴露证书的私密信息的可用此方法组装请求参数，前端只负责请求
// 支持的有 JS支付，手机app支付，公众号支付
// APP支付紧支持单商户模式，公众号支付，扫码支付等支持服务商和单商户模式
func (W *WXPay) PayParams(req *TradeParams) (data *PayParams, e Error) {
	auth := W.token(req.UserID, req.MchID)
	if auth.Status != AuthStatusSucc {
		//授权错误
		e.Code = AuthErr
		return
	}
	return W.ws.PayParams(req)
}

// BindUser 用户绑定
// 将Auth授权绑定到User上去
// 多个用户可使用同一个Auth，可有效防止重复授权导致多个Auth争取token问题
// 绑定了之后 调用其他接口可传UserID查找对应Auth
// 如果业务逻辑不需要绑定，就不要绑定，调用其他接口传MchID即可
func (W *WXPay) BindUser(req *User) (user *User, e Error) {
	if req.UserID == "" || req.MchID == "" {
		e.Code = SysErrParams
		e.Msg = "userid mchid 必传"
		return
	}
	auth := getToken(req.MchID, W.t)
	if auth.ID == "" {
		// 授权不存在
		e.Code = AuthErr
		return
	}
	user = getUser(req.UserID, W.t)
	if user.ID != "" {
		//存在更新授权
		user.MchID = auth.MchID
		user.Status = auth.Status
		updateUser(map[string]interface{}{"userid": user.UserID}, user)
	} else {
		//保存授权
		user = &User{
			UserID: req.UserID,
			MchID:  req.MchID,
			Status: auth.Status,
			ID:     randomTimeString(),
			From:   W.t,
		}
		saveUser(user)
	}
	return
}

// UnBindUser 用户解除绑定
// 将Auth授权和User进行解绑
// 多个用户可使用同一个Auth，可有效防止重复授权导致多个Auth争取token问题
// 解绑之后auth授权依然有效
func (W *WXPay) UnBindUser(req *User) (user *User, e Error) {
	if req.UserID == "" {
		e.Code = SysErrParams
		e.Msg = "userid  必传"
		return
	}
	user = getUser(req.UserID, ALIPAY)
	if user.ID != "" {
		//存在更新授权
		user.MchID = ""
		user.Status = UserWaitVerify
		updateUser(map[string]interface{}{"userid": user.UserID}, user)
	} else {
		//用户不存在
		e.Code = UserErrNotFount
	}
	return
}

// AuthSigned 增加授权
// 刷新/获取授权
// 传入参数为Token格式,微信传入MchId：子商户ID
// req mchid account appid
func (W *WXPay) AuthSigned(req *Auth) (auth *Auth, e Error) {
	auth = W.token("", req.MchID)
	if auth.ID != "" {
		//授权已存在直接返回
		return
	}
	auth = &Auth{
		ID:      randomTimeString(),
		Status:  AuthStatusSucc,
		MchID:   req.MchID,
		From:    WXPAY,
		Account: req.Account,
		AppID:   req.AppID,
	}
	//检测权限是否真实开通
	//临时指定auth状态为AuthStatusSucc 为了后面通过权限验证
	W.rs.auth = auth
	W.ws.rs.auth = auth
	if _, err := W.TradeInfo(&Trade{MchID: auth.MchID, TradeID: "tradeforAuthSignedCheck"}, true); err.Code == AuthErr {
		//查询订单返回权限错误，说明授权存在问题
		e.Code = AuthErr
		e.Msg = err.Msg
		return
	}

	//保存authinfo
	saveToken(auth)
	//更新所有绑定过此auth的用户数据
	updateUserMulti(map[string]interface{}{"mchid": auth.MchID, "type": W.t}, map[string]interface{}{"status": UserSucc})
	return
}

func (W *WXPay) token(userid, mchid string) *Auth {
	if W.rs.auth != nil {
		return W.rs.auth
	}
	W.rs.auth = token(userid, mchid, WXPAY)
	W.ws.rs.auth = W.rs.auth
	return W.rs.auth
}
