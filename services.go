package ppp

// Services hprose rpc模式注册服务
type Services struct {
	// alipay 支付宝
	AliPayPayParams  func(req *TradeParams,tag string) (data *PayParams, e Error)   `name:"alipay_PayParams"`
	AliPayBarPay     func(req *BarPay,tag string) (trade *Trade, e Error)           `name:"alipay_BarPay"`
	AliPayRefund     func(req *Refund,tag string) (refund *Refund, e Error)         `name:"alipay_Refund"`
	AliPayCancel     func(req *Trade,tag string) (e Error)                          `name:"alipay_Cancel"`
	AliPayTradeInfo  func(req *Trade, sync bool,tag string) (trade *Trade, e Error) `name:"alipay_TradeInfo"`
	AliPayAuthSigned func(req *Auth,tag string) (auth *Auth, e Error)               `name:"alipay_AuthSigned"`
	AliPayAuth       func(code string,tag string) (auth *Auth, e Error)             `name:"alipay_auth"`
	AliPayBindUser   func(req *User,tag string) (user *User, e Error)               `name:"alipay_BindUser"`
	AliPayUnBindUser func(req *User,tag string) (user *User, e Error)               `name:"alipay_UnBindUser"`

	// wxpay 微信服务商模式
	WXPayBarPay     func(req *BarPay,tag string) (trade *Trade, e Error)           `name:"wxpay_BarPay"`
	WXPayRefund     func(req *Refund,tag string) (refund *Refund, e Error)         `name:"wxpay_Refund"`
	WXPayTradeInfo  func(req *Trade, sync bool,tag string) (trade *Trade, e Error) `name:"wxpay_TradeInfo"`
	WXPayCancel     func(req *Trade,tag string) (e Error)                          `name:"wxpay_Cancel"`
	WXPayPayParams  func(req *TradeParams,tag string) (data *PayParams, e Error)   `name:"wxpay_PayParams"`
	WXPayBindUser   func(req *User,tag string) (user *User, e Error)               `name:"wxpay_BindUser"`
	WXPayAuthSigned func(req *Auth,tag string) (auth *Auth, e Error)               `name:"wxpay_AuthSigned"`
	WXPayUnBindUser func(req *User,tag string) (user *User, e Error)               `name:"wxpay_UnBindUser"`

	// wxpay_single 微信单商户模式
	WXPaySingleBarPay    func(req *BarPay,tag string) (trade *Trade, e Error)           `name:"wxpay_single_BarPay"`
	WXPaySingleRefund    func(req *Refund,tag string) (refund *Refund, e Error)         `name:"wxpay_single_Refund"`
	WXPaySingleTradeInfo func(req *Trade, sync bool,tag string) (trade *Trade, e Error) `name:"wxpay_single_TradeInfo"`
	WXPaySingleCancel    func(req *Trade,tag string) (e Error)                          `name:"wxpay_single_Cancel"`
	WXPaySinglePayParams func(req *TradeParams,tag string) (data *PayParams, e Error)   `name:"wxpay_single_PayParams"`
	WXPaySingleMchPay    func(req *MchPay,tag string) (string, Error)                   `name:"wxpay_single_MchPay"`
}
