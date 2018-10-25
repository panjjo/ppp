package ppp2

const (
	tradeTable = "trades"

	//TradeStatusWaitPay 订单未支付
	TradeStatusWaitPay Status = 0
	//TradeStatusClose 订单取消/退款
	TradeStatusClose Status = -1
	//TradeStatusRefund 订单取消/退款
	TradeStatusRefund Status = -2
	//TradeStatusSucc 订单成功结束
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
	// CBARPAY 微信扫码支付（顾客扫码）
	CBARPAY TradeType = "CBAR"
)

// TradeType  订单类型
type TradeType string

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
	SourceID   string
	From       string //单据来源 alipay wxpay
}

// TradeParams  支付参数请求结构
type TradeParams struct {
	ReturnURL  string //回调地址,非异步通知地址
	OutTradeID string //商户交易ID 唯一
	TradeName  string //名称
	Amount     int64  //交易总额,单位分
	ItemDes    string //商品表述
	ShopID     string //店铺ID
	Ex         string //共用回传参数
	UserID     string //支付宝使用服务商模式中的自身收款，微信支付有些需要传UserID
	MchID      string
	IPAddr     string
	Scene      TradeScene //场景
	OpenID     string     //与sub_openid二选一 公众号支付必传，openid为在服务商公众号的id
	SubOpenID  string     //与openid 二选一 公众号支付必传，sub_openid为在子商户公众号的id
	Type       TradeType  //订单类型，网页支付公众号支付：JSAPI,扫码支付：NATIVE，app支付：APP
}

// TradeScene 订单支付场景
type TradeScene struct {
	//详情看wxpay的统一下单中的scene
	URL  string //请求地址
	Name string //请求名称
}

func getTrade(q interface{}) *Trade {
	session := DBPool.Get()
	defer session.Close()
	trade := &Trade{}
	res := session.FindOne(tradeTable, q, trade)
	if res != nil {
		trade = res.(*Trade)
	}
	return trade
}

func saveTrade(trade *Trade) error {
	session := DBPool.Get()
	defer session.Close()
	return session.Save(tradeTable, trade)
}

func updateTrade(query, update interface{}) error {
	session := DBPool.Get()
	defer session.Close()
	return session.Update(tradeTable, query, update)
}
func upsertTrade(query, update interface{}) (interface{}, error) {
	session := DBPool.Get()
	defer session.Close()
	return session.UpSert(tradeTable, query, update)
}
