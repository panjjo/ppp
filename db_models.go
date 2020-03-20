package ppp

import (
	proto "github.com/panjjo/ppp/proto"
)

const (
	tradeTable = "trades"
)

func getTrade(q interface{}) *proto.Trade {
	session := DBPool.Get()
	defer session.Close()
	trade := &proto.Trade{}
	res := session.FindOne(tradeTable, q, trade)
	if res != nil {
		trade = res.(*proto.Trade)
	}
	return trade
}

func saveTrade(trade *proto.Trade) error {
	session := DBPool.Get()
	defer session.Close()
	return session.Save(tradeTable, trade)
}

func updateTrade(query, update interface{}) error {
	session := DBPool.Get()
	defer session.Close()
	return session.Update(tradeTable, query, update)
}
func upsertTrade(query, update interface{}) (interface{}, error) {
	session := DBPool.Get()
	defer session.Close()
	return session.UpSert(tradeTable, query, update)
}

const (
	refundTable = "refunds"
)

func saveRefund(refund *proto.Refund) error {
	session := DBPool.Get()
	defer session.Close()
	return session.Save(refundTable, refund)
}


const (
	userTable = "users"
)


// 查询用户
func getUser(userid, t string) *proto.User {
	session := DBPool.Get()
	defer session.Close()
	user := &proto.User{}
	res := session.FindOne(userTable, map[string]interface{}{"userid": userid, "from": t}, user)
	if res != nil {
		user = res.(*proto.User)
	}
	return user
}

// 更新用户
func updateUser(query, update interface{}) error {
	session := DBPool.Get()
	defer session.Close()
	return session.Update(userTable, query, update)
}

//
func updateUserMulti(query, update interface{}) error {
	session := DBPool.Get()
	defer session.Close()
	_, err := session.UpAll(userTable, query, update)
	return err
}

// 保存用户
func saveUser(user *proto.User) error {
	session := DBPool.Get()
	defer session.Close()
	return session.Save(userTable, user)
}


const (
	accountTable = "accounts"

)

// 获取授权
func getToken(mchid, appid string) *proto.Account {
	session := DBPool.Get()
	defer session.Close()
	auth := &proto.Account{}
	res := session.FindOne(accountTable, map[string]interface{}{"mchid": mchid, "appid": appid}, auth)
	if res != nil {
		auth = res.(*proto.Account)
	}
	return auth
}

// 刷新授权
func updateToken(mchid, appid string, update interface{}) error {
	session := DBPool.Get()
	defer session.Close()
	return session.Update(accountTable, map[string]interface{}{"mchid": mchid, "appid": appid}, update)
}

// 保存授权
func saveToken(auth *proto.Account) error {
	session := DBPool.Get()
	defer session.Close()
	return session.Save(accountTable, auth)
}

// 通过userid 或者 mchid 获取授权信息
func token(userid, mchid, t, appid string) *proto.Account {
	auth := &proto.Account{}
	if mchid == "" {
		user := getUser(userid, t)
		if user.Status != proto.Accountstatus_ok {
			return auth
		}
		mchid = user.MchID
	}
	return getToken(mchid, appid)
}