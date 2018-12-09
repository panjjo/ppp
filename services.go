package ppp

// Services hprose rpc模式注册服务
type Services struct {
	// alipay 支付宝
	AliPayPayParams  func(req *TradeParams) (data *PayParams, e Error)   `name:"alipay_PayParams"`
	AliPayBarPay     func(req *BarPay) (trade *Trade, e Error)           `name:"alipay_BarPay"`
	AliPayRefund     func(req *Refund) (refund *Refund, e Error)         `name:"alipay_Refund"`
	AliPayCancel     func(req *Trade) (e Error)                          `name:"alipay_Cancel"`
	AliPayTradeInfo  func(req *Trade, sync bool) (trade *Trade, e Error) `name:"alipay_TradeInfo"`
	AliPayAuthSigned func(req *Auth) (auth *Auth, e Error)               `name:"alipay_AuthSigned"`
	AliPayAuth       func(code string) (auth *Auth, e Error)             `name:"alipay_auth"`
	AliPayBindUser   func(req *User) (user *User, e Error)               `name:"alipay_BindUser"`
	AliPayUnBindUser func(req *User) (user *User, e Error)               `name:"alipay_UnBindUser"`

	// wxpay 微信服务商模式
	WXPayBarPay     func(req *BarPay) (trade *Trade, e Error)           `name:"wxpay_BarPay"`
	WXPayRefund     func(req *Refund) (refund *Refund, e Error)         `name:"wxpay_Refund"`
	WXPayTradeInfo  func(req *Trade, sync bool) (trade *Trade, e Error) `name:"wxpay_TradeInfo"`
	WXPayCancel     func(req *Trade) (e Error)                          `name:"wxpay_Cancel"`
	WXPayPayParams  func(req *TradeParams) (data *PayParams, e Error)   `name:"wxpay_PayParams"`
	WXPayBindUser   func(req *User) (user *User, e Error)               `name:"wxpay_BindUser"`
	WXPayAuthSigned func(req *Auth) (auth *Auth, e Error)               `name:"wxpay_AuthSigned"`
	WXPayUnBindUser func(req *User) (user *User, e Error)               `name:"wxpay_UnBindUser"`

	// wxpay_single 微信单商户模式
	WXPaySingleBarPay    func(req *BarPay) (trade *Trade, e Error)           `name:"wxpay_single_BarPay"`
	WXPaySingleRefund    func(req *Refund) (refund *Refund, e Error)         `name:"wxpay_single_Refund"`
	WXPaySingleTradeInfo func(req *Trade, sync bool) (trade *Trade, e Error) `name:"wxpay_single_TradeInfo"`
	WXPaySingleCancel    func(req *Trade) (e Error)                          `name:"wxpay_single_Cancel"`
	WXPaySinglePayParams func(req *TradeParams) (data *PayParams, e Error)   `name:"wxpay_single_PayParams"`
	WXPaySingleMchPay    func(req *MchPay) (string, Error)                   `name:"wxpay_single_MchPay"`

	// wxpay_app 微信单商户APP支付
	WXPayAPPBarPay    func(req *BarPay) (trade *Trade, e Error)           `name:"wxpay_app_BarPay"`
	WXPayAPPRefund    func(req *Refund) (refund *Refund, e Error)         `name:"wxpay_app_Refund"`
	WXPayAPPTradeInfo func(req *Trade, sync bool) (trade *Trade, e Error) `name:"wxpay_app_TradeInfo"`
	WXPayAPPCancel    func(req *Trade) (e Error)                          `name:"wxpay_app_Cancel"`
	WXPayAPPPayParams func(req *TradeParams) (data *PayParams, e Error)   `name:"wxpay_app_PayParams"`

	// wxpay_minip 微信单商户小程序支付
	WXPayMINIPPayParams func(req *TradeParams) (data *PayParams, e Error)   `name:"wxpay_minip_PayParams"`
	WXPayMINIPTradeInfo func(req *Trade, sync bool) (trade *Trade, e Error) `name:"wxpay_minip_TradeInfo"`
	WXPayMINIPRefund    func(req *Refund) (refund *Refund, e Error)         `name:"wxpay_minip_Refund"`
}
