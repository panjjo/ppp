package ppp

import (
	"gopkg.in/mgo.v2/bson"
)

type Status int

const (
	AuthErr = 9001 //授权错误

	SysErrParams = 1001 //参数错误

	PayErr      = 2000 //支付失败
	PayErrPayed = 2001 //重复支付
	PayErrCode  = 2001 //支付码无效

	TradeErr         = 3000 //交易错误
	TradeErrNotFound = 3001 //交易不存在
	TradeErrStatus   = 3002 //交易状态错误

	RefundErr       = 4000 //退款错误
	RefundErrAmount = 4001 //退款金额错误
	RefundErrExpire = 4002 //退款以超期

	TradeQueryErr = 5000 //查询失败

	UserErrBalance  = 6001 //账户余额错误
	UserErrRegisted = 6002 //账户已存在
	UserErrNotFount = 6603 //账户不存在

	TradeStatusWaitPay Status = 0  //未支付
	TradeStatusClose   Status = -1 //取消/退款
	TradeStatusRefund  Status = -2 //取消/退款
	TradeStatusSucc    Status = 1  //成功结束

	UserWaitVerify Status = 0  //等待审核或等待授权
	UserFreeze     Status = -1 //冻结
	UserSucc       Status = 1  //正常
)

const (
	PAYTYPE_ALIPAY = "alipay"
	PAYTYPE_WXPAY  = "wxpay"
	PAYTYPE_PPP    = "ppp"
)

type rsys struct {
	retry int
	time  int64
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
	TradeId    string //第三方ID
	OutTradeId string //自定义ID
	Status     Status //1:完成， -1：取消
	Type       int    //1:入账，-1：出账
	Amount     int64
	Source     string // alipay,wxpay
	ParentId   string //来源主ID
	PayTime    int64
	UpTime     int64
	Ex         interface{}
	Id         string // PPPID
	Memo       string
}

//支付单返回
type TradeResult struct {
	Data       Trade
	SourceData string
	Code       int
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

//刷新token
type Token struct {
	Code    string `json:"code" description:"第一次获取时需要传入兑换码"`
	refresh bool
	r       rsys
}

//授权
type authBase struct {
	Id      string
	Token   string
	ExAt    int64  //token失效日期
	ReToken string //refresh_token
	MchId   string
	Type    string
}
type Auth struct {
	Id    string
	MchId string
	Type  string
}
type AuthResult struct {
	Data       Auth
	SourceData string
	Code       int
}

//账户
type User struct {
	Id     string
	UserId string //外部用户id
	MchId  string //第三方id
	Status Status
	Amount int64 //账户余额
	Type   string
}

//用户返回
type AccountResult struct {
	Data       User
	SourceData string
	Code       int
}

//用户授权
type AccountAuth struct {
	UserId string
	MchId  string
	Type   string
}

//通用返回
type Response struct {
	SourceData string
	Code       int
}

//列表查询
type ListRequest struct {
	Query       bson.M
	Skip, Limit int
	Sort        string
	r           rsys
}

//总数返回
type CountResult struct {
	Data       int
	Code       int
	SourceData string
}

//对账单列表
type TradeListResult struct {
	Code       int
	Data       []Trade
	SourceData string
}
