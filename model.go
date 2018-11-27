package ppp

import (
	"io/ioutil"

	"github.com/panjjo/ppp/db"

	yaml "gopkg.in/yaml.v1"
)

//Status 类型int
type Status int

const (
	//Succ 请求成功
	Succ = 0

	//next
	netConnErr    Status = -1
	nextStop      Status = 0
	nextWaitAuth  Status = 1
	nextRetry     Status = 2
	nextWaitRetry Status = 3

	//AuthErr 授权错误
	AuthErr = 9001
	//AuthErrNotSigned 未签约
	AuthErrNotSigned = 9002

	//SysErrParams 参数错误
	SysErrParams = 1001
	//SysErrVerify 验签错误
	SysErrVerify = 1002
	//SysErrDB 数据库操作错误
	SysErrDB = 1003

	//PayErr 支付失败
	PayErr = 2000
	//PayErrPayed 重复支付
	PayErrPayed = 2001
	//PayErrCode 支付码无效
	PayErrCode = 2002

	//TradeErr 交易错误
	TradeErr = 3000
	//TradeErrNotFound 交易不存在
	TradeErrNotFound = 3001
	//TradeErrStatus 交易状态错误
	TradeErrStatus = 3002

	//RefundErr 退款错误
	RefundErr = 4000
	//RefundErrAmount 退款金额错误
	RefundErrAmount = 4001
	//RefundErrExpire 退款以超期
	RefundErrExpire = 4002

	//TradeQueryErr 查询失败
	TradeQueryErr = 5000

	//UserErrBalance 账户余额错误
	UserErrBalance = 6001
	//UserErrRegisted 账户已存在
	UserErrRegisted = 6002
	//UserErrNotFount 账户不存在
	UserErrNotFount = 6003
)

//Error 错误类型
type Error struct {
	Code int
	Msg  string
}

type requestSimple struct {
	url  string
	get  string
	body []byte
	fs   func(interface{}, Status, error) (interface{}, error)
	tls  bool

	relKey string
}

// Configs 配置文件地址
type Configs struct {
	AliPay Config `yaml:"alipay"`
	WXPay  Config `yaml:"wxpay"`

	WXSingle WXSingleConfig `yaml:"wxpay_single"`

	DB db.Config `yaml:"database"`

	Sys SysConfig `yaml:"sys"`
}

// WXSingleConfig 单商户模式配置
type WXSingleConfig struct {
	//单商户模式的微信支付，app支付必须单独一套
	APP Config `yaml:"app"`
	// 其他:公众号，扫码，h5等
	Other Config `yaml:"other"`
	// 小程序支付
	MINIP Config `yaml:"minip"`
}

//Config 单项配置文件
type Config struct {
	Use       bool   `yaml:"use"`
	AppID     string `yaml:"appid"`
	Secret    string `yaml:"secret"`
	CertPath  string `yaml:"certpath"`
	URL       string `yaml:"url"`
	ServiceID string `yaml:"serviceid"`
	Notify    string `yaml:"notify"`
}

// SysConfig 系统配置文件
type SysConfig struct {
	ADDR     string `yaml:"addr"`
	LogLevel string `yaml:"loglevel"`
}

// LoadConfig 加载配置文件
func LoadConfig(name string) *Configs {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		Log.ERROR.Panicf("load config file error,file:%s,err:%v", name, err)
	}
	config := &Configs{}
	if err = yaml.Unmarshal(b, &config); err != nil {
		Log.ERROR.Panicf("load config file error,file:%s,err:%v", name, err)
	}
	return config
}

// BarPay  商户主扫支付请求数据
type BarPay struct {
	OutTradeID string //商户交易ID 唯一
	TradeName  string //名称
	Amount     int64  //交易总额,单位分
	ItemDes    string //商品表述
	AuthCode   string //授权码
	UserID     string //收款方对应的userid
	MchID      string //商户号：非服务商模式收款不会存在user信息，可直接传mchid
	ShopID     string //店铺ID
	IPAddr     string
}

//rs 请求过程中的一些信息
type rs struct {
	t      int64 //请求开始时间
	auth   *Auth //权限信息
	userid string
}

// PayParams 获取支付参数
type PayParams struct {
	// SourceData 支付参数的map数据jsonencode
	SourceData string
	// Params urlencode的数据
	Params string
}
