package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"

	"ppp2"

	"github.com/julienschmidt/httprouter"
)

// Response api返回统一格式
type Response struct {
	Code int
	Msg  interface{}
	Data interface{}
}

var alipay *ppp2.AliPay

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
	alipay = &ppp2.AliPay{}

	config := ppp2.LoadConfig("/Users/panjjo/work/go/src/ppp2/config.yml")
	ppp2.NewLogger(config.Sys.LogLevel)
	ppp2.NewDBPool(&config.DB)

	router := httprouter.New()
	router.GET("/alipay/auth", Auth)

	l, e := net.Listen("tcp", config.Sys.ADDR)
	ppp2.Log.INFO.Printf("Listen at:%s", config.Sys.ADDR)
	if e != nil {
		log.Fatal("Listen error:", e)
	}
	http.Serve(l, nil)

	log.Fatal(http.ListenAndServe(":8081", router))
}
