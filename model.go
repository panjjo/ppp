package ppp

import (
	"crypto/rsa"
	"crypto/tls"
	"io/ioutil"

	"github.com/panjjo/ppp/db"

	"gopkg.in/yaml.v1"
)

// Status 类型int
type Status int

const (
	// Succ 请求成功
	Succ = 0

	// next
	netConnErr    Status = -1
	nextStop      Status = 0
	nextWaitAuth  Status = 1
	nextRetry     Status = 2
	nextWaitRetry Status = 3

	// AuthErr 授权错误
	AuthErr = 9001
	// AuthErrNotSigned 未签约
	AuthErrNotSigned = 9002

	// SysErrParams 参数错误
	SysErrParams = 1001
	// SysErrVerify 验签错误
	SysErrVerify = 1002
	// SysErrDB 数据库操作错误
	SysErrDB = 1003

	// PayErr 支付失败
	PayErr = 2000
	// PayErrPayed 重复支付
	PayErrPayed = 2001
	// PayErrCode 支付码无效
	PayErrCode = 2002

	// TradeErr 交易错误
	TradeErr = 3000
	// TradeErrNotFound 交易不存在
	TradeErrNotFound = 3001
	// TradeErrStatus 交易状态错误
	TradeErrStatus = 3002

	// RefundErr 退款错误
	RefundErr = 4000
	// RefundErrAmount 退款金额错误
	RefundErrAmount = 4001
	// RefundErrExpire 退款以超期
	RefundErrExpire = 4002

	// TradeQueryErr 查询失败
	TradeQueryErr = 5000

	// UserErrBalance 账户余额错误
	UserErrBalance = 6001
	// UserErrRegisted 账户已存在
	UserErrRegisted = 6002
	// UserErrNotFount 账户不存在
	UserErrNotFount = 6003
)

// Error 错误类型
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
	ctx  *Context

	relKey string
}

// Configs 配置文件地址
type Configs struct {
	AliPay Config `yaml:"alipay"`
	WXPay  Config `yaml:"wxpay"`

	WXSingle Config `yaml:"wxpay_single"`

	DB db.Config `yaml:"database"`

	Sys SysConfig `yaml:"sys"`
}
type Config struct {
	ConfigSingle `yaml:",inline"`
	Apps         []ConfigSingle `yaml:"apps"` // 多个app
}

// ConfigSingle 单项配置文件
type ConfigSingle struct {
	Use       bool     `yaml:"use"`
	Tag       string   `yaml:"tag"`
	AppID     string   `yaml:"appid"`
	AppIDS    []string `yaml:"appids"` // 存在多个appid使用同一个商户号
	Secret    string   `yaml:"secret"`
	CertPath  string   `yaml:"certpath"`
	URL       string   `yaml:"url"`
	ServiceID string   `yaml:"serviceid"`
	Notify    string   `yaml:"notify"`
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
	OutTradeID string // 商户交易ID 唯一
	TradeName  string // 名称
	Amount     int64  // 交易总额,单位分
	ItemDes    string // 商品表述
	AuthCode   string // 授权码
	UserID     string // 收款方对应的userid
	MchID      string // 商户号：非服务商模式收款不会存在user信息，可直接传mchid
	ShopID     string // 店铺ID
	IPAddr     string
}

// MchPay 企业付款请求数据
type MchPay struct {
	OutTradeID string // 商户交易号
	OpenID     string // appid下用户标识
	UserName   string // 真实姓名
	Amount     int64  // 付款金额
	Desc       string // 付款备注
	IPAddr     string // ipdizhi
}

// PayParams 获取支付参数
type PayParams struct {
	// SourceData 支付参数的map数据jsonencode
	SourceData string
	// Params urlencode的数据
	Params string
}

type config struct {
	appid     string
	private   *rsa.PrivateKey // 应用私钥
	public    *rsa.PublicKey  // 支付宝公钥
	url       string
	serviceid string
	notify    string      // 异步回调地址
	tlsConfig *tls.Config // 应用私钥
	secret    string      // 支付密钥
}
