package ppp

import (
	"crypto/rsa"
	"crypto/tls"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strings"

	"github.com/panjjo/ppp/db"
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
	Code int32
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

	WXSingle Config `yaml:"wxpay_single" mapstructure:"wxpay_single"`

	DB db.Config `yaml:"database" mapstructure:"database"`

	Sys SysConfig `yaml:"sys"`
}
type Config struct {
	ConfigSingle `yaml:"default" mapstructure:"default"`
	Apps         []ConfigSingle `yaml:"apps" mapstructure:"apps"` // 多个app
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
	viper.SetConfigFile(name)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatal("init config error:", err)
	}
	config := &Configs{}
	err = viper.Unmarshal(config)
	if err != nil {
		logrus.Fatal("init config unmarshal error:", err)
	}
	return config
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
