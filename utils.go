package ppp

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

/**
timeid
*/
type U struct {
	prefix string
	c      chan int
	d      chan struct{}
}

func NewU(t int64, n int) *U {
	u := &U{
		prefix: time.Unix(t, 0).Format("060102150405"),
		c:      make(chan int, n),
		d:      make(chan struct{}),
	}
	u.start()
	return u
}
func (u *U) start() {
	go func() {
		i := 0
		for {
			select {
			case u.c <- i:
				i++
			case <-u.d:
				return
			}
		}
	}()
}
func (u *U) stop() {
	u.d <- struct{}{}
	close(u.c)
}

func (u *U) Next() string {
	return u.prefix + fmt.Sprintf("%d", <-u.c)
}

type TimeID struct {
	o *U
	c *U
	n *U

	l int
}

func NewTimeID(l int) *TimeID {
	return &TimeID{l: l}
}
func (u *TimeID) Start() error {
	go func() {
		t := time.NewTicker(time.Second)
		u.n = NewU(time.Now().Unix(), u.l)
		for {
			u.o = u.c
			u.c = u.n
			u.n = NewU(time.Now().Unix()+1, u.l)
			if u.o != nil {
				u.o.stop()
			}
			<-t.C
		}
	}()
	for u.c == nil {
		time.Sleep(1 * time.Millisecond)
	}
	return nil
}
func (u *TimeID) Next() string {
	return u.c.Next()
}

var _systemID *TimeID

func init() {
	_systemID = NewTimeID(10)
	_systemID.Start()
}

func randomTimeString() string {
	return _systemID.Next()
}

/**
  字符串md5
*/
func makeMd5(str string) string {
	h := md5.New()
	io.WriteString(h, str)
	s := fmt.Sprintf("%x", h.Sum(nil))
	return s
}

/**
  生成随机字符串
*/
func randomString(lens int) string {
	now := time.Now()
	return makeMd5(strconv.FormatInt(now.UnixNano(), 10))[:lens]
}

/**
  转化时间戳
*/
func str2Sec(layout, str string) int64 {
	tm2, _ := time.ParseInLocation(layout, str, time.Local)
	return tm2.Unix()
}

/**
  时间戳格式化
*/
func sec2Str(layout string, sec int64) string {
	t := time.Unix(sec, 0)
	nt := t.Format(layout)
	return nt
}

/**
  获取当前时间戳
*/
func getNowSec() int64 {
	return time.Now().Unix()
}

/**
  压json
*/
func jsonEncode(ob interface{}) []byte {
	if b, err := json.Marshal(ob); err == nil {
		return b
	}
	return []byte("")
}

/**
  解json
*/
func jsonDecode(data []byte, ob interface{}) error {
	return json.Unmarshal(data, ob)
}

var base64Base = base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")

/**
  压base64
*/
func base64Encode(data []byte) string {
	return base64Base.EncodeToString(data)
}

/**
解base64
*/
func base64Decode(src string) []byte {
	byt, err := base64Base.DecodeString(src)
	if err != nil {
		return []byte{}
	}
	return byt
}

/**
  将struct转化为map，tag：json,xml
*/
func structToMap(obj interface{}, tag string) map[string]string {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	m := map[string]string{}
	for i := 0; i < t.NumField(); i++ {
		fv := v.Field(i)
		t := t.Field(i).Tag.Get(tag)
		switch v.Field(i).Interface().(type) {
		case string:
			m[t] = fv.String()
		case int, int64:
			m[t] = strconv.FormatInt(fv.Int(), 10)
		}
	}
	return m
}

/**
  map排序
*/

type mapSorter []sortItem
type sortItem struct {
	Key string      `json:"key"`
	Val interface{} `json:"val"`
}

func (ms mapSorter) Len() int {
	return len(ms)
}
func (ms mapSorter) Less(i, j int) bool {
	return ms[i].Key < ms[j].Key // 按键排序
}
func (ms mapSorter) Swap(i, j int) {
	ms[i], ms[j] = ms[j], ms[i]
}

/**
  map排序并根据排序结果kv拼接,empty:是否去除空值
*/
func mapSortAndJoin(m map[string]string, step1, step2 string, empty bool) string {
	ms := make(mapSorter, 0, len(m))
	for k, v := range m {
		ms = append(ms, sortItem{k, v})
	}
	sort.Sort(ms)
	s := []string{}
	for _, p := range ms {
		if p.Val.(string) != "" || !empty {
			s = append(s, p.Key+step1+p.Val.(string))
		}
	}
	return strings.Join(s, step2)
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// float64 四舍五入取整
func round(x float64) int64 {
	return int64(math.Round(x))
}

/**
  Urlencode
*/
func httpBuildQuery(params map[string]string) string {
	qs := url.Values{}
	for k, v := range params {
		qs.Add(k, v)
	}
	return qs.Encode()
}

/*
	发送带有超时的Get请求
*/
func getRequest(url string) ([]byte, error) {
	client := timeoutClient()
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("GetRequest:", url, "Error:", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respbody, nil
}

/*
	发送带有超时的Post请求
*/
func postRequest(url, contentType string, body io.Reader) ([]byte, error) {
	client := timeoutClient()
	resp, err := client.Post(url, contentType, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respbody, nil
}

/*
	发送带有超时的Post https请求
*/
func postRequestTLS(url, contentType string, body io.Reader, tlsConfig *tls.Config) ([]byte, error) {
	client := timeoutClientWithTLS(tlsConfig)
	resp, err := client.Post(url, contentType, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respbody, nil
}

const (
	maxTimeout       int64         = 25
	connectTimeout   time.Duration = 3 * time.Second
	readWriteTimeout time.Duration = 5 * time.Second
)

/**
  网络请求链接定义
*/
func timeoutClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial:                timeoutDialer(connectTimeout, readWriteTimeout),
			MaxIdleConnsPerHost: 200,
			DisableKeepAlives:   true,
		},
	}
}

/**
  网络请求链接定义
*/
func timeoutClientWithTLS(tlsConfig *tls.Config) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:     tlsConfig,
			Dial:                timeoutDialer(connectTimeout, readWriteTimeout),
			MaxIdleConnsPerHost: 200,
			DisableKeepAlives:   true,
		},
	}
}
func timeoutDialer(cTimeout time.Duration,
	rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

func newError(msg string) error {
	return errors.New(msg)
}

func newErrorByE(e Error) error {
	if e.Code == Succ {
		return nil
	}
	return newError(e.Msg)
}
