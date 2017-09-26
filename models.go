package ppp

import (
	"encoding/gob"

	"gopkg.in/mgo.v2/bson"
)

func init() {
	gob.Register(PayType{})
	gob.Register(Status{})
}

type Status int

const (
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

	TradeQueryErr = 5000 //查询失败

	TradeStatusWaitPay Status = 0  //未支付
	TradeStatusClose          = -1 //取消/退款
	TradeStatusRefund         = -2 //取消/退款
	TradeStatusSucc           = 1  //成功结束

)

type PayType string

const (
	PAYTYPE_ALIPAY PayType = "alipay"
	PAYTYPE_WXPAY          = "wxpay"
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
	Source     PayPype // alipay,wxpay
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

//user
type User struct {
	UserId  string
	Source  string
	Token   string
	ExAt    int64
	ReToken string
}

//用户返回
type UserResult struct {
	Data       User
	SourceData string
	Code       int
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
