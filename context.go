package ppp

import (
	"crypto/rsa"
	"crypto/tls"
	"github.com/hprose/hprose-golang/rpc"
	"time"
)

type Context struct {
	t    int64 // 请求开始时间
	auth *Auth // 权限信息
	uid  string
	// Tag 请求收款app的tag
	Tag string
	// Type
	Type string
	cfg  *config

	userData map[string]interface{}
}

var _ rpc.Context = &Context{}

func NewContextWithCfg(ty, tg string) *Context {
	ctx := &Context{Type: ty, Tag: tg}
	ctx.SetCfg(ty, tg)
	return ctx
}
func (c *Context) SetCfg(ty, tg string) {
	c.cfg = &config{}
	switch ty {
	case ALIPAY:
		if tg == "" {
			c.cfg = &alipay.def
		} else {
			if cfg, ok := alipay.cfgs[tg]; ok {
				c.cfg = &cfg
			}
		}
	case WXPAY:
		if tg == "" {
			c.cfg = &wxpay.ws.def
		} else {
			if cfg, ok := wxpay.ws.cfgs[tg]; ok {
				c.cfg = &cfg
			}
		}
	case WXPAYSINGLE:
		if tg == "" {
			c.cfg = &wxpaySingle.def
		} else {
			if cfg, ok := wxpaySingle.cfgs[tg]; ok {
				c.cfg = &cfg
			}
		}
	}
}
func (c *Context) tlsConfig() *tls.Config {
	return c.cfg.tlsConfig
}

func (c *Context) secret() string {
	return c.cfg.secret
}
func (c *Context) token() string {
	return c.auth.Token
}

func (c *Context) appid() string {
	return c.cfg.appid
}
func (c *Context) subappid() string {
	if c.auth != nil {
		return c.auth.SubAppID
	}
	return ""
}
func (c *Context) privateKey() *rsa.PrivateKey {
	return c.cfg.private
}

func (c *Context) Notify() string {
	if c.cfg != nil {
		return c.cfg.notify
	}
	return ""
}
func (c *Context) serviceid() string {
	if c.cfg != nil {
		return c.cfg.serviceid
	}
	return ""
}
func (c *Context) url() string {
	if c.cfg != nil {
		return c.cfg.url
	}
	return ""
}
func (c *Context) gt() int64 {
	if c.t == 0 {
		c.t = time.Now().Unix()
	}
	return c.t
}
func (c *Context) userid() string {
	return c.uid
}
func (c *Context) mchid() string {
	if c.auth != nil {
		return c.auth.MchID
	}
	return ""
}
func (c *Context) getAuth(userid, mchid string) *Auth {
	if c.auth != nil {
		return c.auth
	}
	if userid == "" {
		switch c.Type {
		case ALIPAY:
			c.auth = &Auth{
				MchID:  c.cfg.serviceid,
				Status: AuthStatusSucc,
			}
		case WXPAY:
			c.auth = &Auth{
				MchID:  mchid,
				Status: AuthStatusSucc,
			}
		}
	} else {
		c.uid = userid
		c.auth = token(userid, mchid, c.Type, c.appid())
	}
	return c.auth
}

// UserData return the user data
func (context *Context) UserData() map[string]interface{} {
	return context.userData
}

// GetInt from hprose context
func (context *Context) GetInt(
	key string, defaultValue ...int) int {
	if value, ok := context.userData[key]; ok {
		if value, ok := value.(int); ok {
			return value
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// GetUInt from hprose context
func (context *Context) GetUInt(
	key string, defaultValue ...uint) uint {
	if value, ok := context.userData[key]; ok {
		if value, ok := value.(uint); ok {
			return value
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// GetInt64 from hprose context
func (context *Context) GetInt64(
	key string, defaultValue ...int64) int64 {
	if value, ok := context.userData[key]; ok {
		if value, ok := value.(int64); ok {
			return value
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// GetUInt64 from hprose context
func (context *Context) GetUInt64(
	key string, defaultValue ...uint64) uint64 {
	if value, ok := context.userData[key]; ok {
		if value, ok := value.(uint64); ok {
			return value
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// GetFloat from hprose context
func (context *Context) GetFloat(
	key string, defaultValue ...float64) float64 {
	if value, ok := context.userData[key]; ok {
		if value, ok := value.(float64); ok {
			return value
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// GetBool from hprose context
func (context *Context) GetBool(
	key string, defaultValue ...bool) bool {
	if value, ok := context.userData[key]; ok {
		if value, ok := value.(bool); ok {
			return value
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return false
}

// GetString from hprose context
func (context *Context) GetString(
	key string, defaultValue ...string) string {
	if value, ok := context.userData[key]; ok {
		if value, ok := value.(string); ok {
			return value
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// GetInterface from hprose context
func (context *Context) GetInterface(
	key string, defaultValue ...interface{}) interface{} {
	if value, ok := context.userData[key]; ok {
		return value
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return nil
}

// Get value from hprose context
func (context *Context) Get(key string) interface{} {
	if value, ok := context.userData[key]; ok {
		return value
	}
	return nil
}

// SetInt to hprose context
func (context *Context) SetInt(key string, value int) {
	context.userData[key] = value
}

// SetUInt to hprose context
func (context *Context) SetUInt(key string, value uint) {
	context.userData[key] = value
}

// SetInt64 to hprose context
func (context *Context) SetInt64(key string, value int64) {
	context.userData[key] = value
}

// SetUInt64 to hprose context
func (context *Context) SetUInt64(key string, value uint64) {
	context.userData[key] = value
}

// SetFloat to hprose context
func (context *Context) SetFloat(key string, value float64) {
	context.userData[key] = value
}

// SetBool to hprose context
func (context *Context) SetBool(key string, value bool) {
	context.userData[key] = value
}

// SetString to hprose context
func (context *Context) SetString(key string, value string) {
	context.userData[key] = value
}

// SetInterface to hprose context
func (context *Context) SetInterface(key string, value interface{}) {
	context.userData[key] = value
}

// Set is an alias of SetInterface
func (context *Context) Set(key string, value interface{}) {
	context.userData[key] = value
}
