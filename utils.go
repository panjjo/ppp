package ppp

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

/**
  获取当前时间戳
*/
func getNowSec() int64 {
	return time.Now().Unix()
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

var base64Base *base64.Encoding = base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")

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
  压json
*/
func jsonEncode(ob interface{}) []byte {
	if b, err := json.Marshal(ob); err == nil {
		return b
	} else {
		return []byte("")
	}
}

/**
  解json
*/
func jsonDecode(data []byte, ob interface{}) error {
	return json.Unmarshal(data, ob)
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
  map排序并根据排序结果kv拼接
*/
func mapSortAndJoin(m map[string]string, step1, step2 string) string {
	ms := make(mapSorter, 0, len(m))
	for k, v := range m {
		ms = append(ms, sortItem{k, v})
	}
	sort.Sort(ms)
	s := []string{}
	for _, p := range ms {
		s = append(s, p.Key+step1+p.Val.(string))
	}
	return strings.Join(s, step2)
}

var (
	maxTimeout       int           = 25
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

/*
	发送带有超时的Get请求
*/
func getRequest(url string) ([]byte, error) {
	client := timeoutClient()
	resp, err := client.Get(url)
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

func newError(msg string) error {
	return errors.New(msg)
}
