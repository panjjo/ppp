package ppp

const (
	Succ    = 1000 //成功
	AuthErr = 9001 //授权错误

	PayErr      = 2000 //支付失败
	PayErrPayed = 2001 //重复支付
	PayErrCode  = 2001 //支付码无效

	TradeErr         = 3000 //交易错误
	TradeErrNotFound = 3001 //交易不存在
	TradeErrStatus   = 3002 //交易状态错误

	RefundErr        = 4000 //退款错误
	RefundErrBalance = 4001 //账户余额错误
	RefundErrAmount  = 4001 //退款金额错误

	TradeStatusWaitPay = 0  //未支付
	TradeStatusClose   = -1 //取消/退款
	TradeStatusSucc    = 1  //成功结束
)

type rsys struct {
	retry int
	time  int64
}

//支付结果
type PayResult struct {
	Code       int    `description:"状态码"`
	SourceData string `description:"第三方原始返回"`
	TradeId    string
	OutTradeId string
	Amount     int64
	PayTime    int64
}

//条码支付请求
type BarCodePayRequest struct {
	OutTradeId string      `json:"out_trade_id" description:"商户交易ID 唯一"`
	TradeName  string      `json:"trade_name" description:"名称"`
	Amount     int64       `json:"amount" description:"交易总额,单位分"`
	ItemDes    interface{} `json:"item_des" description:"商品表述"`
	AuthCode   string      `json:"auth_code" description:"授权码"`
	UserId     string      `json:"userid" description:"收款方对应的userid"`
	ShopId     string      `json:"shopid" description:"店铺ID"`
	r          rsys
}

//支付单详情
type TradeRequest struct {
	OutTradeId string `json:"out_trade_id" description:"交易ID"`
	TradeId    string `description:"第三方交易ID "`
	UserId     string `json:"userid" description:"权限对应的UserId"`
	r          rsys
}

//支付单详情
type Trade struct {
	TradeId    string
	OutTradeId string
	Status     int
	Amount     int64
}

//支付单返回
type TradeResult struct {
	Trade
	SourceData string
	Code       int
}

//刷新token
type RefreshToken struct {
	Type   string `json:"type" description:"刷新方式 refush 刷新，code 第一次获取"`
	Code   string `json:"code" description:"第一次获取时需要传入兑换码"`
	UserId string `json:"userid" description:"权限对应的UserId"`
	r      rsys
}

//退款请求
type RefundRequest struct {
	Memo       string
	Amount     int64
	OutTradeId string
	TradeId    string
	RefundId   string
	UserId     string
	r          rsys
}

//退款单
type Refund struct {
	TradeNo    string
	RefundId   string
	OutTradeId string
	Memo       string
	Amount     int64
}

//退款返回
type RefundResult struct {
	Refund
	Code       int
	SourceData string
}

//user
type User struct {
	UserId  string
	Type    string
	Token   string
	ExAt    int64
	ReToken string
}

//用户返回
type UserResult struct {
	User
	SourceData string
	Code       int
}

//通用返回
type Response struct {
	SourceData string
	Code       int
}
