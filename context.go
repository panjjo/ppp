package ppp

import (
	"crypto/rsa"
	"crypto/tls"
	"github.com/sirupsen/logrus"
	proto "github.com/panjjo/ppp/proto"
	"time"
)

type Context struct {
	t    int64 // 请求开始时间
	auth *proto.Account// 权限信息
	uid  string
	// Tag 请求收款app的tag
	Tag string
	// Type
	Type string
	cfg  *config

	userData map[string]interface{}
}

func NewContextWithCfg(ty, tg string) *Context {
	ctx := &Context{Type: ty, Tag: tg}
	ctx.SetCfg(ty, tg)
	return ctx
}
func (c *Context) SetCfg(ty, tg string) {
	logrus.Debugln("req set cfg ,ty:", ty, "tg:", tg)
	c.cfg = &config{}
	switch ty {
	case proto.ALIPAY:
		if tg == "" {
			c.cfg = &alipay.def
		} else {
			if cfg, ok := alipay.cfgs[tg]; ok {
				c.cfg = &cfg
			}
		}
	case proto.WXPAY:
		if tg == "" {
			c.cfg = &wxpay.ws.def
		} else {
			if cfg, ok := wxpay.ws.cfgs[tg]; ok {
				c.cfg = &cfg
			}
		}
	case proto.WXPAYSINGLE:
		if tg == "" {
			c.cfg = &wxpaySingle.def
		} else {
			if cfg, ok := wxpaySingle.cfgs[tg]; ok {
				c.cfg = &cfg
			}
		}
	}
	logrus.Debugf("end cfg:%+v", c.cfg)
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
func (c *Context) getAuth(userid, mchid string) *proto.Account {
	logrus.Debugln("get auth, userid:", userid, "mchid:", mchid)
	if c.auth != nil {
		return c.auth
	}
	if userid == "" {
		switch c.Type {
		case proto.ALIPAY:
			c.auth = &proto.Account{
				MchID:  c.cfg.serviceid,
				Status: proto.Accountstatus_ok,
			}
		case proto.WXPAY:
			c.auth = &proto.Account{
				MchID:  mchid,
				Status: proto.Accountstatus_ok,
			}
		}
	} else {
		c.uid = userid
		c.auth = token(userid, mchid, c.Type, c.appid())
	}
	logrus.Debugf("end auth: %+v", c.auth)
	return c.auth
}

