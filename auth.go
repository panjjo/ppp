package ppp

const (
	authTable = "auths"
)

// Auth 授权使用
type Auth struct {
	ID      string
	Token   string `json:"-"`
	Status  Status
	MchID   string
	From    string
	Account string
	AppID   string //微信子商户appid
}

//获取授权
func getToken(mchid, t string) *Auth {
	session := DBPool.Get()
	defer session.Close()
	auth := &Auth{}
	res := session.FindOne(authTable, map[string]interface{}{"mchid": mchid, "from": t}, auth)
	if res != nil {
		auth = res.(*Auth)
	}
	return auth
}

//刷新授权
func updateToken(mchid, t string, update interface{}) error {
	session := DBPool.Get()
	defer session.Close()
	return session.Update(authTable, map[string]interface{}{"mchid": mchid, "from": t}, update)
}

//保存授权
func saveToken(auth *Auth) error {
	session := DBPool.Get()
	defer session.Close()
	return session.Save(authTable, auth)
}

// 通过userid 或者 mchid 获取授权信息
func token(userid, mchid, t string) *Auth {
	auth := &Auth{}
	if mchid == "" {
		user := getUser(userid, t)
		if user.Status != UserSucc {
			return auth
		}
		mchid = user.MchID
	}
	return getToken(mchid, t)
}
