package ppp

// Services hprose rpc模式注册服务
type Services struct {
	// alipay 支付宝
	AliPayPayParams  func(req *TradeParams,tag string) (data *PayParams, e Error)   `name:"alipay_PayParams" idempotent:"true"`
	AliPayBarPay     func(req *BarPay,tag string) (trade *Trade, e Error)           `name:"alipay_BarPay" idempotent:"true"`
	AliPayRefund     func(req *Refund,tag string) (refund *Refund, e Error)         `name:"alipay_Refund" idempotent:"true"`
	AliPayCancel     func(req *Trade,tag string) (e Error)                          `name:"alipay_Cancel" idempotent:"true"`
	AliPayTradeInfo  func(req *Trade, sync bool,tag string) (trade *Trade, e Error) `name:"alipay_TradeInfo" idempotent:"true"`
	AliPayAuthSigned func(req *Auth,tag string) (auth *Auth, e Error)               `name:"alipay_AuthSigned" idempotent:"true"`
	AliPayAuth       func(code string,tag string) (auth *Auth, e Error)             `name:"alipay_auth" idempotent:"true"`
	AliPayBindUser   func(req *User,tag string) (user *User, e Error)               `name:"alipay_BindUser" idempotent:"true"`
	AliPayUnBindUser func(req *User,tag string) (user *User, e Error)               `name:"alipay_UnBindUser" idempotent:"true"`
	AliPayMchPay     func(req *MchPay,tag string) (string, Error)                   `name:"alipay_MchPay" idempotent:"true"`

	// wxpay 微信服务商模式
	WXPayBarPay     func(req *BarPay,tag string) (trade *Trade, e Error)           `name:"wxpay_BarPay" idempotent:"true"`
	WXPayRefund     func(req *Refund,tag string) (refund *Refund, e Error)         `name:"wxpay_Refund" idempotent:"true"`
	WXPayTradeInfo  func(req *Trade, sync bool,tag string) (trade *Trade, e Error) `name:"wxpay_TradeInfo" idempotent:"true"`
	WXPayCancel     func(req *Trade,tag string) (e Error)                          `name:"wxpay_Cancel" idempotent:"true"`
	WXPayPayParams  func(req *TradeParams,tag string) (data *PayParams, e Error)   `name:"wxpay_PayParams" idempotent:"true"`
	WXPayBindUser   func(req *User,tag string) (user *User, e Error)               `name:"wxpay_BindUser" idempotent:"true"`
	WXPayAuthSigned func(req *Auth,tag string) (auth *Auth, e Error)               `name:"wxpay_AuthSigned" idempotent:"true"`
	WXPayUnBindUser func(req *User,tag string) (user *User, e Error)               `name:"wxpay_UnBindUser" idempotent:"true"`

	// wxpay_single 微信单商户模式
	WXPaySingleBarPay    func(req *BarPay,tag string) (trade *Trade, e Error)           `name:"wxpay_single_BarPay" idempotent:"true"`
	WXPaySingleRefund    func(req *Refund,tag string) (refund *Refund, e Error)         `name:"wxpay_single_Refund" idempotent:"true"`
	WXPaySingleTradeInfo func(req *Trade, sync bool,tag string) (trade *Trade, e Error) `name:"wxpay_single_TradeInfo" idempotent:"true"`
	WXPaySingleCancel    func(req *Trade,tag string) (e Error)                          `name:"wxpay_single_Cancel" idempotent:"true"`
	WXPaySinglePayParams func(req *TradeParams,tag string) (data *PayParams, e Error)   `name:"wxpay_single_PayParams" idempotent:"true"`
	WXPaySingleMchPay    func(req *MchPay,tag string) (string, Error)                   `name:"wxpay_single_MchPay" idempotent:"true"`
}
