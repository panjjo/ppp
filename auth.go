package ppp

import (
	"github.com/panjjo/ppp/db"
)

const (
	authTable = "auths"

	// AuthStatusSucc 权限授权成功
	AuthStatusSucc Status = 1
	// AuthStatusWaitSigned 等待签约
	AuthStatusWaitSigned Status = 0
)

// Auth 授权使用
type Auth struct {
	ID     string
	Token  string
	Status Status
	// MchID 商户号
	MchID string
	// Form 来源 alipay wxpay
	From    string
	Account string
	// AppID 授权对应的appid
	AppID string
	// SubAppID 微信子商户appid
	SubAppID string
}

// 获取授权
func getToken(mchid, appid string) *Auth {
	auth := &Auth{}
	DBClient.Get(authTable, db.M{"mchid": mchid, "appid": appid}, auth)
	return auth
}

// 刷新授权
func updateToken(mchid, appid string, update interface{}) error {
	return DBClient.Update(authTable, db.M{"mchid": mchid, "appid": appid}, db.M{"$set": update})
}

// 保存授权
func saveToken(auth *Auth) error {
	return DBClient.Insert(authTable, auth)
}

// 通过userid 或者 mchid 获取授权信息
func token(userid, mchid, t, appid string) *Auth {
	auth := &Auth{}
	if mchid == "" {
		user := getUser(userid, t)
		if user.Status != UserSucc {
			return auth
		}
		mchid = user.MchID
	}
	return getToken(mchid, appid)
}

// AccessToken 个人授权返回token
type AccessToken struct {
	UserID      string
	AccessToken string
	ExpiresIN   string
	ReToken     string
	ReExpiresIn string
}
