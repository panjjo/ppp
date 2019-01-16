package ppp

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
	session := DBPool.Get()
	defer session.Close()
	auth := &Auth{}
	res := session.FindOne(authTable, map[string]interface{}{"mchid": mchid, "appid": appid}, auth)
	if res != nil {
		auth = res.(*Auth)
	}
	return auth
}

// 刷新授权
func updateToken(mchid, appid string, update interface{}) error {
	session := DBPool.Get()
	defer session.Close()
	return session.Update(authTable, map[string]interface{}{"mchid": mchid, "appid": appid}, update)
}

// 保存授权
func saveToken(auth *Auth) error {
	session := DBPool.Get()
	defer session.Close()
	return session.Save(authTable, auth)
}

// 通过userid 或者 mchid 获取授权信息
func token(userid, mchid,t, appid string) *Auth {
	auth := &Auth{}
	if mchid == "" {
		user := getUser(userid,t)
		if user.Status != UserSucc {
			return auth
		}
		mchid = user.MchID
	}
	return getToken(mchid, appid)
}
