package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"path/filepath"

	config "github.com/panjjo/go-config"
	"github.com/panjjo/ppp"
	"github.com/panjjo/ppp/pool"
)

var configPath = flag.String("path", "", "配置文件地址")

func main() {
	//config
	/*configPath = os.Getenv("GOPATH") + "/src/github.com/panjjo/ppp"*/
	flag.Parse()

	err := config.ReadConfigFile(filepath.Join(*configPath, "./config.yml"))
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
	log.Println("a")
	//wxpay
	config.Mod = ppp.PAYTYPE_WXPAY
	if ok, err := config.GetBool("status"); ok {
		initWXPay()
		wx := new(ppp.WXPay)
		rpc.Register(wx)
	} else {
		log.Fatal(err)
	}
	log.Println("a")

	//wxpaysg
	config.Mod = ppp.PAYTYPE_WXPAYSG
	if ok, err := config.GetBool("status"); ok {
		initWXPaySG()
		wxsg := new(ppp.WXPaySG)
		rpc.Register(wxsg)
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
	config.Mod = "sys"
	ppp.InitLog(config.GetStringDefault("loglevel", "warn"))

	addr := config.GetStringDefault("listen", ":1234")
	l, e := net.Listen("tcp", addr)
	ppp.Log.INFO.Printf("Listen at:%s", addr)
	if e != nil {
		log.Fatal("Listen error:", e)
	}
	http.Serve(l, nil)
}

func initAliPay() {
	ali := ppp.AliPayInit{
		ServiceProviderId: config.GetStringDefault("serviceid", ""),
		ConfigPath:        *configPath,
	}
	var err error
	if ali.AppId, err = config.GetString("appid"); err != nil {
		log.Fatal("Init Error:Not Found alipay:appid")
	}
	if ali.Url, err = config.GetString("url"); err != nil {
		log.Fatal("Init Error:Not Found alipay:url")
	}
	if ali.NotifyUrl, err = config.GetString("notify"); err != nil {
		log.Fatal("Init Error:Not Found alipay:notify_url")
	}

	ali.Init()
}
func initWXPay() {
	wx := ppp.WXPayInit{
		ConfigPath: *configPath,
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
	if wx.ApiKey, err = config.GetString("key"); err != nil {
		log.Fatal("Init Error:Not Found wxpay:apikey")
	}
	if wx.NotifyUrl, err = config.GetString("notify"); err != nil {
		log.Fatal("Init Error:Not Found wxpay:notify_url")
	}
	wx.Init()
}

func initWXPaySG() {
	wx := ppp.WXPaySGInit{
		ConfigPath: *configPath,
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
	if wx.ApiKey, err = config.GetString("key"); err != nil {
		log.Fatal("Init Error:Not Found wxpay:apikey")
	}
	if wx.NotifyUrl, err = config.GetString("notify"); err != nil {
		log.Fatal("Init Error:Not Found wxpay:notify_url")
	}
	wx.Init()
}

