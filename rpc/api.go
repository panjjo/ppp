package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"path/filepath"

	"ppp"

	"github.com/julienschmidt/httprouter"
)

// Response api返回统一格式
type Response struct {
	Code int
	Msg  interface{}
	Data interface{}
}

var alipay *ppp.AliPay
var configPath = flag.String("path", "", "配置文件地址")

func jsonEncode(ob interface{}) []byte {
	if b, err := json.Marshal(ob); err == nil {
		return b
	}
	return []byte("")
}
func writeResponse(w http.ResponseWriter, data Response) {
	w.Write(jsonEncode(data))
}

//Auth 授权
func Auth(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	auth, err := alipay.Auth(ps.ByName("code"))
	writeResponse(w, Response{Code: err.Code, Msg: err.Msg, Data: auth})
}

func main1() {
	alipay = &ppp.AliPay{}
	flag.Parse()

	config := ppp.LoadConfig(filepath.Join(*configPath, "./config.yml"))
	ppp.NewLogger(config.Sys.LogLevel)
	ppp.NewDBPool(&config.DB)

	router := httprouter.New()
	router.GET("/alipay/auth", Auth)

	l, e := net.Listen("tcp", config.Sys.ADDR)
	ppp.Log.INFO.Printf("Listen at:%s", config.Sys.ADDR)
	if e != nil {
		log.Fatal("Listen error:", e)
	}
	http.Serve(l, nil)

	log.Fatal(http.ListenAndServe(":8081", router))
}
