package ppp

import "github.com/panjjo/ppp/db"

const (
	tradeTable = "trades"

	// TradeStatusWaitPay 订单未支付
	TradeStatusWaitPay Status = 0
	// TradeStatusPaying 等待用户输入密码
	TradeStatusPaying Status = 2
	// TradeStatusClose 订单取消/退款
	TradeStatusClose Status = -1
	// TradeStatusRefund 订单取消/退款
	TradeStatusRefund Status = -2
	// TradeStatusSucc 订单成功结束
	TradeStatusSucc Status = 1

	// BARPAY 条码支付(商家扫码)
	BARPAY TradeType = "BAR"
	// WAPPAY 手机网页支付
	WAPPAY TradeType = "WAP"
	// APPPAY app支付
	APPPAY TradeType = "APP"
	// WEBPAY 网站支付
	WEBPAY TradeType = "WEB"
	// JSPAY 公众号支付
	JSPAY TradeType = "JS"
	// MINIPAY 小程序支付
	MINIPAY TradeType = "MINIP"
	// CBARPAY 微信扫码支付（顾客扫码）
	CBARPAY TradeType = "CBAR"
)

// TradeType  订单类型
type TradeType = string

// Trade 交易结构
type Trade struct {
	OutTradeID string
	TradeID    string
	Amount     int64
	ID         string
	Status     Status
	Type       TradeType
	MchID      string
	UserID     string
	UpTime     int64
	PayTime    int64
	Create     int64
	// AppID  收款方id
	AppID string
	// Form 单据来源 alipay wxpay
	From string
}

// TradeParams  支付参数请求结构
type TradeParams struct {
	ReturnURL  string // 回调地址,非异步通知地址
	OutTradeID string // 商户交易ID 唯一
	TradeName  string // 名称
	Amount     int64  // 交易总额,单位分
	ItemDes    string // 商品表述
	ShopID     string // 店铺ID
	Ex         string // 共用回传参数
	UserID     string // 支付宝使用服务商模式中的自身收款，微信支付有些需要传UserID
	MchID      string
	IPAddr     string
	Scene      TradeScene // 场景
	OpenID     string     // 与sub_openid 二选一  支付用户在公众号或小程序的用户openid
	SubOpenID  string     // 与openid 二选一  支付用户在公众号或小程序的openid
	SubAppID   string     // 子商户appid,服务商模式使用，公众号支付为子商户公众号appid，小程序为子商户小程序appid,不传模式使用AuthSigned时传入的subappid
	Type       TradeType  // 订单类型，公众号支付：JSAPI,顾客扫码支付：NATIVE，app支付：APP，小程序支付：MINIP,
	NotifyURL  string     // 异步通知地址， 不传默认使用配置文件中的设置
}

// TradeScene 订单支付场景
type TradeScene struct {
	// 详情看wxpay的统一下单中的scene
	URL  string // 请求地址
	Name string // 请求名称
}

func getTrade(q interface{}) *Trade {
	trade := &Trade{}
	DBClient.Get(tradeTable, q, trade)
	return trade
}

func saveTrade(trade *Trade) error {
	return DBClient.Insert(tradeTable, trade)
}

func updateTrade(query, update interface{}) error {
	return DBClient.Update(tradeTable, query, db.M{"$set": update})
}
