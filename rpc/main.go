package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path/filepath"

	config "github.com/panjjo/go-config"
	"github.com/panjjo/ppp"
	"github.com/panjjo/ppp/pool"
)

var configPath string

func main() {
	//config
	configPath = os.Getenv("GOPATH") + "/src/github.com/panjjo/ppp"

	err := config.ReadConfigFile(filepath.Join(configPath, "/config.yml"))
	if err != nil {
		panic(err)
	}
	statement := new(ppp.Statement)
	rpc.Register(statement)

	//account
	account := new(ppp.Account)
	rpc.Register(account)

	//alipay
	config.Mod = ppp.PAYTYPE_ALIPAY
	if ok, err := config.GetBool("status"); ok {
		initAliPay()
		ali := new(ppp.AliPay)
		rpc.Register(ali)
	} else {
		log.Fatal(err)
	}

	//wxpay
	config.Mod = ppp.PAYTYPE_WXPAY
	if ok, err := config.GetBool("status"); ok {
		initWXPay()
		wx := new(ppp.WXPay)
		rpc.Register(wx)
	} else {
		log.Fatal(err)
	}

	//db
	config.Mod = "database"
	ppp.DBPool = pool.GetPool(&pool.Config{
		Addr:      config.GetStringDefault("host", ""),
		Port:      config.GetIntDefault("port", 0),
		DB:        config.GetStringDefault("db", "ppp"),
		MaxActive: config.GetIntDefault("max", 100),
	})

	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("Listen error:", e)
	}
	http.Serve(l, nil)
}

func initAliPay() {
	ali := ppp.AliPayInit{
		ServiceProviderId: config.GetStringDefault("serviceid", ""),
		ConfigPath:        configPath,
	}
	var err error
	if ali.AppId, err = config.GetString("appid"); err != nil {
		log.Fatal("Init Error:Not Found wxpay:appid")
	}
	if ali.Url, err = config.GetString("url"); err != nil {
		log.Fatal("Init Error:Not Found wxpay:url")
	}
	ali.Init()
}
func initWXPay() {
	wx := ppp.WXPayInit{
		ConfigPath: configPath,
	}
	var err error
	if wx.AppId, err = config.GetString("appid"); err != nil {
		log.Fatal("Init Error:Not Found wxpay:appid")
	}
	if wx.Url, err = config.GetString("url"); err != nil {
		log.Fatal("Init Error:Not Found wxpay:url")
	}
	if wx.MchId, err = config.GetString("mchid"); err != nil {
		log.Fatal("Init Error:Not Found wxpay:mchid")
	}
	wx.Init()
}
